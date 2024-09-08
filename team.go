package main

import (
	"database/sql"
	"log"
)

type Team struct {
	ID int
	Name string
}

type TeamUser struct {
	ID int
	TeamID int
	UserID int
	Role string
}

func createTeamsTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS teams(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create teams table", err.Error())
		panic(err)
	}

	createTeamUsersTable(db)
}

func createTeamUsersTable(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS team_users(
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		team_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		role TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalln("Unable to create team_users table", err.Error())
		panic(err)
	}
}

func CreateTeam(db *sql.DB, name string) (int, error) {
	stmt, err := db.Prepare("INSERT INTO teams(name) VALUES(?)")
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(name)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func CreateTeamUser(db *sql.DB, teamID int, userID int, role string) (int, error) {
	stmt, err := db.Prepare("INSERT INTO team_users(team_id, user_id, role) VALUES(?,?,?)")
	if err != nil {
		return 0, err
	}

	res, err := stmt.Exec(teamID, userID, role)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func GetTeamUserByUserID(db *sql.DB, userID int) (*TeamUser, error) {
	var teamUser TeamUser
	err := db.QueryRow("SELECT * FROM team_users WHERE user_id = ?", userID).Scan(&teamUser.ID, &teamUser.TeamID, &teamUser.UserID, &teamUser.Role)
	if err != nil {
		return nil, err
	}

	return &teamUser, nil
}
