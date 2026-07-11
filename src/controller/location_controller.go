package controller

import (
	"app/src/model"
	"app/src/response"
	"app/src/service"
	"app/src/validation"
	"math"

	"github.com/gofiber/fiber/v2"
)

type LocationController struct {
	LocationService service.LocationService
}

func NewLocationController(locationService service.LocationService) *LocationController {
	return &LocationController{
		LocationService: locationService,
	}
}

func (l *LocationController) GetLocations(c *fiber.Ctx) error {
	query := &validation.QueryLocation{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 10),
		Search: c.Query("search", ""),
	}

	locations, totalResults, err := l.LocationService.GetLocations(c, query)
	if err != nil {
		return err
	}

	totalPages := math.Ceil(float64(totalResults) / float64(query.Limit))

	return c.Status(fiber.StatusOK).
		JSON(response.SuccessWithPaginate[model.Location]{
			Code:         fiber.StatusOK,
			Status:       "success",
			Message:      "Locations retrieved successfully",
			Results:      locations,
			Page:         query.Page,
			Limit:        query.Limit,
			TotalPages:   int64(totalPages),
			TotalResults: totalResults,
		})
}

func (l *LocationController) GetLocationByID(c *fiber.Ctx) error {
	id := c.Params("id")

	location, err := l.LocationService.GetLocationByID(c, id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.SuccessWithData[*model.Location]{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Location retrieved successfully",
			Data:    location,
		})
}

func (l *LocationController) CreateLocation(c *fiber.Ctx) error {
	req := new(validation.CreateLocation)

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	location, err := l.LocationService.CreateLocation(c, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).
		JSON(response.SuccessWithData[*model.Location]{
			Code:    fiber.StatusCreated,
			Status:  "success",
			Message: "Location created successfully",
			Data:    location,
		})
}

func (l *LocationController) UpdateLocation(c *fiber.Ctx) error {
	id := c.Params("id")
	req := new(validation.UpdateLocation)

	if err := c.BodyParser(req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	location, err := l.LocationService.UpdateLocation(c, id, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.SuccessWithData[*model.Location]{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Location updated successfully",
			Data:    location,
		})
}

func (l *LocationController) DeleteLocation(c *fiber.Ctx) error {
	id := c.Params("id")

	err := l.LocationService.DeleteLocation(c, id)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).
		JSON(response.Common{
			Code:    fiber.StatusOK,
			Status:  "success",
			Message: "Location deleted successfully",
		})
}
