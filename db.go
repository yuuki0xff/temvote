package main

import (
	"database/sql"
	"strconv"
	"time"
)

type VoteChoice string
type VoteID uint64

const (
	Hot     = VoteChoice("hot")
	Comfort = VoteChoice("comfort")
	Cold    = VoteChoice("cold")
)

type Vote struct {
	VoteID    VoteID
	RoomID    RoomID
	S         *Session
	Choice    VoteChoice
	Timestamp time.Time
}

// 投票内容を変更する。VoteID, RoomID, Sが指定されていなければならない。
// 初投票の場合は、VoteIDはデフォルト値(VoteID(0))に設定すること。
func (v *Vote) UpdateChoice(tx *sql.Tx, choice VoteChoice) error {
	now := time.Now()
	if v.VoteID == VoteID(0) {
		// 初投票の場合
		if _, err := tx.Exec(`
			INSERT INTO vote(
				session_id, room_id, choice, timestamp
			) VALUES (?, ?, ?, ?)`,
			v.S.SessionID, v.RoomID, string(choice), now,
		); err != nil {
			return err
		}
	} else {
		// 投票内容を変更する場合
		if _, err := tx.Exec(`
			UPDATE vote SET choice=?, timestamp=? WHERE vote_id=?`,
			string(choice), now, v.VoteID,
		); err != nil {
			return err
		}
	}
	v.Choice = choice
	v.Timestamp = now
	return nil
}

type RoomID uint64
type BuildingName string
type FloorID int64

type Room struct {
	RoomID       RoomID
	Name         string
	BuildingName BuildingName
	FloorID      FloorID
}

func StringToRoomID(strid string) (id RoomID, err error) {
	var i uint64
	i, err = strconv.ParseUint(strid, 10, 64)
	id = RoomID(i)
	return
}
