package models

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Order Type
type Order struct {
	ID        primitive.ObjectID `json:"id"`
	Side      Side               `json:"side"`
	Quantity  float64            `json:"quantity"`
	Price     float64            `json:"price"`
	Timestamp primitive.DateTime `json:"timestamp"`
}

// Convert order to struct from json
func (order *Order) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, order)
}

// Convert order to json from order struct
func (order *Order) toJSON() []byte {
	str, _ := json.Marshal(order)
	return str
}
