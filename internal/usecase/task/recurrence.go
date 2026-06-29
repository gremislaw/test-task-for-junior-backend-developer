package task

import (
	"strconv"
	"strings"
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

	var dates []time.Time

	switch rt {
	case taskdomain.TypeDaily:
		dates = calcDaily(rcfg, dueDate, from, to)
	case taskdomain.TypeMonthly:
		dates = calcMonthly(rcfg, dueDate, from, to)
	case taskdomain.TypeYearly:
		dates = calcYearly(rcfg, dueDate, from, to)
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
		if !curDate.Before(from) && curDate != dueDate {
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
	year, month := dueDate.Year(), dueDate.Month()
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

func calcYearly(rcfg *taskdomain.RecurrenceConfig, dueDate time.Time, from, to time.Time) []time.Time {
	if len(rcfg.DatesOfYear) == 0 {
		return nil
	}

	var dates []time.Time
	startYear := from.Year()
	endYear := to.Year()

	for year := startYear; year <= endYear; year++ {
		for _, md := range rcfg.DatesOfYear {
			parts := strings.Split(md, "-")
			if len(parts) != 2 {
				continue
			}
			m, errM := strconv.Atoi(parts[0])
			d, errD := strconv.Atoi(parts[1])
			if errM != nil || errD != nil || m < 1 || m > 12 || d < 1 || d > 31 {
				continue
			}

			candidate := time.Date(year, time.Month(m), d,
				dueDate.Hour(), dueDate.Minute(), dueDate.Second(), dueDate.Nanosecond(), time.UTC)

			if candidate.Month() != time.Month(m) {
				continue
			}

			if !candidate.Before(from) && !candidate.After(to) {
				dates = append(dates, candidate)
			}
		}
	}
	return dates
}

func calcEvenOdd(rcfg *taskdomain.RecurrenceConfig, dueDate time.Time, from, to time.Time) []time.Time {
	if rcfg.Parity != "even" && rcfg.Parity != "odd" {
		return nil
	}

	var dates []time.Time
	isEven := rcfg.Parity == "even"
	curDate := time.Date(from.Year(), from.Month(), from.Day(),
		dueDate.Hour(), dueDate.Minute(), dueDate.Second(), dueDate.Nanosecond(), time.UTC)

	if (curDate.Day()%2 == 0) != isEven {
		curDate = curDate.AddDate(0, 0, 1)
	}

	for !curDate.After(to) {
		dates = append(dates, curDate)
		curDate = curDate.AddDate(0, 0, 2)
	}

	return dates
}

func FilterExcluded(dates []time.Time, excluded []time.Time) []time.Time {
	if len(excluded) == 0 {
		return dates
	}
	exclSet := make(map[string]struct{}, len(excluded))
	for _, t := range excluded {
		exclSet[t.Format(time.RFC3339)] = struct{}{}
	}
	var out []time.Time
	for _, d := range dates {
		if _, skip := exclSet[d.Format(time.RFC3339)]; !skip {
			out = append(out, d)
		}
	}
	return out
}
