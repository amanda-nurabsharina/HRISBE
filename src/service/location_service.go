package service

import (
	"app/src/model"
	"app/src/utils"
	"app/src/validation"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type LocationService interface {
	GetLocations(c *fiber.Ctx, params *validation.QueryLocation) ([]model.Location, int64, error)
	GetLocationByID(c *fiber.Ctx, id string) (*model.Location, error)
	CreateLocation(c *fiber.Ctx, req *validation.CreateLocation) (*model.Location, error)
	UpdateLocation(c *fiber.Ctx, id string, req *validation.UpdateLocation) (*model.Location, error)
	DeleteLocation(c *fiber.Ctx, id string) error
}

type locationService struct {
	Log      *logrus.Logger
	DB       *gorm.DB
	Validate *validator.Validate
}

func NewLocationService(db *gorm.DB, validate *validator.Validate) LocationService {
	return &locationService{
		Log:      utils.Log,
		DB:       db,
		Validate: validate,
	}
}

func (s *locationService) GetLocations(c *fiber.Ctx, params *validation.QueryLocation) ([]model.Location, int64, error) {
	var locations []model.Location
	var totalResults int64

	if err := s.Validate.Struct(params); err != nil {
		return nil, 0, err
	}

	query := s.DB.WithContext(c.Context()).Model(&model.Location{}).Order("created_at desc")

	if params.Search != "" {
		query = query.Where("name LIKE ? OR department LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	err := query.Count(&totalResults).Error
	if err != nil {
		s.Log.Errorf("Failed to count locations: %v", err)
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Apply pagination
	if params.Page > 0 && params.Limit > 0 {
		offset := (params.Page - 1) * params.Limit
		query = query.Offset(offset).Limit(params.Limit)
	}

	err = query.Find(&locations).Error
	if err != nil {
		s.Log.Errorf("Failed to query locations: %v", err)
		return nil, 0, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	return locations, totalResults, nil
}

func (s *locationService) GetLocationByID(c *fiber.Ctx, id string) (*model.Location, error) {
	var location model.Location

	err := s.DB.WithContext(c.Context()).First(&location, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "Location not found")
		}
		s.Log.Errorf("Failed to query location by ID: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	return &location, nil
}

func (s *locationService) CreateLocation(c *fiber.Ctx, req *validation.CreateLocation) (*model.Location, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	location := model.Location{
		ID:         uuid.New().String(),
		Name:       req.Name,
		Department: req.Department,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
		Radius:     req.Radius,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.DB.WithContext(c.Context()).Create(&location).Error
	if err != nil {
		s.Log.Errorf("Failed to create location: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to create location")
	}

	return &location, nil
}

func (s *locationService) UpdateLocation(c *fiber.Ctx, id string, req *validation.UpdateLocation) (*model.Location, error) {
	if err := s.Validate.Struct(req); err != nil {
		return nil, err
	}

	location, err := s.GetLocationByID(c, id)
	if err != nil {
		return nil, err
	}

	location.Name = req.Name
	location.Department = req.Department
	location.Latitude = req.Latitude
	location.Longitude = req.Longitude
	location.Radius = req.Radius
	location.UpdatedAt = time.Now()

	err = s.DB.WithContext(c.Context()).Save(location).Error
	if err != nil {
		s.Log.Errorf("Failed to save location update: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to update location")
	}

	return location, nil
}

func (s *locationService) DeleteLocation(c *fiber.Ctx, id string) error {
	location, err := s.GetLocationByID(c, id)
	if err != nil {
		return err
	}

	err = s.DB.WithContext(c.Context()).Delete(location).Error
	if err != nil {
		s.Log.Errorf("Failed to delete location: %v", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to delete location")
	}

	return nil
}
