package router

import (
	"app/src/controller"
	m "app/src/middleware"
	"app/src/service"

	"github.com/gofiber/fiber/v2"
)

func LocationRoutes(v1 fiber.Router, locationService service.LocationService, userService service.UserService) {
	locationController := controller.NewLocationController(locationService)

	location := v1.Group("/locations")

	location.Get("/", m.Auth(userService), locationController.GetLocations)
	location.Get("/:id", m.Auth(userService), locationController.GetLocationByID)
	location.Post("/", m.Auth(userService), locationController.CreateLocation)
	location.Put("/:id", m.Auth(userService), locationController.UpdateLocation)
	location.Delete("/:id", m.Auth(userService), locationController.DeleteLocation)
}
