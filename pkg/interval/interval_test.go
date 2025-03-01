package interval_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/pkg/interval"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	i := interval.New(5, interval.UnitMinutes)
	assert.Equal(t, 5, i.Value)
	assert.Equal(t, interval.UnitMinutes, i.Unit)
}

func TestFromDuration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		duration time.Duration
		expected interval.Interval
	}{
		{"Minutes", time.Minute * 30, interval.New(30, interval.UnitMinutes)},
		{"Hours", time.Hour * 2, interval.New(2, interval.UnitHours)},
		{"Days", time.Hour * 48, interval.New(2, interval.UnitDays)},
		{"Weeks", time.Hour * 24 * 14, interval.New(2, interval.UnitWeeks)},
		{"Months", time.Hour * 24 * 30 * 2, interval.New(2, interval.UnitMonths)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			i := interval.FromDuration(tc.duration)
			assert.Equal(t, tc.expected, i)
		})
	}
}

func TestInterval_Duration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval interval.Interval
		expected time.Duration
	}{
		{"Minutes", interval.New(30, interval.UnitMinutes), time.Minute * 30},
		{"Hours", interval.New(2, interval.UnitHours), time.Hour * 2},
		{"Days", interval.New(2, interval.UnitDays), time.Hour * 48},
		{"Weeks", interval.New(2, interval.UnitWeeks), time.Hour * 24 * 14},
		{"Months", interval.New(2, interval.UnitMonths), time.Hour * 24 * 30 * 2},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			duration := tc.interval.Duration()
			assert.Equal(t, tc.expected, duration)
		})
	}
}

func TestInterval_IsZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval interval.Interval
		expected bool
	}{
		{"ZeroInterval", interval.New(0, ""), true},
		{"NonZeroInterval", interval.New(5, interval.UnitMinutes), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			isZero := tc.interval.IsZero()
			assert.Equal(t, tc.expected, isZero)
		})
	}
}
