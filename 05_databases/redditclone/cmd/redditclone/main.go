package main

import (
	"context"
	"fmt"
	"hw6/pkg/handlers"
	"hw6/pkg/items"
	"hw6/pkg/middleware"
	"hw6/pkg/session"
	"hw6/pkg/user"
	"net/http"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	//mgo "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func notFound(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "../../static/html/index.html")
}

func main() {
	
	// основные настройки к базе
	dsn := "root:love@tcp(localhost:3306)/golang?"
	// указываем кодировку
	dsn += "&charset=utf8"
	// отказываемся от prapared statements
	// параметры подставляются сразу
	dsn += "&interpolateParams=true"

	dbSession, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	dbUsers, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	dbSession.SetMaxOpenConns(10)
	dbUsers.SetMaxOpenConns(10)

	err = dbUsers.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}
	err = dbSession.Ping() // вот тут будет первое подключение к базе
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		panic(err)
	}
	collection := client.Database("coursera").Collection("items")
	// если коллекции не будет, то она создасться автоматически
	//collection := sess.DB("coursera").C("items")

	posts := items.NewMemoryRepo(collection)
	users := user.NewMemoryRepo(dbUsers)
	sm := session.NewSessionsManager(dbSession)
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
