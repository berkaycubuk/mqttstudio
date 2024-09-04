package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

/*
*	TODO: Sesssion
*/

type User struct {
	ID int
	Name string
	Email string
	//Password string
}

func createUsersTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		password TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create products table", err.Error())
		panic(err)
	}
}

func loginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl := template.Must(template.ParseFiles("./views/auth-layout.html", "./views/login.html"))
			tmpl.Execute(w, "")
		} else if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}

			email := r.FormValue("email")
			password := r.FormValue("password")

			var user User
			// TODO: add password hashing!!!
			err := db.QueryRow("SELECT id, name, email FROM users WHERE email = ? AND password = ?", email, password).Scan(&user.ID, &user.Name, &user.Email)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			fmt.Fprint(w, "Only GET and POST methods are supported.")
		}
	}
}

func signupHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			tmpl := template.Must(template.ParseFiles("./views/auth-layout.html", "./views/signup.html"))
			tmpl.Execute(w, "")
		} else if r.Method == "POST" {
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}

			name := r.FormValue("name")
			email := r.FormValue("email")
			password := r.FormValue("password")

			foundId := 0
			// check is there a user with the same email
			_ = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&foundId)

			if foundId != 0 { // email taken
				// TODO: notify user
				return
			}

			// TODO: add password hashing!!!
			stmt, err := db.Prepare("INSERT INTO users(name,email,password) VALUES(?,?,?)")
			if err != nil {
				log.Fatal(err)
				return
			}

			res, err := stmt.Exec(name, email, password)
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = res.LastInsertId()
			if err != nil {
				log.Fatal(err)
				return
			}

			http.Redirect(w, r, "/", http.StatusFound)
		} else {
			fmt.Fprint(w, "Only GET and POST methods are supported.")
		}
	}
}
