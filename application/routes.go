package application

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/raymondreddd/golnag/handler"
	"github.com/raymondreddd/golnag/repository/order"
)

func (a *App) loadRoutes() {
	// define new router
	router := chi.NewRouter()

	// add a logger
	router.Use(middleware.Logger)

	// for /
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// for /order
	router.Route("/orders", a.loadOrderRoutes)

	a.router = router
}

func (a *App) loadOrderRoutes(router chi.Router) {

	// creates a new instance of order struct, since we take the address of it, orderHandler becomes pointer to order instance
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}

	/*
			A POST request to the webserver has one endpoint that receives a request and routes it
		to kafka topic via goroutine-1. Another goroutine-2 will receive this data from kafka topic
		and routes it to REDIS database
	*/
	router.Post("/", orderHandler.Create)

	router.Get("/", orderHandler.List)
	/*
			A GET request will retrieve the data from REDIS to Kafka via goroutine-3, and delivers
		the response
	*/
	router.Get("/{id}", orderHandler.GetById)
	router.Put("/", orderHandler.UpdateById)
	router.Post("/", orderHandler.DeleteById)

}
