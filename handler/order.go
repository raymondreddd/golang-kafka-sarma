package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/raymondreddd/golnag/model"
	"github.com/raymondreddd/golnag/repository/order"
)

type Order struct {
	Repo *order.RedisRepo
}

// CRUD methods on Order
func (o *Order) Create(w http.ResponseWriter, r *http.Request) {

	// data we expect from client
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items`
	}

	// decode the json data
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	// create a order of order type
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	// call insert, and if error 500
	if err := o.Repo.Create(r.Context(), order); err != nil {
		fmt.Println("failed to insert:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// encode to json the order before sending
	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("failed to encode marshal:", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// all is ok, send 201
	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) List(w http.ResponseWriter, r *http.Request) {
}

func (o *Order) GetById(w http.ResponseWriter, r *http.Request) {

}

func (o *Order) UpdateById(w http.ResponseWriter, r *http.Request) {

}

func (o *Order) DeleteById(w http.ResponseWriter, r *http.Request) {

}
