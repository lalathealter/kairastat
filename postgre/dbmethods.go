package postgre

import (
	"database/sql"
	"log"
)

type wrapper struct {
	db *sql.DB
}

func (wr wrapper) GetUserFor(clientIP string) int {
	var userID int
	dbrow := wr.db.QueryRow(SelectUserIDQuery, clientIP)
	dbrow.Scan(&userID)
	if userID == 0 {
		dbrow := wr.db.QueryRow(CreateUserQuery, clientIP)
		if err := dbrow.Err(); err != nil {
			log.Panicln(err)
		}
		dbrow.Scan(&userID)
	}

	return userID
}

func (wr wrapper) SetUserAuthorized(userID int, isAuthorized bool) {
	_, err := wr.db.Exec(UpdateUserAuthorizedQuery, isAuthorized, userID)
	if err != nil {
		log.Panicln(err)
	} 
}

func (wr wrapper) SaveEvent(eventName string, userID int) {
	_, err := wr.db.Exec(CreateEventQuery, eventName, userID)
	if err != nil {
		log.Panicln(err)
	}
}

const (
	SelectUserIDQuery = `
		SELECT user_id 
		FROM kairastat.users
		WHERE ip_address=$1
	;`

	SelectUserIPQuery = `
		SELECT ip_address
		FROM kairastat.users
		WHERE user_id=$1
	;`

	CreateUserQuery = `
		INSERT INTO kairastat.users
		(ip_address)
		VALUES
		($1)
		RETURNING user_id
	;`

	UpdateUserAuthorizedQuery = `
		UPDATE kairastat.users
		SET authorized = $1
		WHERE user_id=$2
	;`

	SelectEventIDQuery = `
		SELECT event_id 
		FROM kairastat.events
		WHERE event_name=$1
	;`

	SelectEventQuery = `
		SELECT * 
		FROM kairastat.events
		WHERE event_name=$1
	;`

	CreateEventQuery = `
		INSERT INTO kairastat.events AS evs
		(event_name, author_id)
		VALUES 
		($1, $2)
		ON CONFLICT (event_name, author_id)
		DO UPDATE
			SET endorsements_count = evs.endorsements_count + 1
	;`
)
