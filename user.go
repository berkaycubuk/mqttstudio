package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
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

func loginHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			session, _ := store.Get(r, "mqtt-studio-session")

			if auth, ok := session.Values["authenticated"].(bool); ok || auth {
				http.Redirect(w, r, "/projects", http.StatusFound)
				return
			}

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
			var hashedPassword string
			err := db.QueryRow("SELECT id, name, email, password FROM users WHERE email = ?", email).Scan(&user.ID, &user.Name, &user.Email, &hashedPassword)
			if err != nil {
				http.NotFound(w, r)
				return
			}

			if !VerifyUserPassword(password, hashedPassword) {
				http.NotFound(w, r)
				return
			}

			session, _ := store.Get(r, "mqtt-studio-session")
			session.Values["authenticated"] = true
			session.Values["user_id"] = user.ID
			session.Options = &sessions.Options{
				Path: "/",
				MaxAge: 3600, // 1 hour
				HttpOnly: true,
			}
			err = session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/projects", http.StatusFound)
		} else {
			fmt.Fprint(w, "Only GET and POST methods are supported.")
		}
	}
}

func signupHandler(db *sql.DB, store *sessions.CookieStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			session, _ := store.Get(r, "mqtt-studio-session")

			if auth, ok := session.Values["authenticated"].(bool); ok || auth {
				http.Redirect(w, r, "/projects", http.StatusFound)
				return
			}

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

			hashedPassword, err := HashUserPassword(password)
			if err != nil {
				log.Fatal(err)
				return
			}

			stmt, err := db.Prepare("INSERT INTO users(name,email,password) VALUES(?,?,?)")
			if err != nil {
				log.Fatal(err)
				return
			}

			res, err := stmt.Exec(name, email, hashedPassword)
			if err != nil {
				log.Fatal(err)
				return
			}

			userID, err := res.LastInsertId()
			if err != nil {
				log.Fatal(err)
				return
			}

			// Create team
			teamID, err := CreateTeam(db, name + "'s Team")
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = CreateTeamUser(db, teamID, int(userID), "OWNER")
			if err != nil {
				log.Fatal(err)
				return
			}

			session, _ := store.Get(r, "mqtt-studio-session")
			session.Values["authenticated"] = true
			session.Values["user_id"] = int(userID)
			session.Options = &sessions.Options{
				Path: "/",
				MaxAge: 3600, // 1 hour
				HttpOnly: true,
			}
			err = session.Save(r, w)

			http.Redirect(w, r, "/projects", http.StatusFound)
		} else {
			fmt.Fprint(w, "Only GET and POST methods are supported.")
		}
	}
}

func HashUserPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func VerifyUserPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
