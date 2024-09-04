package main

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func homeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}

		rows, err := db.Query("SELECT * FROM projects")
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		var projects []Project

		for rows.Next() {
			var project Project
			err = rows.Scan(&project.ID, &project.Name, &project.Slug)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			projects = append(projects, project)
		}

		rows.Close()

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/index.html"))
		tmpl.Execute(w, projects)
	}
}

func main() {
	isDbNew := false

	if _, err := os.Stat("./core.db"); errors.Is(err, os.ErrNotExist) {
		f, err := os.Create("./core.db")
		if err != nil {
			panic(err)
		}
		f.Close()

		isDbNew = true
	}

	db, err := sql.Open("sqlite3", "./core.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if isDbNew { // run migrations
		createProjectsTable(db)
		createUsersTable(db)
		createLogsTable(db)
	}

	var connections []*Connection

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/login", loginHandler(db))
	mux.HandleFunc("/signup", signupHandler(db))

	mux.HandleFunc("/add-to-cart", addToCartHandler(db))
	mux.HandleFunc("/cart", cartHandler(db))
	mux.HandleFunc("/projects/{slug}", projectViewHandler(db))
	mux.HandleFunc("/projects/{slug}/new-section", projectNewSectionHandler(db))
	mux.HandleFunc("/projects/{slug}/new-widget", projectNewWidgetHandler(db))
	mux.HandleFunc("/projects/{slug}/connect", projectConnectHandler(db, &connections))
	mux.HandleFunc("/projects/{slug}/disconnect", projectDisconnectHandler(db, &connections))
	mux.HandleFunc("/projects/{slug}/connection", projectConnectionHandler(db, &connections))
	mux.HandleFunc("/projects/{slug}/data", projectDataHandler(db, &connections))

	// Admin routes
	mux.HandleFunc("/admin/projects", adminProjectsHandler(db))
	mux.HandleFunc("/admin/projects/new", adminNewProjectHandler(db))

	mux.HandleFunc("/admin/products", adminProductsHandler(db))
	mux.HandleFunc("/admin/products/{id}", adminEditProductHandler(db))
	mux.HandleFunc("/admin/delete-product", adminDeleteProductHandler(db))
	mux.HandleFunc("/admin/new-product", adminNewProductHandler(db))

	// General routes ?
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))
	mux.HandleFunc("/", homeHandler(db))

	log.Println("Server started on port 8090")
	http.ListenAndServe(":8090", mux)
}
