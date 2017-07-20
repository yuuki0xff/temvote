package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type RouterOption struct {
	StaticDir   string `envconfig:"STATIC_DIR"`
	DeployDir   string `envconfig:"DEPLOY_DIR"`
	SecretFile  string `envconfig:"SECRET_FILE"`
	MetricsFile string `envconfig:"METRICS_FILE"`
}

func getRouter(opt RouterOption) *mux.Router {
	staticHandler := http.FileServer(http.Dir(opt.StaticDir))
	deployHandler := http.FileServer(http.Dir(opt.DeployDir))

	rsm := NewRoomStatusManager()
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
		w.Header().Set("Cache-Control", "no-store")

		roomId := req.URL.Query().Get("room")
		stat, err := rsm.GetStatus(roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}
		js, err := json.Marshal(stat)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		w.Write(js)
	}).Methods("GET")
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		roomId := req.URL.Query().Get("room")

		switch req.FormValue("vote") {
		case "hot":
			rsm.Vote(roomId, 1, 0)
		case "cold":
			rsm.Vote(roomId, 0, 1)
		default:
			w.WriteHeader(400)
			return
		}

		stat, err := rsm.GetStatus(roomId)
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.WriteHeader(200)
		js, err := json.Marshal(*stat)
		if err != nil {
			w.WriteHeader(500)
			return
		}
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
	router := getRouter(opt)
	if err := startHttpServer(ctx, router); err != nil {
		panic(err)
	}
}
