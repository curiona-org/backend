package object_test

import (
	"testing"
	"time"

	"github.com/curiona-org/backend/internal/domain/object"
	"github.com/stretchr/testify/assert"
)

func TestNewInterval(t *testing.T) {
	t.Parallel()
	interval := object.NewInterval(5, object.IntervalUnitMinutes)
	assert.Equal(t, 5, interval.Value)
	assert.Equal(t, object.IntervalUnitMinutes, interval.Unit)
}

func TestNewIntervalFromDuration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		duration time.Duration
		expected object.Interval
	}{
		{"Minutes", time.Minute * 30, object.NewInterval(30, object.IntervalUnitMinutes)},
		{"Hours", time.Hour * 2, object.NewInterval(2, object.IntervalUnitHours)},
		{"Days", time.Hour * 48, object.NewInterval(2, object.IntervalUnitDays)},
		{"Weeks", time.Hour * 24 * 14, object.NewInterval(2, object.IntervalUnitWeeks)},
		{"Months", time.Hour * 24 * 30 * 2, object.NewInterval(2, object.IntervalUnitMonths)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			interval := object.NewIntervalFromDuration(tc.duration)
			assert.Equal(t, tc.expected, interval)
		})
	}
}

func TestInterval_Duration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		interval object.Interval
		expected time.Duration
	}{
		{"Minutes", object.NewInterval(30, object.IntervalUnitMinutes), time.Minute * 30},
		{"Hours", object.NewInterval(2, object.IntervalUnitHours), time.Hour * 2},
		{"Days", object.NewInterval(2, object.IntervalUnitDays), time.Hour * 48},
		{"Weeks", object.NewInterval(2, object.IntervalUnitWeeks), time.Hour * 24 * 14},
		{"Months", object.NewInterval(2, object.IntervalUnitMonths), time.Hour * 24 * 30 * 2},
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
		interval object.Interval
		expected bool
	}{
		{"ZeroInterval", object.NewInterval(0, ""), true},
		{"NonZeroInterval", object.NewInterval(5, object.IntervalUnitMinutes), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			isZero := tc.interval.IsZero()
			assert.Equal(t, tc.expected, isZero)
		})
	}
}
