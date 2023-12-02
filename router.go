package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/redds-be/rlinks/database"
)

type Database struct {
	db *sql.DB
}

func createRouter() *chi.Mux {
	// Creating the main router
	router := chi.NewRouter()

	return router
}

func configRouter(router *chi.Mux) *chi.Mux {
	// Add some configuration to the main router
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	return router
}

func subRouters(dbURL string) *chi.Mux {
	// Create a router for '/'
	rootRouter := chi.NewRouter()
	rootRouter.Get("/status", handlerReadiness)
	rootRouter.Get("/error", handlerErr)

	// Connect to the database and create the links table
	db := &Database{db: database.DbConnect(dbURL)}
	database.CreateLinksTable(db.db)

	rootRouter.Get("/garbage", db.handlerGarbage)

	rootRouter.Post("/", db.handlerCreateLink)
	rootRouter.Get("/*", db.handlerRedirectToUrl)

	return rootRouter
}

func mountRouter(router *chi.Mux, dbURL string) *chi.Mux {
	// Mount the sub router to the main router to /v1/
	rootRouter := subRouters(dbURL)
	router.Mount("/", rootRouter)

	return router
}

func getRouter(dbURL string) *chi.Mux {
	// Create the main router
	router := createRouter()

	// Add some configuration to the main router
	router = configRouter(router)

	// Mount other routers to the main one
	router = mountRouter(router, dbURL)

	return router
}
