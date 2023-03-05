package user

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrExistUser  = errors.New("already exist")
	ErrNoUser  = errors.New("user not found")
	ErrBadPass = errors.New("invalid password")
)

type UserMemoryRepository struct {
	dbUsers *sql.DB
}

func NewMemoryRepo(db *sql.DB) *UserMemoryRepository {
	return &UserMemoryRepository{
		dbUsers: db,
	}
}

func (repo *UserMemoryRepository) Authorize(login, pass string) (*User, error) {
	u := &User{}
	row := repo.dbUsers.QueryRow("SELECT id, password FROM users WHERE login = ?", login)
	err := row.Scan(&u.ID, &u.password)
	u.Login = login
	if err == sql.ErrNoRows {
		return nil, ErrNoUser
	}
	if u.password != pass {
		return nil, ErrBadPass
	}
	return u, nil
}

func(repo *UserMemoryRepository) Register(login, pass string) (*User, error){
	user := &User{Login: login, password: pass}
		row := repo.dbUsers.QueryRow("SELECT id FROM users WHERE login = ?", login)
		var id int
		err := row.Scan(&id)
	if(err == sql.ErrNoRows){
		row = repo.dbUsers.QueryRow("SELECT id FROM users ORDER BY id DESC LIMIT 1")
		err := row.Scan(&id)
		if(err == sql.ErrNoRows){
			id = 0
		}else{
			id++
		}
		_, err = repo.dbUsers.Exec("INSERT INTO users (`id`, `login`, `password`) VALUES (?, ?, ?)", id, login, pass)				
		if err != nil{
			return nil, errors.New("Error exec of DB: " + err.Error())
		}
		user.ID = uint64(id)
	}else{
		return nil, ErrExistUser
	}
	
	return user, nil
}