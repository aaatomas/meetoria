package service

import (
	"context"

	"github.com/google/uuid"

	apperrors "github.com/meetoria/meetoria/backend/internal/common/errors"
	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
	"github.com/meetoria/meetoria/backend/internal/schedule"
	schedulerepo "github.com/meetoria/meetoria/backend/internal/schedule/repository"
)

type Service interface {
	GetBranchSchedule(ctx context.Context, orgID, branchID uuid.UUID) ([]schedule.DaySchedule, error)
	SetBranchSchedule(ctx context.Context, orgID uuid.UUID, req schedule.SetWorkingHoursRequest) error
	SeedDefaultHours(ctx context.Context, orgID, branchID uuid.UUID) error
}

type scheduleService struct {
	repo schedulerepo.Repository
}

func NewService(repo schedulerepo.Repository) Service {
	return &scheduleService{repo: repo}
}

func (s *scheduleService) GetBranchSchedule(ctx context.Context, orgID, branchID uuid.UUID) ([]schedule.DaySchedule, error) {
	hours, err := s.repo.ListBranchWorkingHours(ctx, orgID, branchID)
	if err != nil {
		return nil, apperrors.Internal("failed to get working hours", err)
	}
	if len(hours) == 0 {
		hours = schedule.DefaultBranchWorkingHours(orgID, branchID)
	}
	return groupHoursByDay(hours), nil
}

func (s *scheduleService) SetBranchSchedule(ctx context.Context, orgID uuid.UUID, req schedule.SetWorkingHoursRequest) error {
	if req.BranchID == nil {
		return apperrors.Validation("branch_id is required")
	}

	var records []schedule.WorkingHours
	for _, day := range req.Schedule {
		for _, slot := range day.Slots {
			start, err := schedule.NewClockTimeFromString(slot.StartTime)
			if err != nil {
				return apperrors.Validation("invalid start_time format, use HH:MM")
			}
			end, err := schedule.NewClockTimeFromString(slot.EndTime)
			if err != nil {
				return apperrors.Validation("invalid end_time format, use HH:MM")
			}
			if !end.Time.After(start.Time) {
				return apperrors.Validation("end_time must be after start_time")
			}
			branchID := *req.BranchID
			records = append(records, schedule.WorkingHours{
				OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
				BranchID:           &branchID,
				DayOfWeek:          day.DayOfWeek,
				StartTime:          start,
				EndTime:            end,
				IsActive:           true,
			})
		}
	}

	return s.repo.SetWorkingHours(ctx, orgID, req.BranchID, req.EmployeeID, records)
}

func (s *scheduleService) SeedDefaultHours(ctx context.Context, orgID, branchID uuid.UUID) error {
	return s.repo.SetWorkingHours(ctx, orgID, &branchID, nil, schedule.DefaultBranchWorkingHours(orgID, branchID))
}

func groupHoursByDay(hours []schedule.WorkingHours) []schedule.DaySchedule {
	byDay := make(map[int][]schedule.TimeRange)
	for _, wh := range hours {
		if !wh.IsActive {
			continue
		}
		byDay[wh.DayOfWeek] = append(byDay[wh.DayOfWeek], schedule.TimeRange{
			StartTime: schedule.FormatClockTime(wh.StartTime.Time),
			EndTime:   schedule.FormatClockTime(wh.EndTime.Time),
		})
	}

	result := make([]schedule.DaySchedule, 0, len(byDay))
	for day := 0; day <= 6; day++ {
		slots, ok := byDay[day]
		if !ok {
			continue
		}
		result = append(result, schedule.DaySchedule{
			DayOfWeek: day,
			Slots:     slots,
		})
	}
	return result
}
