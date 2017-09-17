package main

import (
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
	strid := req.Header.Get(SESSION_ID_COOKIE)
	id, err := strconv.ParseUint(strid, 10, 64)
	if err != nil {
		return nil
	}

	s := &Session{
		SessionID: id,
		req:       req,
		w:         w,
		tx:        tx,
		writen:    true,
	}

	row := tx.QueryRow(`
		SELECT secret_sha256 FROM s
		WHERE session_id=? AND expire>=?
	`, id, time.Now())
	var hashedSecret string
	if row.Scan(&hashedSecret) != nil {
		return nil
	}
	secretSHA256 := sha256.Sum256([]byte(req.Header.Get(SESSION_SECRET_COOKIE)))
	if hashedSecret != hex.EncodeToString(secretSHA256[:]) {
		return nil
	}
	return s
}

func NewSession(w http.ResponseWriter, req *http.Request, tx *sql.Tx) (*Session, error) {
	// generate secret
	randomData := make([]byte, 32)
	if _, err := rand.Read(randomData); err != nil {
		return nil, err
	}
	secret := hex.EncodeToString(randomData)
	secretSHA256 := sha256.Sum256([]byte(secret))

	if _, err := tx.Exec(`
		INSERT INTO s(
			secret_sha256,
			timestamp
		) VALUES (?, ?)`,
		hex.EncodeToString(secretSHA256[:]),
		time.Now(),
	); err != nil {
		return nil, err
	}
	return &Session{
		SessionID: 0,
		req:       req,
		w:         w,
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
