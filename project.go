package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
)

type Project struct {
	ID int
	Name string
	Slug string
}

func createProjectsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS projects(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create products table", err.Error())
		panic(err)
	}
}

func projectViewHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		slugParameter := req.PathValue("slug")

		if slugParameter == "" {
			http.NotFound(w, req)
			return
		}

		var project Project
		err := db.QueryRow("SELECT * FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug)
		if err != nil {
			http.NotFound(w, req)
			return
		}

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/project.html"))
		tmpl.Execute(w, project)
	}
}
