package main

import (
	"database/sql"
	"log"
)

type User struct {
	ID int
	Name string
	Email string
}

func createUsersTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create products table", err.Error())
		panic(err)
	}
}
