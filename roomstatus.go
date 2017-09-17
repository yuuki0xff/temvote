package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"
)

type RoomNameMap map[RoomID]string
type RoomGroupMap map[BuildingName]map[FloorID][]RoomID

const (
	INTERVAL = 1 * time.Minute
)

type RoomStatus struct {
	RoomID  RoomID         `json:"id"`
	Sensors []SensorStatus `json:"sensors"`

	Hot     uint64 `json:"hot"`
	Comfort uint64 `json:"comfort"`
	Cold    uint64 `json:"cold"`
	lock    sync.RWMutex
}

type MyVote struct {
	Vote      VoteChoice `json:"vote"`
	Timestamp int64      `json:"timestamp"`
}

type RoomStatusManager struct {
	db        *sql.DB
	thingworx *ThingWorxClient

	sensorCache map[RoomID]map[ThingName]SensorStatus
	cacheLock   sync.RWMutex
}

type RoomStatusTx struct {
	rsm *RoomStatusManager
	tx  *sql.Tx
	s   *Session
}

type SensorStatus struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	IsConnected bool    `json:"isConnected"`

	expire time.Time
}

func NewRoomStatusManager(db *sql.DB, thingworx *ThingWorxClient, ctx context.Context) *RoomStatusManager {
	// create RSM
	rs := &RoomStatusManager{}
	rs.db = db
	rs.thingworx = thingworx
	rs.sensorCache = make(map[RoomID]map[ThingName]SensorStatus)

	go rs.cacheUpdater(ctx)
	return rs
}

func (rsm *RoomStatusManager) GetTx(w http.ResponseWriter, req *http.Request) (*RoomStatusTx, error) {
	tx, err := rsm.db.Begin()
	if err != nil {
		return nil, err
	}
	s := GetSession(w, req, tx)
	if s == nil {
		s, err = NewSession(w, req, tx)
		if err != nil {
			return nil, err
		}
	}
	return &RoomStatusTx{
		rsm: rsm,
		tx:  tx,
		s:   s,
	}, nil
}

func (rst *RoomStatusTx) Rollback() error {
	return rst.tx.Rollback()
}

func (rst *RoomStatusTx) Commit() error {
	return rst.tx.Commit()
}

func (rst *RoomStatusTx) GetRoomName(id RoomID) (name string, err error) {
	err = rst.tx.QueryRow(
		`SELECT name FROM room
		WHERE room_id=?`,
		id,
	).Scan(&name)
	return
}

// 投票内容を取得する。未投票の場合はnilを返す
func (rst *RoomStatusTx) GetMyVote(id RoomID) (vote *MyVote, err error) {
	var v Vote
	if err = rst.tx.QueryRow(
		`SELECT choice, timestamp FROM vote
			WHERE session_id=? AND room_id=?`,
		rst.s.SessionID, id,
	).Scan(&v.Choice, &v.Timestamp); err != nil {
		if err == sql.ErrNoRows {
			// 未投票の状態。
			err = nil
			return
		}
		return
	}

	vote = &MyVote{
		Vote:      v.Choice,
		Timestamp: v.Timestamp.Unix(),
	}
	return vote, err
}

func (rst *RoomStatusTx) GetStatus(id RoomID) (*RoomStatus, error) {
	rs := &RoomStatus{
		RoomID: id,
	}

	var ok bool
	rs.Sensors, ok = rst.rsm.getSensorStatusFromCache(id)
	if !ok {
		// センサーの状態を更新できていない状態。
		rs.Sensors = []SensorStatus{}
	}

	rows, err := rst.tx.Query(
		`SELECT choice, count(vote_id) FROM vote
		WHERE room_id=?
		GROUP BY choice`,
		id,
	)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var choice VoteChoice
		var count uint64
		if err := rows.Scan(&choice, &count); err != nil {
			return nil, err
		}
		switch choice {
		case Hot:
			rs.Hot = count
		case Comfort:
			rs.Comfort = count
		case Cold:
			rs.Cold = count
		}
	}
	return rs, nil
}

