package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func GetEnv(key string) string {
	return os.Getenv(key)
}

func InitializePostgresDB() *sql.DB {
	dbuser := GetEnv("POSTGRES_USER")
	dbpassword := GetEnv("POSTGRES_PASSWORD")
	dbname := GetEnv("POSTGRES_DB")
	dbsslmode := GetEnv("POSTGRES_SSLMODE")
	host := GetEnv("POSTGRES_HOST")
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s", host, dbuser, dbpassword, dbname, dbsslmode)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	// defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to db")

	return db

}
