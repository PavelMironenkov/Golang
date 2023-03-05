package session

import (
	"database/sql"
	"errors"
	// "fmt"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/dgrijalva/jwt-go"
)

var Key = []byte("osfhvjfblkvbke")

type SessionsManager struct {
	dbSession *sql.DB
}

func NewSessionsManager(dbSession *sql.DB) *SessionsManager {
	return &SessionsManager{
		dbSession: dbSession,
	}
}

func (sm *SessionsManager) Check(w http.ResponseWriter, r *http.Request) (*Session, error) {
	tokenString := r.Header.Get("Authorization")
	_, tokenString, ok := strings.Cut(tokenString, "Bearer ")
	if !ok {
		return nil, ErrNoAuth
	}

	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) { return Key, nil })
	if err == nil {
		user := (claims["user"]).(map[string] interface{})
		
		row := sm.dbSession.QueryRow("SELECT id, login FROM session WHERE id = ?", user["id"].(string))
		id, err := strconv.Atoi(user["id"].(string))
		if err != nil{
			return nil, errors.New("DBsession error: bad id")
		}
		sess := &Session{}
		err = row.Scan(&sess.UserID, &sess.Login)
		if err == sql.ErrNoRows || sess.Login != user["username"].(string) || sess.UserID != uint64(id) {
			return nil, ErrNoAuth
		}

		return sess, nil
	}
	return nil, ErrNoAuth
}

func (sm *SessionsManager) Create(w http.ResponseWriter, userID uint64, login string) (*Session, error) {
	sess := NewSession(userID, login)
	row := sm.dbSession.QueryRow("SELECT login FROM session WHERE id ?", userID)
	var Login string
	err := row.Scan(&Login)
	if err == sql.ErrNoRows{
		_, err = sm.dbSession.Exec("INSERT INTO session (`id`, `login`) VALUES (?, ?)",  sess.UserID, sess.Login)
		if err != nil{
			return nil, err
		}
	}
	return sess, nil
}