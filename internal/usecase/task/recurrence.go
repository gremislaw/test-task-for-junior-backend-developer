package task

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func CalculateRecurrenceDates(
	rt taskdomain.RecurrenceType,
	rcfg *taskdomain.RecurrenceConfig,
	dueDate time.Time,
	from, to time.Time,
) []time.Time {
	if rcfg == nil {
		return nil
	}

	if to.Sub(from) > 365*24*time.Hour {
		return nil
	}

	var dates []time.Time

	switch rt {
	case taskdomain.TypeDaily:
		dates = calcDaily(rcfg, dueDate, from, to)
	case taskdomain.TypeMonthly, taskdomain.TypeSpecific:
		dates = calcMonthly(rcfg, dueDate, from, to)
	case taskdomain.TypeEvenOdd:
		dates = calcEvenOdd(rcfg, dueDate, from, to)
	}

	return dates
}

func calcDaily(rcfg *taskdomain.RecurrenceConfig, dueDate time.Time, from, to time.Time) []time.Time {
	if rcfg.IntervalDays <= 0 {
		return nil
	}

	var dates []time.Time
	curDate := dueDate

	for !curDate.After(to) {
		if !curDate.Before(from) && !curDate.After(to) && curDate != dueDate {
			dates = append(dates, curDate)
		}
		curDate = curDate.AddDate(0, 0, rcfg.IntervalDays)
	}

	return dates
}

func calcMonthly(rcfg *taskdomain.RecurrenceConfig, dueDate time.Time, from, to time.Time) []time.Time {
	if len(rcfg.DaysOfMonth) == 0 {
		return nil
	}

	var dates []time.Time
	year, month := from.Year(), from.Month()
	endYear, endMonth := to.Year(), to.Month()

	for year < endYear || (year == endYear && month <= endMonth) {
		for _, day := range rcfg.DaysOfMonth {
			if day > 31 {
				return nil
			}

			curDate := time.Date(year, month, day, dueDate.Hour(), dueDate.Minute(), 0, 0, time.UTC)

			if !curDate.Before(from) && !curDate.After(to) && curDate != dueDate {
				dates = append(dates, curDate)
			}
		}
		month++
		if month > 12 {
			month = 1
			year++
		}
	}

	return dates
}

func calcEvenOdd(cfg *taskdomain.RecurrenceConfig, dueDate time.Time, from, to time.Time) []time.Time {
	if cfg.Parity != "even" && cfg.Parity != "odd" {
		return nil
	}

	var dates []time.Time
	isEven := cfg.Parity == "even"
	curDate := dueDate

	for !curDate.After(to) && curDate != dueDate {
		curIsEven := curDate.Day()%2 == 0
		if isEven == curIsEven {
			dates = append(dates, curDate)
		}
		curDate = curDate.AddDate(0, 0, 1)
	}

	return dates
}
