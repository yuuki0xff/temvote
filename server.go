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

type Status struct {
	Templature float32 `json:"templature"`
	Hot        uint    `json:"hot"`
	Cold       uint    `json:"cold"`
	lock       sync.RWMutex
}

func getRouter() *mux.Router {
	cwd, _ := os.Getwd()
	docroot := http.Dir(cwd + "/static")
	stat := &Status{
		Templature: 30.0,
		Hot: 20,
		Cold: 1,
	}

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/status", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Cache-Control", "no-store")

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