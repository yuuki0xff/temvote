package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

const (
	SESSION_ID_COOKIE     = "temvote_session_id"
	SESSION_SECRET_COOKIE = "temvote_session_secret"
	COOKIE_MAX_AGE        = 600 // means 10 minutes
)

type SessionID uint64

type Session struct {
	SessionID uint64

	req    *http.Request
	w      http.ResponseWriter
	tx     *sql.Tx
	writen bool
	secret string
}

func GetSession(w http.ResponseWriter, req *http.Request, tx *sql.Tx) *Session {
	cookie, err := req.Cookie(SESSION_ID_COOKIE)
	if err != nil {
		return nil
	}
	id, err := strconv.ParseUint(cookie.Value, 10, 64)
	if err != nil {
		return nil
	}

	row := tx.QueryRow(`
		SELECT secret_sha256 FROM session
		WHERE session_id=? AND expire>=?
	`, id, time.Now())
	var tmp string
	if row.Scan(&tmp) != nil {
		return nil
	}
	hashedSecret, err := hex.DecodeString(tmp)
	if err != nil {
		return nil
	}

	cookie, err = req.Cookie(SESSION_SECRET_COOKIE)
	if err != nil {
		return nil
	}
	secret := cookie.Value
	randomData, err := hex.DecodeString(cookie.Value)
	if err != nil {
		return nil
	}
	hashedSecret2 := sha256.Sum256(randomData)
	if bytes.Compare(hashedSecret, hashedSecret2[:]) != 0 {
		return nil
	}

	return &Session{
		SessionID: id,
		req:       req,
		w:         w,
		tx:        tx,
		writen:    true,
		secret:    secret,
	}
}

func NewSession(w http.ResponseWriter, req *http.Request, tx *sql.Tx) (*Session, error) {
	// generate secret
	randomData := make([]byte, 32)
	if _, err := rand.Read(randomData); err != nil {
		return nil, err
	}
	secret := hex.EncodeToString(randomData)
	secretSHA256 := sha256.Sum256([]byte(randomData))

	if _, err := tx.Exec(`
		INSERT INTO session(
			secret_sha256,
			expire
		) VALUES (?, ?)`,
		hex.EncodeToString(secretSHA256[:]),
		time.Now().Add(COOKIE_MAX_AGE*time.Second),
	); err != nil {
		return nil, err
	}

	var sid uint64
	row := tx.QueryRow(`SELECT LAST_INSERT_ID()`)
	if err := row.Scan(&sid); err != nil {
		return nil, err
	}
	return &Session{
		SessionID: sid,
		req:       req,
		w:         w,
		tx:        tx,
		writen:    false,
		secret:    secret,
	}, nil
}

// 既存のCookieの有効期限を延長する
func (s *Session) ExtendExpiration() error {
	s.Save()

	if _, err := s.tx.Exec(`
		UPDATE session SET expire=? WHERE session_id=?`,
		time.Now().Add(COOKIE_MAX_AGE*time.Second),
		s.SessionID,
	); err != nil {
		return err
	}
	return nil
}

func (s *Session) Save() {
	// TODO: add secure attribute
	http.SetCookie(s.w, &http.Cookie{
		Name:     SESSION_ID_COOKIE,
		Value:    strconv.FormatUint(s.SessionID, 10),
		MaxAge:   COOKIE_MAX_AGE,
		HttpOnly: true,
	})
	http.SetCookie(s.w, &http.Cookie{
		Name:     SESSION_SECRET_COOKIE,
		Value:    s.secret,
		MaxAge:   COOKIE_MAX_AGE,
		HttpOnly: true,
	})
	s.writen = true
}
