package task

import (
	"time"

	taskdomain "example.com/taskservice/internal/domain/task"
)

func CalculateRecurrenceDates(
	rt taskdomain.RecurrenceType,
	rcfg *taskdomain.RecurrenceConfig,
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
		dates = calcDaily(rcfg, from, to)
	case taskdomain.TypeMonthly:
		dates = calcMonthly(rcfg, from, to)
	case taskdomain.TypeEvenOdd:
		dates = calcEvenOdd(rcfg, from, to)
	}

	return dates
}

func calcDaily(rcfg *taskdomain.RecurrenceConfig, from, to time.Time) []time.Time {
	if rcfg.IntervalDays <= 0 {
		return nil
	}

	var dates []time.Time
	curDate := from

	for !curDate.After(to) {
		if !curDate.Before(from) && !curDate.After(to){
			dates = append(dates, curDate)
		}
		curDate = curDate.AddDate(0, 0, rcfg.IntervalDays)
	}

	return dates
}

func calcMonthly(rcfg *taskdomain.RecurrenceConfig, from, to time.Time) []time.Time {
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

			curDate := time.Date(year, month, day, from.Hour(), from.Minute(), 0, 0, time.UTC)
			
			if !curDate.Before(from) && !curDate.After(to) {
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

func calcEvenOdd(cfg *taskdomain.RecurrenceConfig, from, to time.Time) []time.Time {
	if cfg.Parity != "even" && cfg.Parity != "odd" {
		return nil
	}

	var dates []time.Time
	isEven := cfg.Parity == "even"
	curDate := from

	for !curDate.After(to) {
		curIsEven := curDate.Day()%2 == 0
		if isEven == curIsEven {
			dates = append(dates, curDate)
		}
		curDate = curDate.AddDate(0, 0, 1)
	}

	return dates
}