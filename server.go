package main

import (
	"os"
	"os/signal"
	"context"
	"syscall"
	"net/http"
	"github.com/gorilla/mux"
	"encoding/json"
	"fmt"
	"sync"
)

var roomIds = []string{
	"kougi201",
	"kougi202",
	"kougi203",
	"kougi204",

	"kougi301",
	"kougi302",
	"kougi303",
	"kougi304",
}

type RoomStatus struct {
	RoomID     string  `json:"id"`
	Templature float32 `json:"templature"`
	Hot        uint    `json:"hot"`
	Cold       uint    `json:"cold"`
	lock       sync.RWMutex
}

func getRouter() *mux.Router {
	cwd, _ := os.Getwd()
	docroot := http.Dir(cwd + "/static")
	statMap := make(map[string]*RoomStatus)
	for _, id := range roomIds {
		statMap[id] = &RoomStatus{
			RoomID: id,
			Templature: 30.0,
			Hot: 0,
			Cold: 0,
			lock: sync.RWMutex{},
		}
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

		roomId := req.URL.Query().Get("room")
		stat := statMap[roomId]
		if stat == nil {
			w.WriteHeader(500)
			return
		}
		stat.lock.RLock()
		defer stat.lock.RUnlock()

		js, err := json.Marshal(*stat)
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
		stat := statMap[roomId]
		if stat == nil {
			w.WriteHeader(500)
			return
		}
		stat.lock.Lock()
		defer stat.lock.Unlock()

		switch req.FormValue("vote"){
		case "hot":
			stat.Hot++
		case "cold":
			stat.Cold++
		default:
			w.WriteHeader(400)
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
	router.Handle(`/`, http.FileServer(docroot)).Methods("GET")
	router.Handle(`/{name:.*}`, http.FileServer(docroot)).Methods("GET")

	return router
}

func startHttpServer(ctx context.Context, router *mux.Router) (err error) {
	srv := http.Server{
		Addr: "0.0.0.0:8080",
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

	router := getRouter()
	if err := startHttpServer(ctx, router); err != nil {
		panic(err)
	}
}