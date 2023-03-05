package handlers

import (
	"encoding/json"
	"hw5/pkg/session"
	"hw5/pkg/user"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.uber.org/zap"
)

type UserHandler struct {
	Logger   *zap.SugaredLogger
	UserRepo user.UserRepo
	Sessions *session.SessionsManager
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	type LoginData struct {
		Username string
		Password string
	}
	if r.Header.Get("Content-Type") != "application/json" {
		jsonError(w, http.StatusBadRequest, "unknown payload")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "cant read request body")
	}
	r.Body.Close()

	fd := &LoginData{}
	err = json.Unmarshal(body, fd)
	if err != nil {
		jsonError(w, http.StatusBadRequest, "cant unpack payload")
		return
	}

	var user *user.User
	switch r.URL.Path {
	case "/api/register":
		user, err = h.UserRepo.Register(fd.Username, fd.Password)
	case "/api/login":
		user, err = h.UserRepo.Authorize(fd.Username, fd.Password)
	}

	if err == nil {
		sess, error := h.Sessions.Create(w, user.ID, user.Login)
		if error != nil {
			http.Error(w, `Session isn't create`, http.StatusInternalServerError)
			return
		}
		h.Logger.Infof("Created session for user with id: %v ", sess.UserID)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user": map[string]interface{}{
				"username": user.Login,
				"id":       strconv.Itoa(int(user.ID)),
			},
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
		})
		tokenString, error := token.SignedString(session.Key)
		if error != nil {
			jsonError(w, http.StatusInternalServerError, err.Error())
			return
		}
		resp, error := json.Marshal(map[string]interface{}{
			"token":       tokenString,
			"Status code": http.StatusFound,
		})
		if error != nil {
			http.Error(w, "Marshalling err", http.StatusBadRequest)
			return
		}
		_, err = w.Write(resp)
		if err != nil {
			http.Error(w, "Writing response err", http.StatusInternalServerError)
			return
		}
		h.Logger.Infof("Send token on client for user with id: %v ", sess.UserID)
	} else {
		var resp []byte
		var error error

		switch r.URL.Path {
		case "/api/register":
			w.WriteHeader(http.StatusUnprocessableEntity)
			errors := make([]map[string]string, 0)
			errors = append(errors, map[string]string{
				"location": "body",
				"msg": err.Error(),
				"param": "username",
				"value": fd.Username ,
			})
			resp, error = json.Marshal(map[string][]map[string]string{
				"errors": errors,
			})
		case "/api/login":
			w.WriteHeader(http.StatusUnauthorized)
			resp, error = json.Marshal(map[string]interface{}{
				"message": err.Error(),
			})
		}

		if error != nil {
			http.Error(w, "Marshalling err", http.StatusBadRequest)
			return
		}
		_, error = w.Write(resp)
		if error != nil {
			http.Error(w, "Writing response err", http.StatusInternalServerError)
		}
	}
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	resp, err := json.Marshal(map[string]interface{}{
		"status": status,
		"error":  msg,
	})
	if err != nil {
		http.Error(w, "Marshalling err", http.StatusBadRequest)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, "Writing response err", http.StatusInternalServerError)
	}
}
