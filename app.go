package main

import (
	"zerologix-coding/configs"
	"zerologix-coding/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	//run database
	configs.ConnectDB()

	//routes
	routes.TradeRoute(app)

	app.Listen(":6000")
}
