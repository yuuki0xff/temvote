package main

import (
	"context"
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	ServerErrorMsg = "500 Internal Server Error"
)

type RouterOption struct {
	StaticDir       string `envconfig:"STATIC_DIR"`
	DBDriver        string `envconfig:"DB_DRIVER"`
	DBUrl           string `envconfig:"DB_URL"`
	ThingWorxURL    string `envconfig:"THINGWORX_URL"`
	ThingWorxAppKey string `envconfig:"THINGWORX_APP_KEY"`
}

type StatusAPIResponse struct {
	Status *RoomStatus `json:"status"`
	MyVote *MyVote     `json:"myvote"`
}

func getRouter(opt RouterOption, db *sql.DB, ctx context.Context) *mux.Router {
	staticHandler := http.FileServer(http.Dir(opt.StaticDir))
	tmpl, err := template.ParseGlob("template/*.html")
	if err != nil {
		panic(err)
	}

	thingworx := &ThingWorxClient{
		URL:    opt.ThingWorxURL,
		AppKey: opt.ThingWorxAppKey,
	}

	rsm := NewRoomStatusManager(db, thingworx, ctx)

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		var err error
		var res StatusAPIResponse

		w.Header().Set("Cache-Control", "no-store")

		tx, err := rsm.GetTx(w, req, false)
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

		tx, err := rsm.GetTx(w, req, true)
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

		tx.s.ExtendExpiration()
		tx.Commit()
		w.WriteHeader(200)
		w.Write(js)
	}).Methods("POST")

	router.Handle("/", http.RedirectHandler("/select_room.html", 303)).Methods("GET")
	router.HandleFunc("/vote/{roomid}", func(w http.ResponseWriter, req *http.Request) {
		tx, err := rsm.GetTx(w, req, false)
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
		tx, err := rsm.GetTx(w, req, false)
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
	// set up logger
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

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
