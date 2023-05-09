package postgre


import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// "github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var currentDB *sql.DB
func init() {
	// _ = godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatalln(err)
	// }

	currentDB = connect()
	presetTables(currentDB)
}

func GetWrapper() Wrapper {
	return Wrapper{currentDB}
}

func GetEnv(key string) string {
	val, found := os.LookupEnv(key)
	if !found {
		log.Fatalln("An env var is missing: ", key)
	}
	return val
}

func getpsqlconn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		GetEnv("dbhost"), GetEnv("dbport"), GetEnv("dbuser"), GetEnv("dbpassword"), GetEnv("dbname"),
	)
}
func connect() *sql.DB {
	db, err := sql.Open("postgres", getpsqlconn())
	if err != nil {
		log.Panicln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Panicln(err)
	}

	fmt.Println("Database connected")
	return db
}

func presetTables(db *sql.DB) {
	initcommands := [...]string{
		`
			CREATE SCHEMA IF NOT EXISTS kairastat
		;`,

		`
			CREATE TABLE IF NOT EXISTS kairastat.users (
				user_id SERIAL PRIMARY KEY NOT NULL,
				ip_address INET NOT NULL,
				authorized BOOL NOT NULL DEFAULT FALSE
			)
		;`,

	
		`
			CREATE TABLE IF NOT EXISTS kairastat.events (
				event_id SERIAL PRIMARY KEY NOT NULL,
				event_name VARCHAR(128) NOT NULL,
				endorsements_count INT NOT NULL DEFAULT 1,
				created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), 
				author_id SERIAL NOT NULL,
				CONSTRAINT uniqueness 
					UNIQUE (event_name, author_id),
				CONSTRAINT author_id
					FOREIGN KEY(author_id)
						REFERENCES kairastat.users(user_id)
						ON DELETE CASCADE
			)
		;`,
	}

	for _, comm := range initcommands {
		_, err := db.Exec(comm)
		if err != nil {
			log.Panicln(err.Error())
		}
	}
}
