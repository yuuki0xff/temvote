package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	ServerErrorMsg = "500 Internal Server Error"
)

type RouterOption struct {
	StaticDir    string `envconfig:"STATIC_DIR"`
	DeployDir    string `envconfig:"DEPLOY_DIR"`
	SecretFile   string `envconfig:"SECRET_FILE"`
	MetricsFile  string `envconfig:"METRICS_FILE"`
	CookieSecret string `envconfig:"COOKIE_SECRET"`
	DBDriver     string `envconfig:"DB_DRIVER"`
	DBUrl        string `envconfig:"DB_URL"`
	ThingWorxURL string `envconfig:"THINGWORX_URL"`
}

type StatusAPIResponse struct {
	Status *RoomStatus `json:"status"`
	MyVote *MyVote     `json:"myvote"`
}

func getRouter(opt RouterOption, db *sql.DB, ctx context.Context) *mux.Router {
	staticHandler := http.FileServer(http.Dir(opt.StaticDir))
	deployHandler := http.FileServer(http.Dir(opt.DeployDir))
	tmpl, err := template.ParseGlob("template/*.html")
	if err != nil {
		panic(err)
	}

	thingworx := &ThingWorxClient{
		URL: opt.ThingWorxURL,
	}

	rsm := NewRoomStatusManager(db, thingworx, ctx)
	sm, err := NewSecretManager(opt.SecretFile)
	if err != nil {
		panic(err)
	}
	mw, err := NewMetricsWriter(opt.MetricsFile)
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		var err error
		var res StatusAPIResponse

		w.Header().Set("Cache-Control", "no-store")

		tx, err := rsm.GetTx(w, req)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		strRoomID := req.URL.Query().Get("room")
		roomID, err := StringToRoomID(strRoomID)
		if err != nil {
			log.Printf("WARN: can not parse RoomID(%s): %s\n", strRoomID, err.Error())
			http.Error(w, "room parameter is invalid", http.StatusBadRequest)
			return
		}
		res.Status, err = tx.GetStatus(roomID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		res.MyVote, err = tx.GetMyVote(roomID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(res)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(200)
		w.Write(js)
	}).Methods("GET")
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		var err error
		var res StatusAPIResponse

		w.Header().Set("Cache-Control", "no-store")

		tx, err := rsm.GetTx(w, req)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		strRoomID := req.URL.Query().Get("room")
		roomID, err := StringToRoomID(strRoomID)
		if err != nil {
			log.Printf("WARN: can not parse RoomID(%s): %s\n", strRoomID, err.Error())
			http.Error(w, "room parameter is invalid", http.StatusBadRequest)
			return
		}

		choice := VoteChoice(req.FormValue("vote"))
		switch choice {
		case Hot:
		case Comfort:
		case Cold:
		default:
			log.Printf("WARN: vote parameter is invalid: vote=%d\n", choice)
			http.Error(w, "vote parameter is invalid", http.StatusBadRequest)
			return
		}
		err = tx.Vote(roomID, choice)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		res.Status, err = tx.GetStatus(roomID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		res.MyVote, err = tx.GetMyVote(roomID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(res)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		tx.Commit()
		tx.s.ExtendExpiration()
		w.WriteHeader(200)
		w.Write(js)
	}).Methods("POST")

	router.HandleFunc("/metrics", func(w http.ResponseWriter, req *http.Request) {
		hostid := req.Header.Get("X-HOSTID")
		secret := req.Header.Get("X-SECRET")
		if !sm.hasAuth(hostid, secret) {
			w.WriteHeader(400)
			return
		}

		tag := req.Header.Get("X-TAG")
		rawBody, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("ERROR:", err)
			w.WriteHeader(500)
			println(err.Error(), req.ContentLength)
			return
		}

		if err := mw.Write(Metrics{
			HostID:    hostid,
			Tag:       tag,
			Body:      string(rawBody),
			Timestamp: time.Now().Unix(),
		}); err != nil {
			log.Println("ERROR:", err)
			w.WriteHeader(500)
			println(err.Error())
			return
		}
	}).Methods("POST")

	deployFileServer := http.StripPrefix("/deploy/", deployHandler)
	router.HandleFunc("/deploy/{name:.*}", func(w http.ResponseWriter, req *http.Request) {
		hostid := req.Header.Get("X-HOSTID")
		secret := req.Header.Get("X-SECRET")
		if !sm.hasAuth(hostid, secret) {
			w.WriteHeader(400)
			return
		}

		deployFileServer.ServeHTTP(w, req)
	}).Methods("GET")
	router.Handle("/", http.RedirectHandler("/select_room.html", 303)).Methods("GET")
	router.HandleFunc("/vote/{roomid}", func(w http.ResponseWriter, req *http.Request) {
		tx, err := rsm.GetTx(w, req)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		vars := mux.Vars(req)
		strRoomID := vars["roomid"]
		roomID, err := StringToRoomID(strRoomID)
		if err != nil {
			log.Printf("WARN: can not parse RoomID(%s): %s\n", strRoomID, err.Error())
			http.Error(w, "roomid parameter is invalid", http.StatusBadRequest)
			return
		}

		roomName, err := tx.GetRoomName(roomID)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "vote.html", &struct {
			RoomID   RoomID
			RoomName string
		}{
			RoomID:   roomID,
			RoomName: roomName,
		})

	}).Methods("GET")
	router.HandleFunc("/select_room.html", func(w http.ResponseWriter, req *http.Request) {
		tx, err := rsm.GetTx(w, req)
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		names, groups, err := tx.GetAllRoomsInfo()
		if err != nil {
			log.Println("ERROR:", err)
			http.Error(w, ServerErrorMsg, http.StatusInternalServerError)
			return
		}

		tmpl.ExecuteTemplate(w, "select_room.html", &struct {
			RoomNames  RoomNameMap
			RoomGroups RoomGroupMap
		}{
			RoomNames:  names,
			RoomGroups: groups,
		})
	}).Methods("GET")
	router.Handle("/{name:.*}", staticHandler).Methods("GET")

	return router
}

func startHttpServer(ctx context.Context, router *mux.Router) (err error) {
	srv := http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: router,
	}
	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()
	log.Println("start server")
	srv.ListenAndServe()
	return
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		log.Println("signal handled")
		cancel()
	}()

	var opt RouterOption
	envconfig.Process("TEMVOTE", &opt)

	db, err := sql.Open(opt.DBDriver, opt.DBUrl)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	router := getRouter(opt, db, ctx)
	if err := startHttpServer(ctx, router); err != nil {
		panic(err)
	}
}
