package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

var currentDB *sql.DB

func init() {
	currentDB = connect()
}

func GetDB() *sql.DB {
	return currentDB
}

func getEnv(key string) string {
	val, found := os.LookupEnv(key)
	if !found {
		log.Fatalln("An env var is missing: ", key)
	}
	return val
}

func getpsqlconn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("dbhost"), getEnv("dbport"), getEnv("dbuser"), getEnv("dbpassword"), getEnv("dbname"),
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


