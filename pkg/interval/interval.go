package interval

import "time"

type Interval struct {
	Value int  `json:"value"`
	Unit  Unit `json:"unit"`
}

func New(value int, unit Unit) Interval {
	return Interval{
		Value: value,
		Unit:  unit,
	}
}

func FromDuration(duration time.Duration) Interval {
	value := int(duration.Minutes())
	unit := UnitMinutes

	if value >= 60 {
		value = int(duration.Hours())
		unit = UnitHours
	}
	if value >= 24 && unit == UnitHours {
		value = int(duration.Hours() / 24)
		unit = UnitDays
	}
	if value >= 7 && unit == UnitDays {
		value = int(duration.Hours() / (24 * 7))
		unit = UnitWeeks
	}
	if value >= 4 && unit == UnitWeeks {
		value = int(duration.Hours() / (24 * 30))
		unit = UnitMonths
	}

	return Interval{
		Value: value,
		Unit:  unit,
	}
}

func (t Interval) Duration() time.Duration {
	switch t.Unit {
	case UnitMinutes:
		return time.Duration(t.Value) * time.Minute
	case UnitHours:
		return time.Duration(t.Value) * time.Hour
	case UnitDays:
		return time.Duration(t.Value) * 24 * time.Hour
	case UnitWeeks:
		return time.Duration(t.Value) * 24 * 7 * time.Hour
	case UnitMonths:
		return time.Duration(t.Value) * 24 * 30 * time.Hour
	default:
		return 0
	}
}

func (t Interval) IsZero() bool {
	return t.Value == 0 && t.Unit == ""
}
