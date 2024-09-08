package main

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func GetAuthSession(store *sessions.CookieStore, r *http.Request) *sessions.Session {
	session, _ := store.Get(r, "mqtt-studio-session")

	if auth, ok := session.Values["authenticated"].(bool); !ok || !auth {
		return nil
	}

	return session
}
