package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	_ "github.com/jackc/pgx"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

var (
	dbuser     string
	dbpassword string
	dbname     string
)

func init() {
	flag.StringVar(&dbuser, "dbuser", "postgres", "PostgreSQL username")
	flag.StringVar(&dbpassword, "pass", "12345", "PostgreSQL password")
	flag.StringVar(&dbname, "dbname", "test", "PostgreSQL database name")
	flag.Parse()
}

func main() {
	db, err := sqlx.Connect("pgx", fmt.Sprintf("user=%s password=%s host=0.0.0.0 port=5432 dbname=%s sslmode=disable", dbuser, dbpassword, dbname))
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer db.Close()
	handler := NewHandler(NewBuildingStorage(db))
	router := chi.NewRouter()

	router.Use(render.SetContentType(render.ContentTypeJSON))
	router.Route("/building", func(r chi.Router) {
		r.Post("/", handler.AddBuilding)
		r.Get("/{id}", handler.GetCompaniesFromBuilding)
	})
	router.Get("/rubric/{id}", handler.GetCompaniesFromRubric)
	router.Get("/company/{id}", handler.GetCompany)

	fmt.Println("Server started:")
	http.ListenAndServe(":8080", router)
}
