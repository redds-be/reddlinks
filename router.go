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

func subRouters(dbURL string) *chi.Mux {
	// Create a router for '/'
	rootRouter := chi.NewRouter()
	rootRouter.Get("/status", handlerReadiness)
	rootRouter.Get("/error", handlerErr)

	// Connect to the database and create a config
	apiCfg := dbConnect(dbURL)

	rootRouter.Get("/garbage", apiCfg.handlerGarbage)

	rootRouter.Post("/", apiCfg.handlerCreateLink)
	rootRouter.Get("/*", apiCfg.getURL(apiCfg.handlerGetLink))

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
