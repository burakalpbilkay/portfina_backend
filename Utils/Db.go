package utils

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pass, name)
	fmt.Println("Connecting to Postgres with:")
	fmt.Println("Host:", host)
	fmt.Println("Port:", port)
	fmt.Println("User:", user)
	fmt.Println("Pass:", pass)
	fmt.Println("DB:", name)

	return sql.Open("postgres", dsn)
}
