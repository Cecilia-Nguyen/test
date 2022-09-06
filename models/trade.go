package models

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Trade struct {
	TakerOrderID primitive.ObjectID `json:"taker_order_id"`
	MakerOrderID primitive.ObjectID `json:"maker_order_id"`
	Quantity     float64            `json:"quantity"`
	Price        float64            `json:"price"`
	Timestamp    primitive.DateTime `json:"timestamp"`
}

// struct to json
func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
