package routes

import (
	"zerologix-coding/controllers"

	"github.com/gofiber/fiber/v2"
)

func TradeRoute(app *fiber.App) {
	app.Post("/process", controllers.Process)
}
