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
	"errors"
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

type RoomStatusManager struct {
	statMap map[string]*RoomStatus
}

func NewRoomStatusManager() (*RoomStatusManager) {
	rs := &RoomStatusManager{}
	rs.statMap = make(map[string]*RoomStatus)
	for _, id := range roomIds {
		rs.statMap[id] = &RoomStatus{
			RoomID: id,
			Templature: 30.0,
			Hot: 0,
			Cold: 0,
			lock: sync.RWMutex{},
		}
	}
	return rs
}

func (rs *RoomStatusManager) GetStatus(id string) (*RoomStatus, error) {
	stat := rs.statMap[id]
	if stat == nil {
		return nil, errors.New(fmt.Sprintf(`invalid id: "%s"`, id))
	}

	var newStat = *stat
	return &newStat, nil
}

func (rs *RoomStatusManager) setter(id string, callback func(*RoomStatus) error) error {
	stat := rs.statMap[id]
	if stat == nil {
		return nil
	}
	stat.lock.Lock()
	defer stat.lock.Unlock()
	return callback(stat)
}

func (rs *RoomStatusManager) Vote(id string, hot, cold int) error {
	return rs.setter(id, func(status *RoomStatus) error {
		status.Hot = uint(int(status.Hot) + hot)
		status.Cold = uint(int(status.Cold) + cold)
		return nil
	})
}

func getRouter() *mux.Router {
	cwd, _ := os.Getwd()
	docroot := http.Dir(cwd + "/static")
	rsm := NewRoomStatusManager()

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

		switch req.FormValue("vote"){
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