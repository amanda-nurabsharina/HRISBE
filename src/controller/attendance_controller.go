package controller

import (
	"app/src/service"

	"github.com/gofiber/fiber/v2"
)

type AttendanceController struct {
	AttendanceService service.AttendanceService
}

func NewAttendanceController(attendanceService service.AttendanceService) *AttendanceController {
	return &AttendanceController{
		AttendanceService: attendanceService,
	}
}

type CorrectionRequest struct {
	Status   string `json:"status"`
	ClockIn  string `json:"clockIn"`
	ClockOut string `json:"clockOut"`
	Reason   string `json:"reason"`
}

func (ctrl *AttendanceController) GetAttendances(c *fiber.Ctx) error {
	records, err := ctrl.AttendanceService.GetAttendances(c)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  "success",
		"results": records,
	})
}

func (ctrl *AttendanceController) PostCorrection(c *fiber.Ctx) error {
	id := c.Params("id")
	var req CorrectionRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request payload")
	}

	record, err := ctrl.AttendanceService.UpdateCorrection(c, id, req.Status, req.ClockIn, req.ClockOut, req.Reason)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"status":  "success",
		"message": "Koreksi absensi berhasil disimpan",
		"result":  record,
	})
}
