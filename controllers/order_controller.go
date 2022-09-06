package controllers

import (
	"context"
	"net/http"
	"time"
	"zerologix-coding/configs"
	"zerologix-coding/models"
	"zerologix-coding/responses"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var bidCollection *mongo.Collection = configs.GetCollection(configs.DB, "bids")

var askCollection *mongo.Collection = configs.GetCollection(configs.DB, "asks")

var tradeCollection *mongo.Collection = configs.GetCollection(configs.DB, "trades")

func Process(c *fiber.Ctx) error {

	//validate the request body
	order := new(models.Order)
	if err := c.BodyParser(&order); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})

	}
	id := primitive.NewObjectID()

	newOrder := models.Order{
		ID:        id,
		Side:      order.Side,
		Quantity:  order.Quantity,
		Price:     order.Price,
		Timestamp: primitive.NewDateTimeFromTime(time.Now()),
	}
	if order.Side.String() == "buy" {
		return ProcessLimitBuyOrder(c, newOrder)
	}
	return ProcessLimitSellOrder(c, newOrder)

}

func ProcessLimitBuyOrder(c *fiber.Ctx, order models.Order) error {

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"price", -1}, {"timestamp", -1}})
	cursor, err := askCollection.Find(context.TODO(), filter, opts)
	var results []models.Order
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	n := len(results)
	if n == 0 {
		return AddBuyOrder(c, order)
	}
	if n != 0 || results[len(results)-1].Price <= order.Price {
		for i := n - 1; i >= 0; i-- {
			sellOrder := results[i]
			if sellOrder.Price > order.Price {
				break
			}

			if sellOrder.Quantity >= order.Quantity {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer cancel()
				newTrade := models.Trade{

					TakerOrderID: order.ID,     // TakerOrderID
					MakerOrderID: sellOrder.ID, // Maker OrderID
					Quantity:     order.Quantity,
					Price:        sellOrder.Price,
					Timestamp:    order.Timestamp,
				}

				t, err := tradeCollection.InsertOne(ctx, newTrade)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
				}
				sellOrder.Quantity -= order.Quantity
				if sellOrder.Quantity == 0 {
					RemoveSellOrder(c, sellOrder)
				}
				return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": t}})

			}

			if sellOrder.Quantity < order.Quantity {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer cancel()
				newTrade := models.Trade{

					TakerOrderID: order.ID,     // TakerOrderID
					MakerOrderID: sellOrder.ID, // Maker OrderID
					Quantity:     sellOrder.Quantity,
					Price:        sellOrder.Price,
					Timestamp:    order.Timestamp,
				}

				_, err := tradeCollection.InsertOne(ctx, newTrade)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
				}

				order.Quantity -= sellOrder.Quantity
				// remove the sell Order as all quantities are filled by bid
				RemoveSellOrder(c, sellOrder)
				continue
			}
		}
	}

	return AddBuyOrder(c, order)

}

func ProcessLimitSellOrder(c *fiber.Ctx, order models.Order) error {

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"price", -1}, {"timestamp", -1}})
	cursor, err := bidCollection.Find(context.TODO(), filter, opts)
	var results []models.Order
	if err = cursor.All(context.TODO(), &results); err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	n := len(results)
	if n == 0 {
		return AddSellOrder(c, order)
	}
	if n != 0 || results[n-1].Price >= order.Price {
		for i := n - 1; i >= 0; i-- {
			buyOrder := results[i]
			if buyOrder.Price < order.Price {
				break
			}

			if buyOrder.Quantity >= order.Quantity {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer cancel()
				newTrade := models.Trade{

					TakerOrderID: order.ID,    // TakerOrderID
					MakerOrderID: buyOrder.ID, // Maker OrderID
					Quantity:     order.Quantity,
					Price:        buyOrder.Price,
					Timestamp:    order.Timestamp,
				}

				t, err := tradeCollection.InsertOne(ctx, newTrade)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
				}
				buyOrder.Quantity -= order.Quantity
				if buyOrder.Quantity == 0 {
					RemoveBuyOrder(c, buyOrder)

				}
				return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": t}})

			}

			if buyOrder.Quantity < order.Quantity {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				defer cancel()
				newTrade := models.Trade{

					TakerOrderID: order.ID,    // TakerOrderID
					MakerOrderID: buyOrder.ID, // Maker OrderID
					Quantity:     buyOrder.Quantity,
					Price:        buyOrder.Price,
					Timestamp:    order.Timestamp,
				}

				_, err := tradeCollection.InsertOne(ctx, newTrade)
				if err != nil {
					return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
				}

				order.Quantity -= buyOrder.Quantity
				// remove the sell Order as all quantities are filled by bid
				RemoveBuyOrder(c, buyOrder)
				continue
			}
		}
	}

	return AddSellOrder(c, order)

}

func AddBuyOrder(c *fiber.Ctx, order models.Order) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	result, err := bidCollection.InsertOne(ctx, order)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}

func AddSellOrder(c *fiber.Ctx, order models.Order) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	result, err := askCollection.InsertOne(ctx, order)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(responses.TradeResponse{Status: http.StatusInternalServerError, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}

func RemoveSellOrder(c *fiber.Ctx, order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	result, err := askCollection.DeleteOne(ctx, bson.M{"id": order.ID})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}
func RemoveBuyOrder(c *fiber.Ctx, order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	result, err := bidCollection.DeleteOne(ctx, bson.M{"id": order.ID})
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}
func UpdateBuyOrder(c *fiber.Ctx, order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	update := bson.D{{"quantity", order.Quantity}}
	result, err := bidCollection.UpdateOne(ctx, bson.M{"id": order.ID}, update)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}
func UpdateSellOrder(c *fiber.Ctx, order models.Order) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()
	update := bson.D{
		{"$set", bson.D{{"id", order.Quantity}}},
	}
	result, err := askCollection.UpdateOne(ctx, bson.M{"id": order.ID}, update)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(responses.TradeResponse{Status: http.StatusBadRequest, Message: "error", Data: &fiber.Map{"data": err.Error()}})
	}

	return c.Status(http.StatusCreated).JSON(responses.TradeResponse{Status: http.StatusCreated, Message: "success", Data: &fiber.Map{"data": result}})
}
