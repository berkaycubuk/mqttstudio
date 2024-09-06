package main

import (
	"database/sql"
	"log"
	"time"
)

type DataLog struct {
	ID			int
	Topic		string
	Data		[]byte
	CreatedAt	time.Time
}

func createDataLogsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS data_logs(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		topic TEXT NOT NULL,
		data BLOB,
		created_at DATETIME
	);`)
	if err != nil {
		log.Fatalln("Unable to create data_logs table", err.Error())
		panic(err)
	}
}

func createDataLog(db *sql.DB, topic string, data []byte) {
	stmt, err := db.Prepare("INSERT INTO data_logs(topic, data, created_at) VALUES(?,?,?)")
	if err != nil {
		log.Fatal(err)
		return
	}

	res, err := stmt.Exec(topic, data, time.Now())
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = res.LastInsertId()
	if err != nil {
		log.Fatal(err)
		return
	}
}
