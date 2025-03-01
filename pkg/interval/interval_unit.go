package interval

type Unit string

const (
	UnitMinutes Unit = "minutes"
	UnitHours   Unit = "hours"
	UnitDays    Unit = "days"
	UnitWeeks   Unit = "weeks"
	UnitMonths  Unit = "months"
)

func (u Unit) String() string {
	return string(u)
}
