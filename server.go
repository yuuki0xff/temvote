package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RouterOption struct {
	StaticDir    string `envconfig:"STATIC_DIR"`
	DeployDir    string `envconfig:"DEPLOY_DIR"`
	SecretFile   string `envconfig:"SECRET_FILE"`
	MetricsFile  string `envconfig:"METRICS_FILE"`
	CookieSecret string `envconfig:"COOKIE_SECRET"`
	DBFile       string `envconfig:"DB_FILE"`
}

type StatusAPIResponse struct {
	Status *RoomStatus `json:"status"`
	MyVote *MyVote     `json:"myvote"`
}

func getRouter(opt RouterOption, db *bolt.DB) *mux.Router {
	staticHandler := http.FileServer(http.Dir(opt.StaticDir))
	deployHandler := http.FileServer(http.Dir(opt.DeployDir))

	store := sessions.NewCookieStore([]byte(opt.CookieSecret))

	rsm := NewRoomStatusManager(db)
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
		w.Header().Set("Cache-Control", "no-store")
		res := StatusAPIResponse{}
		sf := func(callback func(r *http.Request, w *http.ResponseWriter, s *sessions.CookieStore)) {
			callback(req, &w, store)
		}

		roomId := req.URL.Query().Get("room")
		res.Status, err = rsm.GetStatus(roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		res.MyVote, err = rsm.GetMyVote(sf, roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		js, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write(js)
	}).Methods("GET")
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		var err error
		w.Header().Set("Cache-Control", "no-store")
		res := StatusAPIResponse{}
		sf := func(callback func(r *http.Request, w *http.ResponseWriter, s *sessions.CookieStore)) {
			callback(req, &w, store)
		}

		roomId := req.URL.Query().Get("room")

		switch req.FormValue("vote") {
		case "hot":
			err = rsm.Vote(sf, roomId, 1, 0)
		case "cold":
			err = rsm.Vote(sf, roomId, 0, 1)
		default:
			w.WriteHeader(400)
			return
		}
		if err != nil {
			w.WriteHeader(500)
			return
		}

		res.Status, err = rsm.GetStatus(roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		res.MyVote, err = rsm.GetMyVote(sf, roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		js, err := json.Marshal(res)
		if err != nil {
			w.WriteHeader(500)
			return
		}

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
	router.Handle(`/`, staticHandler).Methods("GET")
	router.Handle(`/{name:.*}`, staticHandler).Methods("GET")

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
	fmt.Println("start server")
	srv.ListenAndServe()
	return
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sig
		fmt.Println("signal handled")
		cancel()
	}()

	var opt RouterOption
	envconfig.Process("TEMVOTE", &opt)

	db, err := bolt.Open(opt.DBFile, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	router := getRouter(opt, db)
	if err := startHttpServer(ctx, router); err != nil {
		panic(err)
	}
}
