package user

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

func TestAutorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	// good query
	rows := sqlmock.NewRows([]string{"id", "password"})
	login := "pavel"
	expect := []*User{
		{ID: 0, Login: login, password: "12345678"},
	}
	for _, user := range expect {
		rows = rows.AddRow(user.ID, user.password)
	}

	mock.
		ExpectQuery("SELECT id, password FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnRows(rows)

	repo := NewMemoryRepo(db)

	customer, err := repo.Authorize(login, "12345678")
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(customer, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], customer)
		return
	}

	// query error
	mock.
		ExpectQuery("SELECT id, password FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)
	_, err = repo.Authorize(login, "")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	//Bad password
	rows = sqlmock.NewRows([]string{"id", "password"})
	expect = []*User{
		{0, login, "12345678"},
	}
	for _, user := range expect {
		rows = rows.AddRow(user.ID, user.password)
	}
	mock.
		ExpectQuery("SELECT id, password FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnRows(rows)

	_, err = repo.Authorize(login, "qwertyui")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

}

func TestRegistrate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	// good query bd is empty
	login := "pavel"
	rows := sqlmock.NewRows([]string{"id"})
	expect := []*User{
		{ID: 0, Login: login, password: "12345678"},
	}
	for _, user := range expect {
		rows = rows.AddRow(user.ID)
	}

	mock.
		ExpectQuery("SELECT id FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectQuery("SELECT id FROM users ORDER BY id DESC LIMIT 1").
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(0, login, "12345678").
		WillReturnResult(sqlmock.NewResult(0, 1))
	repo := NewMemoryRepo(db)

	customer, err := repo.Register(login, "12345678")
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(customer, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], customer)
		return
	}

	//  good query bd isn't empty
	mock.
		ExpectQuery("SELECT id FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectQuery("SELECT id FROM users ORDER BY id DESC LIMIT 1").
		WillReturnRows(rows)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(1, login, "12345678").
		WillReturnResult(sqlmock.NewResult(0, 1))
	expect[0].ID = 1
	customer, err = repo.Register(login, "12345678")
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if !reflect.DeepEqual(customer, expect[0]) {
		t.Errorf("results not match, want %v, have %v", expect[0], customer)
		return
	}

	//exec error - db error
	mock.
		ExpectQuery("SELECT id FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectQuery("SELECT id FROM users ORDER BY id DESC LIMIT 1").
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(0, login, "12345678").
		WillReturnError(errors.New("There is no connection to the database"))
	_, err = repo.Register(login, "12345678")

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected err, got nil")
		return
	}

	//user is exist
	mock.
		ExpectQuery("SELECT id FROM users WHERE login = ?").
		WithArgs(login).
		WillReturnError(nil)
	_, err = repo.Register(login, "12345678")
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
	if err == nil {
		t.Errorf("expected err, got nil")
		return
	}
}
