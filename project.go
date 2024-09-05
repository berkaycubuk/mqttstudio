package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

type Project struct {
	ID int
	Name string
	Slug string
	BrokerClientID string
	BrokerAddress string
	BrokerPort int
	BrokerProtocol string
	//Sections []ProjectSection
}

type ProjectSection struct {
	ID int
	ProjectID int
	Name string
	Widgets []ProjectWidget
}

type ProjectWidget struct {
	ID int
	ProjectSectionID int
	Widget string
	Config []byte
	ConfigParsed any
}

type ProjectViewData struct {
	Project		Project
	Sections	[]ProjectSection
}

type TextWidgetConfig struct {
	Topic string
}

type ButtonWidgetConfig struct {
	Topic string
	Message string
}

type WidgetData struct {
	ID int
	Data any
}

func createProjectsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS projects(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		slug TEXT NOT NULL,
		broker_client_id TEXT NOT NULL,
		broker_address TEXT NOT NULL,
		broker_port INTEGER NOT NULL,
		broker_protocol TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create projects table", err.Error())
		panic(err)
	}

	createProjectSectionsTable(db)
	createProjectWidgetsTable(db)
}

func createProjectSectionsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS project_sections(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		name TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create project sections table", err.Error())
		panic(err)
	}
}

func createProjectWidgetsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS project_widgets(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		project_section_id INTEGER NOT NULL,
		widget TEXT,
		config BLOB
	);`)
	if err != nil {
		log.Fatalln("Unable to create project widgets table", err.Error())
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
		err := db.QueryRow("SELECT id, name, slug FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug)
		if err != nil {
			http.NotFound(w, req)
			return
		}

		rows, err := db.Query("SELECT id, name FROM project_sections WHERE project_id = ?", project.ID)
		if err != nil {
			log.Fatal(err)
			return
		}

		var projectSections []ProjectSection
		for rows.Next() {
			var projectSection ProjectSection

			err = rows.Scan(&projectSection.ID, &projectSection.Name)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			// Query the widgets that are related to this section
			widgetRows, err := db.Query("SELECT id, widget, config FROM project_widgets WHERE project_section_id = ?", projectSection.ID)
			if err != nil {
				log.Fatal(err)
				return
			}

			var projectWidgets []ProjectWidget
			for widgetRows.Next() {
				var projectWidget ProjectWidget

				err = widgetRows.Scan(&projectWidget.ID, &projectWidget.Widget, &projectWidget.Config)
				if err != nil {
					fmt.Fprintf(w, err.Error())
					continue
				}

				/*
				// parse config
				if projectWidget.Config != nil && len(projectWidget.Config) > 0 {
					if projectWidget.Widget == "TEXT" {
					}
				}
				*/

				projectWidgets = append(projectWidgets, projectWidget)
			}

			widgetRows.Close()

			projectSection.Widgets = projectWidgets

			projectSections = append(projectSections, projectSection)
		}

		rows.Close()

		tmpl := template.Must(template.ParseFiles("./views/layout.html", "./views/project.html"))
		tmpl.Execute(w, ProjectViewData{
			Project: project,
			Sections: projectSections,
		})
	}
}

func adminProjectsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, name, slug FROM projects")
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

		tmpl := template.Must(template.ParseFiles("./views/admin/layout.html", "./views/admin/projects.html"))
		tmpl.Execute(w, projects)
	}
}

func adminNewProjectHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "GET" {
			tmpl := template.Must(template.ParseFiles("./views/admin/layout.html", "./views/admin/new-project.html"))
			tmpl.Execute(w, "")
		} else if req.Method == "POST" {
			if err := req.ParseForm(); err != nil {
				fmt.Fprintf(w, "ERROR: %v", err)
				return
			}

			name := req.FormValue("name")
			slug := req.FormValue("slug")
			brokerClientID := req.FormValue("broker-client-id")
			brokerAddress := req.FormValue("broker-address")
			brokerPort := req.FormValue("broker-port")
			brokerProtocol := req.FormValue("broker-protocol")

			stmt, err := db.Prepare("INSERT INTO projects(name,slug,broker_client_id,broker_address,broker_port,broker_protocol) VALUES(?,?,?,?,?,?)")
			if err != nil {
				log.Fatal(err)
				return
			}

			res, err := stmt.Exec(name, slug, brokerClientID, brokerAddress, brokerPort, brokerProtocol)
			if err != nil {
				log.Fatal(err)
				return
			}

			_, err = res.LastInsertId()
			if err != nil {
				log.Fatal(err)
				return
			}

			http.Redirect(w, req, "/admin/projects", http.StatusFound)
		} else {
			fmt.Fprintf(w, "Only GET and POST methods are supported.")
		}
	}
}

func projectNewSectionHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprint(w, "Only POST method is supported.")
			return
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR: %v", err)
			return
		}

		// TODO: validate fields
		id := r.FormValue("id")
		slug := r.PathValue("slug")
		name := r.FormValue("name")

		stmt, err := db.Prepare("INSERT INTO project_sections(project_id,name) VALUES(?,?)")
		if err != nil {
			log.Fatal(err)
			return
		}

		res, err := stmt.Exec(id, name)
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = res.LastInsertId()
		if err != nil {
			log.Fatal(err)
			return
		}

		http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
	}
}

func projectNewWidgetHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprint(w, "Only POST method is supported.")
			return
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR: %v", err)
			return
		}

		// TODO: validate fields
		id := r.FormValue("id")
		topic := r.FormValue("topic")
		slug := r.PathValue("slug")
		widget := r.FormValue("widget")

		if id == "" {
			http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
			return
		}

		if widget == "" {
			http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
			return
		}

		stmt, err := db.Prepare("INSERT INTO project_widgets(project_section_id, widget, config) VALUES(?,?,?)")
		if err != nil {
			log.Fatal(err)
			return
		}

		var config any

		if widget == "TEXT" {
			if topic == "" {
				http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
				return
			}

			config, err = json.Marshal(TextWidgetConfig{
				Topic: topic,
			})
		} else if widget == "BUTTON" {
			message := r.FormValue("message")
			if topic == "" {
				http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
				return
			}
			if message == "" {
				http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
				return
			}

			config, err = json.Marshal(ButtonWidgetConfig{
				Topic: topic,
				Message: message,
			})
		}

		res, err := stmt.Exec(id, widget, config)
		if err != nil {
			log.Fatal(err)
			return
		}

		_, err = res.LastInsertId()
		if err != nil {
			log.Fatal(err)
			return
		}

		http.Redirect(w, r, "/projects/" + slug, http.StatusFound)
	}
}

func projectConnectHandler(db *sql.DB, connections *[]*Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slugParameter := r.PathValue("slug")

		var project Project
		err := db.QueryRow("SELECT id, name, slug, broker_address, broker_port, broker_protocol FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug, &project.BrokerAddress, &project.BrokerPort, &project.BrokerProtocol)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var connection *Connection

		// find connection
		connectionFound := false
		for i := 0; i < len(*connections); i++ {
			if (*connections)[i].ProjectID == project.ID {
				connectionFound = true
				connection = (*connections)[i]
				break
			}
		}

		if !connectionFound { // create one
			connection = &Connection{
				ProjectID: project.ID,
				Status: 0,
				Broker: fmt.Sprintf("tcp://%s:%d", project.BrokerAddress, project.BrokerPort),
				ClientID: project.BrokerClientID,
			}
			*connections = append(*connections, connection)
		}

		if connection.Status == 0 {
			// get widgets
			rows, err := db.Query("SELECT id, name FROM project_sections WHERE project_id = ?", project.ID)
			if err != nil {
				log.Fatal(err)
				return
			}

			var topics []string

			var projectSections []ProjectSection
			for rows.Next() {
				var projectSection ProjectSection

				err = rows.Scan(&projectSection.ID, &projectSection.Name)
				if err != nil {
					fmt.Fprintf(w, err.Error())
					continue
				}

				// Query the widgets that are related to this section
				widgetRows, err := db.Query("SELECT id, widget, config FROM project_widgets WHERE project_section_id = ?", projectSection.ID)
				if err != nil {
					log.Fatal(err)
					return
				}

				var projectWidgets []ProjectWidget
				for widgetRows.Next() {
					var projectWidget ProjectWidget

					err = widgetRows.Scan(&projectWidget.ID, &projectWidget.Widget, &projectWidget.Config)
					if err != nil {
						fmt.Fprintf(w, err.Error())
						continue
					}

					if projectWidget.Widget == "TEXT" {
						var textWidgetConfig TextWidgetConfig
						err = json.Unmarshal(projectWidget.Config, &textWidgetConfig)
						if err != nil {
							log.Fatal(err)
							return
						}

						topics = append(topics, textWidgetConfig.Topic)
					}

					projectWidgets = append(projectWidgets, projectWidget)
				}

				widgetRows.Close()

				projectSection.Widgets = projectWidgets

				projectSections = append(projectSections, projectSection)
			}

			rows.Close()

			go func() {
				err = connection.Connect()
				if err != nil {
					log.Fatal(err)
					return
				}

				for i := 0; i < len(topics); i++ {
					go connection.Subscribe(topics[i])
				}
			}()
		}

		http.Redirect(w, r, "/projects/" + slugParameter, http.StatusFound)
	}
}

func projectDisconnectHandler(db *sql.DB, connections *[]*Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slugParameter := r.PathValue("slug")

		var project Project
		err := db.QueryRow("SELECT id, name, slug, broker_address, broker_port, broker_protocol FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug, &project.BrokerAddress, &project.BrokerPort, &project.BrokerProtocol)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var connection *Connection

		// find connection
		connectionFound := false
		for i := 0; i < len(*connections); i++ {
			if (*connections)[i].ProjectID == project.ID {
				connectionFound = true
				connection = (*connections)[i]
				break
			}
		}

		if connectionFound && connection.Status == 1 {
			go connection.Disconnect()
		}

		http.Redirect(w, r, "/projects/" + slugParameter, http.StatusFound)
	}
}

func projectConnectionHandler(db *sql.DB, connections *[]*Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slugParameter := r.PathValue("slug")

		var project Project
		err := db.QueryRow("SELECT id, name, slug, broker_address, broker_port, broker_protocol FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug, &project.BrokerAddress, &project.BrokerPort, &project.BrokerProtocol)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		var connection *Connection

		// find connection
		connectionFound := false
		for i := 0; i < len(*connections); i++ {
			if (*connections)[i].ProjectID == project.ID {
				connectionFound = true
				connection = (*connections)[i]
				break
			}
		}

		if connectionFound {
			if connection.Status == 1 {
				fmt.Fprint(w, "online")
			} else {
				fmt.Fprint(w, "offline")
			}
			return
		}

		fmt.Fprintf(w, "offline")
	}
}

func projectDataHandler(db *sql.DB, connections *[]*Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slugParameter := r.PathValue("slug")

		var project Project
		err := db.QueryRow("SELECT id, name, slug, broker_address, broker_port, broker_protocol FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug, &project.BrokerAddress, &project.BrokerPort, &project.BrokerProtocol)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// find connection
		var connection *Connection
		connectionFound := false
		for i := 0; i < len(*connections); i++ {
			if (*connections)[i].ProjectID == project.ID {
				connectionFound = true
				connection = (*connections)[i]
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if !connectionFound {
			resp, err := json.Marshal(nil)
			if err != nil {
				log.Fatal(err)
				return
			}
			w.Write(resp)
			return
		}

		// get widgets and match topics
		rows, err := db.Query("SELECT id FROM project_sections WHERE project_id = ?", project.ID)
		if err != nil {
			log.Fatal(err)
			return
		}

		var data []WidgetData

		for rows.Next() {
			var projectSection ProjectSection

			err = rows.Scan(&projectSection.ID)
			if err != nil {
				fmt.Fprintf(w, err.Error())
				continue
			}

			// Query the widgets that are related to this section
			widgetRows, err := db.Query("SELECT id, widget, config FROM project_widgets WHERE project_section_id = ?", projectSection.ID)
			if err != nil {
				log.Fatal(err)
				return
			}

			for widgetRows.Next() {
				var projectWidget ProjectWidget

				err = widgetRows.Scan(&projectWidget.ID, &projectWidget.Widget, &projectWidget.Config)
				if err != nil {
					fmt.Fprintf(w, err.Error())
					continue
				}

				topic := ""

				// check Config is not empty

				if projectWidget.Widget == "TEXT" {
					var textWidgetConfig TextWidgetConfig
					err = json.Unmarshal(projectWidget.Config, &textWidgetConfig)
					if err != nil {
						log.Fatal(err)
						return
					}

					topic = textWidgetConfig.Topic
					if topic == "" {
						data = append(data, WidgetData{
							ID: projectWidget.ID,
							Data: nil,
						})
						continue
					}

					if connection.DataBuffer[topic] == nil || len(connection.DataBuffer[topic]) <= 0 {
						data = append(data, WidgetData{
							ID: projectWidget.ID,
							Data: nil,
						})
						continue
					}

					data = append(data, WidgetData{
						ID: projectWidget.ID,
						Data: string(connection.DataBuffer[topic][len(connection.DataBuffer[topic]) - 1]),
					})

					continue
				}

				if topic == "" {
					data = append(data, WidgetData{
						ID: projectWidget.ID,
						Data: nil,
					})
					continue
				}

				if connection.DataBuffer[topic] == nil || len(connection.DataBuffer[topic]) <= 0 {
					data = append(data, WidgetData{
						ID: projectWidget.ID,
						Data: nil,
					})
					continue
				}

				data = append(data, WidgetData{
					ID: projectWidget.ID,
					Data: connection.DataBuffer[topic][0],
				})
			}

			widgetRows.Close()
		}

		rows.Close()

		resp, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
			return
		}
		w.Write(resp)
	}
}

func projectSubmitValueHandler(db *sql.DB, connections *[]*Connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			fmt.Fprintf(w, "Only POST method is supported.")
			return
		}

		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ERROR: %v", err)
			return
		}

		slugParameter := r.PathValue("slug")
		id := r.FormValue("id")

		var project Project
		err := db.QueryRow("SELECT id, name, slug, broker_address, broker_port, broker_protocol FROM projects WHERE slug = ?", slugParameter).Scan(&project.ID, &project.Name, &project.Slug, &project.BrokerAddress, &project.BrokerPort, &project.BrokerProtocol)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// find connection
		var connection *Connection
		connectionFound := false
		for i := 0; i < len(*connections); i++ {
			if (*connections)[i].ProjectID == project.ID {
				connectionFound = true
				connection = (*connections)[i]
				break
			}
		}

		if !connectionFound {
			http.Redirect(w, r, "/projects/" + slugParameter, http.StatusFound)
			return
		}

		// find widget
		var projectWidget ProjectWidget
		err = db.QueryRow("SELECT id, widget, config FROM project_widgets WHERE id = ?", id).Scan(&projectWidget.ID, &projectWidget.Widget, &projectWidget.Config)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// parse config
		if projectWidget.Widget == "BUTTON" {
			var config ButtonWidgetConfig
			err = json.Unmarshal(projectWidget.Config, &config)
			if err != nil {
				log.Fatal(err)
				return
			}

			connection.SendMessage(config.Topic, string(config.Message))
		}

		http.Redirect(w, r, "/projects/" + slugParameter, http.StatusFound)
	}
}
