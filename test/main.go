package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	dsn := "file:app.db?mode=rwc&_foreign_keys=on"

	db1, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Database connection successful.")

	if err := createTable(db1); err != nil {
		log.Fatal(err)
	}
	_ = db1.Close()

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	id, err := insertUser(db, "Boss")
	if err != nil {
		log.Fatal(err)
	}

	name, err := getUserById(db, id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User ID: %d, Name: %s\n", id, name)
}

func getUserById(db *sql.DB, id int64) (string, error) {
	var name string
	err := db.QueryRow(`select name from users where id = ?`, id).Scan(&name)
	return name, err
}

func insertUser(db *sql.DB, name string) (int64, error) {
	res, err := db.Exec(`insert into users(name) values (?)`, name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(`create table if not exists users (
	id integer primary key autoincrement,
	name text not null
	);
	`)
	return err
}
