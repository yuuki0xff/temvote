package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
	"sync"
	"time"
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

const (
	SESSION_NAME = "temvote_myvote"
)

type SessionFunc func(func(r *http.Request, w *http.ResponseWriter, s *sessions.CookieStore))

type RoomStatus struct {
	RoomID     string  `json:"id"`
	Templature float32 `json:"templature"`
	Hot        uint    `json:"hot"`
	Cold       uint    `json:"cold"`
	lock       sync.RWMutex
}

type MyVote struct {
	Vote      string `json:"vote"`
	Timestamp int64  `json:"timestamp"`
}

type RoomStatusManager struct {
	statMap map[string]*RoomStatus
}

func NewRoomStatusManager() *RoomStatusManager {
	rs := &RoomStatusManager{}
	rs.statMap = make(map[string]*RoomStatus)
	for _, id := range roomIds {
		rs.statMap[id] = &RoomStatus{
			RoomID:     id,
			Templature: 30.0,
			Hot:        0,
			Cold:       0,
			lock:       sync.RWMutex{},
		}
	}
	return rs
}

func getSessionName(id string) string {
	return SESSION_NAME + "/" + id
}

func (rs *RoomStatusManager) GetMyVote(sf SessionFunc, id string) (*MyVote, error) {
	var err error
	var vote *MyVote

	sf(func(r *http.Request, w *http.ResponseWriter, store *sessions.CookieStore) {
		s, err := store.Get(r, getSessionName(id))
		if err != nil {
			return
		}

		vote = &MyVote{
			Vote:      s.Values["vote"].(string),
			Timestamp: s.Values["timestamp"].(int64),
		}
	})
	return vote, err
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

func (rs *RoomStatusManager) Vote(sf SessionFunc, id string, hot, cold int) error {
	if !(hot == 1 || cold == 1) {
		return errors.New("Invalid args")
	}

	sf(func(r *http.Request, w *http.ResponseWriter, store *sessions.CookieStore) {
		s, err := store.Get(r, getSessionName(id))
		if err != nil {
			s = sessions.NewSession(store, getSessionName(id))
		}

		// 以前の投票を取り消す
		if s.Values["vote"] != nil {
			switch s.Values["vote"].(string) {
			case "hot":
				hot -= 1
			case "cold":
				cold -= 1
			}
		}

		// 投票結果をCookieに保存
		if hot != 0 {
			s.Values["vote"] = "hot"
		} else if cold != 0 {
			s.Values["vote"] = "cold"
		}
		s.Values["timestamp"] = time.Now().Unix()

		s.Save(r, *w)
	})

	return rs.setter(id, func(status *RoomStatus) error {
		status.Hot = uint(int(status.Hot) + hot)
		status.Cold = uint(int(status.Cold) + cold)
		return nil
	})
}