func (rst *RoomStatusTx) Vote(id RoomID, choice VoteChoice) error {
	vote := Vote{
		RoomID: id,
		S:      rst.s,
	}

	if err := rst.tx.QueryRow(
		`SELECT vote_id FROM vote
		WHERE session_id=? AND room_id=?`,
		rst.s.SessionID, id,
	).Scan(&vote.VoteID); err != nil {
		switch err {
		case sql.ErrNoRows:
			// 未投票であることを表す、0を代入
			vote.VoteID = VoteID(0)
		default:
			return err
		}
	}

	return vote.UpdateChoice(rst.tx, choice)
}

func (rst *RoomStatusTx) GetAllRoomsInfo() (names RoomNameMap, groups RoomGroupMap, err error) {
	// NOTE: roomテーブルの行数は少ないことを想定しているため、テーブルスキャンをしている。
	{
		names = make(RoomNameMap)
		var rows *sql.Rows
		rows, err = rst.tx.Query(`
			SELECT room_id, name FROM room
		`)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			var id RoomID
			var name string
			if err = rows.Scan(&id, &name); err != nil {
				return
			}
			names[id] = name
		}
	}

	{
		groups = make(RoomGroupMap)
		var rows *sql.Rows
		rows, err = rst.tx.Query(`
			SELECT building_name, floor, room_id FROM room
			GROUP BY building_name, floor, room_id
		`)
		defer rows.Close()
		for rows.Next() {
			var bname BuildingName
			var floor FloorID
			var id RoomID
			if err = rows.Scan((*string)(&bname), &floor, &id); err != nil {
				return
			}

			if _, ok := groups[bname]; !ok {
				groups[bname] = make(map[FloorID][]RoomID)
			}
			if _, ok := groups[bname][floor]; !ok {
				groups[bname][floor] = make([]RoomID, 0)
			}
			groups[bname][floor] = append(groups[bname][floor], id)
		}
	}
	return
}

func (rsm *RoomStatusManager) getSensorStatusFromCache(id RoomID) ([]SensorStatus, bool) {
	rsm.cacheLock.RLock()
	defer rsm.cacheLock.RUnlock()

	cache, ok := rsm.sensorCache[id]
	if ok {
		array := make([]SensorStatus, 0, len(cache))
		for i := range cache {
			if cache[i].expire.After(time.Now()) {
				array = append(array, cache[i])
			}
		}
		return array, len(array) > 0
	}
	return []SensorStatus{}, false
}

// すべてのセンサーの状態をキャッシュする
func (rsm *RoomStatusManager) cacheUpdater(ctx context.Context) {
	log.Println("starting UpdateStatusWorker")

	tick := time.NewTicker(INTERVAL)
	for {
		log.Println("update status")

		for err := range rsm.updateAllSensorStatuses() {
			log.Println(err)
		}

		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

func (rsm *RoomStatusManager) updateAllSensorStatuses() []error {
	errCh := make(chan error)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		tx, err := rsm.db.Begin()
		if err != nil {
			errCh <- err
			return
		}
		defer tx.Rollback()

		rows, err := tx.Query(
			`SELECT room_id, thing_name FROM thing`,
		)
		if err != nil {
			errCh <- err
			return
		}

		for rows.Next() {
			var id RoomID
			var name ThingName
			rows.Scan(&id, &name)

			// start async update
			wg.Add(1)
			go func(id RoomID, name ThingName) {
				defer wg.Done()
				if err := rsm.updateSensorStatus(id, name); err != nil {
					errCh <- err
					return
				}
			}(id, name)
		}
	}()

	go func() {
		wg.Wait()
		close(errCh)
	}()

	errs := []error{}
	for err := range errCh {
		errs = append(errs, err)
	}
	return errs
}

// センサーで測定した部屋の状態を、DBに反映する。
func (rsm *RoomStatusManager) updateSensorStatus(id RoomID, thingName ThingName) error {
	var stat SensorStatus

	prop, err := rsm.thingworx.Properties(thingName)
	if err != nil {
		return err
	}
	stat.Temperature, err = prop.M("temperature").Float64()
	if err != nil {
		return err
	}
	stat.Humidity, err = prop.M("humidity").Float64()
	if err != nil {
		return err
	}
	stat.IsConnected, err = prop.M("isConnected").Bool()
	if err != nil {
		return err
	}

	rsm.cacheLock.Lock()
	defer rsm.cacheLock.Unlock()
	if _, ok := rsm.sensorCache[id]; !ok {
		rsm.sensorCache[id] = map[ThingName]SensorStatus{}
	}
	rsm.sensorCache[id][thingName] = stat
	return nil
}
