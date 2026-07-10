package service

import (
	"app/src/model"
	"app/src/utils"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AttendanceService interface {
	GetAttendances(c *fiber.Ctx) ([]model.Attendance, error)
	UpdateCorrection(c *fiber.Ctx, id string, status string, clockIn string, clockOut string, reason string) (*model.Attendance, error)
}

type attendanceService struct {
	Log *logrus.Logger
	DB  *gorm.DB
}

func NewAttendanceService(db *gorm.DB) AttendanceService {
	return &attendanceService{
		Log: utils.Log,
		DB:  db,
	}
}

func (s *attendanceService) GetAttendances(c *fiber.Ctx) ([]model.Attendance, error) {
	var records []model.Attendance

	query := s.DB.WithContext(c.Context()).Order("id asc")

	// Apply filter parameters if any (optional extension point)
	site := c.Query("site", "")
	if site != "" && site != "All Sites" {
		query = query.Where("site = ?", site)
	}

	dept := c.Query("department", "")
	if dept != "" && dept != "All Departments" {
		query = query.Where("department = ?", dept)
	}

	search := c.Query("search", "")
	if search != "" {
		query = query.Where("name LIKE ? OR role LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Find(&records).Error
	if err != nil {
		s.Log.Errorf("Failed to retrieve attendance records: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to retrieve attendance records")
	}

	return records, nil
}

func (s *attendanceService) UpdateCorrection(c *fiber.Ctx, id string, status string, clockIn string, clockOut string, reason string) (*model.Attendance, error) {
	var record model.Attendance

	err := s.DB.WithContext(c.Context()).First(&record, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusNotFound, "Attendance record not found")
		}
		s.Log.Errorf("Failed to find attendance record: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database error")
	}

	// Update fields
	record.Status = status
	record.ClockIn = clockIn
	record.ClockOut = clockOut
	record.CorrectionReason = reason
	record.UpdatedAt = time.Now()

	err = s.DB.WithContext(c.Context()).Save(&record).Error
	if err != nil {
		s.Log.Errorf("Failed to save attendance correction: %v", err)
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Failed to save correction")
	}

	return &record, nil
}
