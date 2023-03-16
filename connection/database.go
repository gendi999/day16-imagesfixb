package connection

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4"
)

var Conn *pgx.Conn

func DatabaseConnect() {

	var err error
	databaseUrl := "postgres://postgres:12345gendi@localhost:5432/personal-web45"
	Conn, err = pgx.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unabel apakah bisa to database: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("succes connect to database")
}
