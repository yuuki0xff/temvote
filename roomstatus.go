package main

import (
	"sync"
	"fmt"
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
