package modules

import (
	_ "api/docs"
	// "os/user"
	// authenRoute "api/modules/authen/routes"
	"api/controller"
	classRoute "api/modules/classes/routes"
	studentRoute "api/modules/student/routes"
	tblRoute "api/modules/tbl/routes"
	teacherRoute "api/modules/teachers/routes"
	"api/modules/users/controller"
	userRoute "api/modules/users/routes"

	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"
	"api/datatest"
	// "api/middleware"
)

func InitRoutes(app *fiber.App) {
	// authenRoute.InitAuthenRoutes(app)
	studentRoute.InitStudentsRoutes(app)
	classRoute.InitClassRoutes(app)
	teacherRoute.InitTeacherRoutes(app)
	tblRoute.InitTblRoutes(app)
	userRoute.InitUserRoutes(app)
	
	// Wellcome
	app.Get("/", controllerr.Wellcome)

	// Document
	app.Get("/document/*", swagger.HandlerDefault)
	app.Get("/test", controllerr.Test)

	// Create data test
	admin := app.Group("/admin")
	admin.Post("/create-data-test", datatest.Createuser)

	app.Post("/login", controller.Login)
}
