package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gorilla/sessions"
	"net/http"
	"sync"
	"time"
)

var roomIds = []string{
	"room1",
	"room2",

	"kougi201",
	"kougi202",
	"kougi203",
	"kougi204",

	"kougi301",
	"kougi302",
	"kougi303",
	"kougi304",
}

var RoomNames = map[string]string{
	"room1": "研究室1",
	"room2": "研究室2",

	"kougi201": "講義棟201",
	"kougi202": "講義棟202",
	"kougi203": "講義棟203",
	"kougi204": "講義棟204",

	"kougi301": "講義棟301",
	"kougi302": "講義棟302",
	"kougi303": "講義棟303",
	"kougi304": "講義棟304",
}

var RoomGroups = map[string]map[string][]string{
	"片研": {
		"11階": {
			"room1",
			"room2",
		},
	},
	"講義棟": {
		"2階": {
			"kougi201",
			"kougi202",
			"kougi203",
			"kougi204",
		},
		"3階": {
			"kougi301",
			"kougi302",
			"kougi303",
			"kougi304",
		},
	},
}

const (
	SESSION_NAME = "temvote_myvote"
	INTERVAL     = 1 * time.Minute
)

var (
	BUCKET_NAME = []byte("room_status")
)

type SessionFunc func(func(r *http.Request, w *http.ResponseWriter, s *sessions.CookieStore))

type RoomStatus struct {
	RoomID      string  `json:"id"`
	Temperature float32 `json:"temperature"`
	Humidity    float32 `json:"humidity"`
	Hot         uint    `json:"hot"`
	Comfort     uint    `json:"comfort"`
	Cold        uint    `json:"cold"`
	IsConnected bool    `json:"isConnected"`
	lock        sync.RWMutex
}

type MyVote struct {
	Vote      string `json:"vote"`
	Timestamp int64  `json:"timestamp"`
}

type RoomStatusManager struct {
	db        *bolt.DB
	thingworx *ThingWorxClient
	statMap   map[string]*RoomStatus
}

func NewRoomStatusManager(db *bolt.DB, thingworx *ThingWorxClient, ctx context.Context) *RoomStatusManager {
	// initialize db
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(BUCKET_NAME)
		return err
	}); err != nil {
		panic(err)
	}

	// create RSM
	rs := &RoomStatusManager{}
	rs.db = db
	rs.thingworx = thingworx
	rs.statMap = make(map[string]*RoomStatus)
	if err := rs.tx(func(bucket *bolt.Bucket) {
		for _, id := range roomIds {
			js := bucket.Get([]byte(id))
			if len(js) == 0 {
				// 新しい教室なら、デフォルト値を格納しておく
				rs.statMap[id] = &RoomStatus{
					RoomID:      id,
					Temperature: -1,
					Hot:         0,
					Cold:        0,
					IsConnected: false,
					lock:        sync.RWMutex{},
				}
				continue
			}

			stat := &RoomStatus{
				lock: sync.RWMutex{},
			}
			if err := json.Unmarshal(js, stat); err != nil {
				panic(err)
			}
			rs.statMap[id] = stat
		}
	}); err != nil {
		panic(err)
	}

	go rs.updateStatusWorker(ctx)
	return rs
}

func getSessionName(id string) string {
	return SESSION_NAME + "___" + id
}

// 読み取り専用のトランザクションを開始する
func (rs *RoomStatusManager) tx(callback func(*bolt.Bucket)) error {
	tx, err := rs.db.Begin(false)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(BUCKET_NAME)
	if bucket == nil {
		return errors.New(fmt.Sprintf("Bucket is not exists: %s", BUCKET_NAME))
	}
	callback(bucket)
	return nil
}

