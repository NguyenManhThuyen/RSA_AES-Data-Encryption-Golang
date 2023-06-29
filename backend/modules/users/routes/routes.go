package routes

import (
	"api/modules/users/controller"

	"github.com/gofiber/fiber/v2"
)

func InitUserRoutes(app *fiber.App) {
	user := app.Group("/users/user")
	user.Get("", controller.GetUser)
	user.Post("/insert", controller.InsertUser)
	user.Put("/update", controller.UpdateUser)
	user.Delete("/delete", controller.DeleteUser)
	user.Put("/change-password", controller.ChangePasswordUser)
}
