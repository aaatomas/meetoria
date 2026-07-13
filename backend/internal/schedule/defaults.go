package schedule

import (
	"time"

	"github.com/google/uuid"

	commonmodel "github.com/meetoria/meetoria/backend/internal/common/model"
)

func ParseClockTime(value string) (time.Time, error) {
	t, err := time.Parse("15:04", value)
	if err != nil {
		return time.Time{}, err
	}
	return time.Date(0, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC), nil
}

func FormatClockTime(t time.Time) string {
	return t.Format("15:04")
}

// DefaultBranchWorkingHours returns Mon–Fri 09:00–17:00 for a branch.
func DefaultBranchWorkingHours(orgID, branchID uuid.UUID) []WorkingHours {
	start, _ := NewClockTimeFromString("09:00")
	end, _ := NewClockTimeFromString("17:00")
	var hours []WorkingHours
	for day := 1; day <= 5; day++ {
		hours = append(hours, WorkingHours{
			OrganizationScoped: commonmodel.OrganizationScoped{OrganizationID: orgID},
			BranchID:           &branchID,
			DayOfWeek:          day,
			StartTime:          start,
			EndTime:            end,
			IsActive:           true,
		})
	}
	return hours
}

func DefaultHoursForDay(dayOfWeek int) []WorkingHours {
	if dayOfWeek < 1 || dayOfWeek > 5 {
		return nil
	}
	start, _ := NewClockTimeFromString("09:00")
	end, _ := NewClockTimeFromString("17:00")
	return []WorkingHours{{
		DayOfWeek: dayOfWeek,
		StartTime: start,
		EndTime:   end,
		IsActive:  true,
	}}
}
