package router

import (
	"app/src/controller"
	m "app/src/middleware"
	"app/src/service"

	"github.com/gofiber/fiber/v2"
)

func AttendanceRoutes(v1 fiber.Router, attendanceService service.AttendanceService, userService service.UserService) {
	attendanceController := controller.NewAttendanceController(attendanceService)

	attendance := v1.Group("/attendance")

	attendance.Get("/", m.Auth(userService), attendanceController.GetAttendances)
	attendance.Put("/:id/correction", m.Auth(userService), attendanceController.PostCorrection)
}