// 書き込み可能なトランザクションを開始する
// callbackがtrueを返せばcommitし、falseを返せばrollbackする。
func (rs *RoomStatusManager) txW(callback func(*bolt.Bucket) bool) error {
	tx, err := rs.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	bucket := tx.Bucket(BUCKET_NAME)
	if bucket == nil {
		return errors.New(fmt.Sprintf("Bucket is not exists: %s", BUCKET_NAME))
	}
	if callback(bucket) == false {
		return nil
	}
	tx.Commit()
	return nil
}

func (rs *RoomStatusManager) GetMyVote(sf SessionFunc, id string) (*MyVote, error) {
	var err error
	var vote *MyVote

	sf(func(r *http.Request, w *http.ResponseWriter, store *sessions.CookieStore) {
		s, err := store.Get(r, getSessionName(id))
		if err != nil {
			return
		}

		if s.Values["vote"] == nil || s.Values["timestamp"] == nil {
			// セッションが存在しない場合
			vote = nil
		} else {
			vote = &MyVote{
				Vote:      s.Values["vote"].(string),
				Timestamp: s.Values["timestamp"].(int64),
			}
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
	if err := callback(stat); err != nil {
		return err
	}
	if err := rs.txW(func(bucket *bolt.Bucket) bool {
		js, err := json.Marshal(stat)
		if err != nil {
			return false
		}
		bucket.Put([]byte(id), js)
		return true
	}); err != nil {
		return err
	}
	return nil
}

func (rs *RoomStatusManager) Vote(sf SessionFunc, id string, hot, comfort, cold int) error {
	var err error
	if !(hot == 1 || comfort == 1 || cold == 1) {
		return errors.New("Invalid args")
	}

	sf(func(r *http.Request, w *http.ResponseWriter, store *sessions.CookieStore) {
		var s *sessions.Session
		s, err = store.Get(r, getSessionName(id))
		if err != nil {
			err = nil
			s = sessions.NewSession(store, getSessionName(id))
			s.Options = &sessions.Options{
				Path:     "/",
				HttpOnly: true,
			}
		}

		// 以前の投票を取り消す
		if s.Values["vote"] != nil {
			switch s.Values["vote"].(string) {
			case "hot":
				hot -= 1
			case "comfort":
				comfort -= 1
			case "cold":
				cold -= 1
			}
		}

		// 投票結果をCookieに保存
		if hot > 0 {
			s.Values["vote"] = "hot"
		} else if comfort > 0 {
			s.Values["vote"] = "comfort"
		} else if cold > 0 {
			s.Values["vote"] = "cold"
		}
		s.Values["timestamp"] = time.Now().Unix()

		err = s.Save(r, *w)
	})

	if err != nil {
		return err
	}

	return rs.setter(id, func(status *RoomStatus) error {
		status.Hot = uint(int(status.Hot) + hot)
		status.Comfort = uint(int(status.Comfort) + comfort)
		status.Cold = uint(int(status.Cold) + cold)
		return nil
	})
}

func (rs *RoomStatusManager) updateStatusWorker(ctx context.Context) {
	fmt.Println("starting UpdateStatusWorker")

	tick := time.NewTicker(INTERVAL)
	for {
		fmt.Println("update status")
		// room1
		if err := rs.updateStatus("room1", "Room1_yuuki"); err != nil {
			fmt.Println(err)
		}

		// room2
		if err := rs.updateStatus("room2", "Room2_yuuki"); err != nil {
			fmt.Println(err)
		}

		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func (rs *RoomStatusManager) updateStatus(roomId, thingName string) error {
	prop, err := rs.thingworx.Properties(thingName)
	if err != nil {
		return err
	}
	temp, err := prop.M("temperature").Float64()
	if err != nil {
		return err
	}
	isConnected, err := prop.M("isConnected").Bool()
	if err != nil {
		return err
	}

	if err := rs.setter(roomId, func(stat *RoomStatus) error {
		stat.Temperature = float32(temp)
		stat.IsConnected = isConnected
		return nil
	}); err != nil {
		return err
	}
	return nil
}
