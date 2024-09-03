package main

import (
	"database/sql"
	"log"
	"time"
)

const (
	LogInfo int = iota
	LogWarning
	LogError
)

type Log struct {
	ID int
	Type int
	Text string
	CreatedAt time.Time
}

func createLogsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS logs(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		type INTEGER,
		text TEXT,
		created_at DATETIME
	);`)
	if err != nil {
		log.Fatalln("Unable to create logs table", err.Error())
		panic(err)
	}
}
