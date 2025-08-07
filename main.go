package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"example.com/app/internal/app"
	"example.com/app/internal/routes"
)

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "go backend server port")
	flag.Parse()

	app, err := app.NewApplication()
	if err != nil {
		panic(err)
	}

	defer app.DB.Close()

	app.Logger.Println("we are running our app on port", port)

	r := routes.SetupRoutes(app)
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err = server.ListenAndServe()

	if err != nil {
		app.Logger.Fatal(err)
	}
}
