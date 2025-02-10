package object

import "time"

type Interval struct {
	Value int          `json:"value"`
	Unit  IntervalUnit `json:"unit"`
}

func NewInterval(value int, unit IntervalUnit) Interval {
	return Interval{
		Value: value,
		Unit:  unit,
	}
}

func NewIntervalFromDuration(duration time.Duration) Interval {
	value := int(duration.Minutes())
	unit := IntervalUnitMinutes

	if value >= 60 {
		value = int(duration.Hours())
		unit = IntervalUnitHours
	}
	if value >= 24 && unit == IntervalUnitHours {
		value = int(duration.Hours() / 24)
		unit = IntervalUnitDays
	}
	if value >= 7 && unit == IntervalUnitDays {
		value = int(duration.Hours() / (24 * 7))
		unit = IntervalUnitWeeks
	}
	if value >= 4 && unit == IntervalUnitWeeks {
		value = int(duration.Hours() / (24 * 30))
		unit = IntervalUnitMonths
	}

	return Interval{
		Value: value,
		Unit:  unit,
	}
}

func (t Interval) Duration() time.Duration {
	switch t.Unit {
	case IntervalUnitMinutes:
		return time.Duration(t.Value) * time.Minute
	case IntervalUnitHours:
		return time.Duration(t.Value) * time.Hour
	case IntervalUnitDays:
		return time.Duration(t.Value) * 24 * time.Hour
	case IntervalUnitWeeks:
		return time.Duration(t.Value) * 24 * 7 * time.Hour
	case IntervalUnitMonths:
		return time.Duration(t.Value) * 24 * 30 * time.Hour
	default:
		return 0
	}
}

func (t Interval) IsZero() bool {
	return t.Value == 0 && t.Unit == ""
}
