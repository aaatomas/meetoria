package schedule

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type ClockTime struct {
	time.Time
}

func (t ClockTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(fmt.Sprintf(`"%s"`, t.Time.Format("15:04"))), nil
}

func (t *ClockTime) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case string:
		for _, layout := range []string{"15:04:05", "15:04"} {
			parsed, err := time.Parse(layout, v)
			if err == nil {
				t.Time = time.Date(0, 1, 1, parsed.Hour(), parsed.Minute(), parsed.Second(), 0, time.UTC)
				return nil
			}
		}
		return fmt.Errorf("invalid time string: %s", v)
	case time.Time:
		t.Time = time.Date(0, 1, 1, v.Hour(), v.Minute(), v.Second(), 0, time.UTC)
		return nil
	default:
		return fmt.Errorf("unsupported time value: %T", value)
	}
}

func (t ClockTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.Time.Format("15:04:05"), nil
}

func NewClockTimeFromString(value string) (ClockTime, error) {
	parsed, err := ParseClockTime(value)
	if err != nil {
		return ClockTime{}, err
	}
	return ClockTime{Time: parsed}, nil
}
