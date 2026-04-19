package task

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type RecurrenceType string

const (
	TypeDaily   RecurrenceType = "daily"
	TypeMonthly RecurrenceType = "monthly"
	TypeYearly  RecurrenceType = "yearly"
	TypeEvenOdd RecurrenceType = "even_odd"
)

type RecurrenceConfig struct {
	IntervalDays int      `json:"interval_days,omitempty"`
	DaysOfMonth  []int    `json:"days_of_month,omitempty"`
	DatesOfYear  []string `json:"dates_of_year,omitempty"`
	Parity       string   `json:"parity,omitempty"`
}

func (rc *RecurrenceConfig) Valid(rt RecurrenceType) bool {
	if rc == nil {
		return false
	}

	switch rt {
	case TypeDaily, TypeMonthly, TypeYearly, TypeEvenOdd:
		return true
	default:
		return false
	}
}

func (c *RecurrenceConfig) Scan(value any) error {
	if value == nil {
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("expected []byte or string for JSONB, got %T", value)
	}

	return json.Unmarshal(bytes, c)
}

func (c *RecurrenceConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}
