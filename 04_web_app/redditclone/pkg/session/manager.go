package session

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/dgrijalva/jwt-go"
)

var Key = []byte("osfhvjfblkvbke")

type SessionsManager struct {
	data map[string]*Session
	mu   *sync.RWMutex
}

func NewSessionsManager() *SessionsManager {
	return &SessionsManager{
		data: make(map[string]*Session, 10),
		mu:   &sync.RWMutex{},
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
		user := (claims["user"]).(map[string]interface{})
		
		sm.mu.Lock()
		sess, ok := sm.data[user["id"].(string)]
		sm.mu.Unlock()

		if !ok {
			return nil, ErrNoAuth
		}

		return sess, nil
	}
	return nil, ErrNoAuth
}

func (sm *SessionsManager) Create(w http.ResponseWriter, userID uint64, login string) (*Session, error) {
	sess := NewSession(userID, login)

	sm.mu.Lock()
	sm.data[strconv.Itoa(int(userID))] = sess
	sm.mu.Unlock()

	return sess, nil
}