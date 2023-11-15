package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

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

func subRouters() *chi.Mux {
	// Create a router for '/'
	RootRouter := chi.NewRouter()
	RootRouter.Get("/ok", handlerReadiness)
	RootRouter.Get("/err", handlerErr)
	RootRouter.Get("/", handlerRedirect)

	return RootRouter
}

func mountRouter(router *chi.Mux) *chi.Mux {
	// Mount the sub router to the main router to /v1/
	RootRouter := subRouters()
	router.Mount("/", RootRouter)

	return router
}

func getRouter() *chi.Mux {
	// Create the main router
	router := createRouter()

	// Add some configuration to the main router
	router = configRouter(router)

	// Mount other routers to the main one
	router = mountRouter(router)

	return router
}
