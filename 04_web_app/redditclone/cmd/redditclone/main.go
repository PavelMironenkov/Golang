package main

import (
	"fmt"
	"hw5/pkg/handlers"
	"hw5/pkg/items"
	"hw5/pkg/middleware"
	"hw5/pkg/session"
	"hw5/pkg/user"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../../static/html/index.html")
}

func main() {
	posts := items.NewMemoryRepo()
	users := user.NewMemoryRepo()
	sm := session.NewSessionsManager()
	zapLogger, err := zap.NewProduction()
	if err != nil{
		fmt.Println(err)
	}
	defer zapLogger.Sync() // nolint
	logger := zapLogger.Sugar()

	postsHandler := handlers.ItemsHandler{
		ItemsRepo: posts,
		Logger:    logger,
	}
	userHandler := handlers.UserHandler{
		UserRepo: users,
		Sessions: sm,
		Logger:   logger,
	}
	r := mux.NewRouter()
	r.Handle("/", http.FileServer(http.Dir("../../static/html/")))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("../../static/"))))
	r.HandleFunc("/api/register", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/login", userHandler.Login).Methods("POST")
	r.HandleFunc("/api/posts/", postsHandler.List).Methods("GET")
	r.HandleFunc("/api/posts", postsHandler.Add).Methods("POST")
	r.HandleFunc("/api/posts/{CATEGORY_NAME}", postsHandler.GetAllAtTheCategory).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postsHandler.ListPost).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postsHandler.AddComment).Methods("POST")
	r.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", postsHandler.DeleteComment).Methods("DELETE")
	r.HandleFunc("/api/post/{POST_ID}/upvote", postsHandler.Vote).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/unvote", postsHandler.Unvote).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}/downvote", postsHandler.Vote).Methods("GET")
	r.HandleFunc("/api/user/{USER_LOGIN}", postsHandler.GetAllAtUser).Methods("GET")
	r.HandleFunc("/api/post/{POST_ID}", postsHandler.Delete).Methods("DELETE")
	r.NotFoundHandler = http.HandlerFunc(notFound)

	mux := middleware.Auth(sm, r)
	mux = middleware.AccessLog(logger, mux)
	mux = middleware.Panic(mux)
	

	addr := ":8080"
	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		fmt.Println(err)
	}
}
