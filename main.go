package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/BurntSushi/toml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"github.com/gorilla/sessions"
)

// TODO: Get key from the env
var store = sessions.NewCookieStore([]byte("secret-key"))

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
		createDataLogsTable(db)
		createTeamsTable(db)
	}

	var connections []*Connection

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.MustLoadMessageFile("./lang/core.en.toml")

	localizer := i18n.NewLocalizer(bundle, "en")

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/login", loginHandler(db, store))
	mux.HandleFunc("/signup", signupHandler(db, store))
	mux.HandleFunc("/logout", logoutHandler(db, store))

	// Project routes
	mux.HandleFunc("/projects", projectsHandler(db, localizer, store))
	mux.HandleFunc("/projects/{slug}", projectViewHandler(db, localizer, store))
	mux.HandleFunc("/projects/{slug}/new-section", projectNewSectionHandler(db, store))
	mux.HandleFunc("/projects/{slug}/new-widget", projectNewWidgetHandler(db, store))
	mux.HandleFunc("/projects/{slug}/connect", projectConnectHandler(db, &connections, store))
	mux.HandleFunc("/projects/{slug}/disconnect", projectDisconnectHandler(db, &connections, store))
	mux.HandleFunc("/projects/{slug}/connection", projectConnectionHandler(db, &connections, store))
	mux.HandleFunc("/projects/{slug}/data", projectDataHandler(db, &connections, store))
	mux.HandleFunc("/projects/{slug}/submit-value", projectSubmitValueHandler(db, &connections, store))
	mux.HandleFunc("/projects/{slug}/delete-widget", projectDeleteWidgetHandler(db, store))
	mux.HandleFunc("/projects/{slug}/edit-widget", projectEditWidgetHandler(db, store))
	mux.HandleFunc("/projects/{slug}/edit-section", projectEditSectionHandler(db, store))
	mux.HandleFunc("/projects/{slug}/delete-section", projectDeleteSectionHandler(db, store))
	mux.HandleFunc("/projects/{slug}/settings", projectSettingsViewHandler(db, store))

	// Account routes
	mux.HandleFunc("/account", accountHandler(db, store))

	// Admin routes
	mux.HandleFunc("/admin/projects", adminProjectsHandler(db, store))
	mux.HandleFunc("/admin/projects/new", adminNewProjectHandler(db, store))
	mux.HandleFunc("/admin/projects/delete", adminDeleteProjectHandler(db, store))

	// General routes ?
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./public"))))

	log.Println("Server started on port 8090")
	http.ListenAndServe(":8090", mux)
}
