package period

import (
	"fmt"
	"time"
)

const (
	UTS               = "1136239445"          //  unixtimestamp
	ICSFORMAT         = "20060102T150405Z"    //ics date time format
	YMDHIS            = "2006-01-02 15:04:05" // Y-m-d H:i:S time format
	ICSFORMATWHOLEDAY = "20060102"            // ics date format ( describes a whole day)
)

var StartWeek time.Weekday = time.Monday
var Timezone *time.Location = time.UTC

type Period struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Compare DateTimeInterface objects including microseconds
func compareDate(date1, date2 time.Time) int {
	if date1.After(date2) {
		return 1
	}

	if date1.Before(date2) {
		return -1
	}

	return 0
}

func CreateFromDay(year int, month int, day int) (p Period, err error) {
	if month, err = validateRange(month, 1, 12); err != nil {
		return p, err
	}

	if day, err = validateRange(day, 1, 31); err != nil {
		return p, err
	}

	p.Start = time.Date(year, time.Month(month), day, 0, 0, 0, 0, Timezone)
	p.End = p.Start.AddDate(0, 0, 1)

	return p, nil
}

// Create a Period object from a Year and a Week.
func CreateFromWeek(year, week int) (p Period, err error) {
	if week, err = validateRange(week, 1, 53); err != nil {
		return p, err
	}

	p.Start = time.Date(year, 0, 0, 0, 0, 0, 0, Timezone)
	isoYear, isoWeek := p.Start.ISOWeek()
	for p.Start.Weekday() != time.Monday { // iterate back to Monday (ISO_8601 first day of week = Monday)
		p.Start = p.Start.AddDate(0, 0, -1)
		isoYear, isoWeek = p.Start.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		p.Start = p.Start.AddDate(0, 0, 1)
		isoYear, isoWeek = p.Start.ISOWeek()
		// fmt.Printf("isoYear: %s, year: %s\n", isoYear, year)
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		p.Start = p.Start.AddDate(0, 0, 1)
		isoYear, isoWeek = p.Start.ISOWeek()
	}
	for p.Start.Weekday() != StartWeek { // iterate back to StartWeek
		var diff int
		diff = int(StartWeek) - int(time.Monday)
		p.Start = p.Start.AddDate(0, 0, diff)
		isoYear, isoWeek = p.Start.ISOWeek()
	}

	// add 1 week
	p.End = p.Start.AddDate(0, 0, 7)

	return p, nil
}

func validateRange(value, min, max int) (int, error) {
	if value < min || value > max {
		return value, OutOfRangeError
	}
	return value, nil
}

// Create a Period object from a Year and a Month.
func CreateFromMonth(year, month int) (p Period, err error) {
	if month, err = validateRange(month, 1, 12); err != nil {
		return p, err
	}

	p.Start, err = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, month))
	if err != nil {
		return p, err
	}

	p.End = p.Start.AddDate(0, 1, 0)

	return p, nil
}

// Create a Period object from a Year and a Quarter.
func CreateFromQuarter(year, quarter int) (p Period, err error) {
	if quarter, err = validateRange(quarter, 1, 4); err != nil {
		return p, err
	}

	p.Start, err = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, ((quarter-1)*3)+1))
	if err != nil {
		return p, err
	}

	p.End = p.Start.AddDate(0, 3, 0)

	return p, nil
}

// Create a Period object from a Year and a Quarter.
func CreateFromSemester(year, semester int) (p Period, err error) {
	if semester, err = validateRange(semester, 1, 2); err != nil {
		return p, err
	}

	p.Start, err = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, ((semester-1)*6)+1))
	if err != nil {
		return p, err
	}

	p.End = p.Start.AddDate(0, 6, 0)

	return p, nil
}

// Create a Period object from a Year
func CreateFromYear(year int) (p Period, err error) {
	p.Start, err = time.Parse(YMDHIS, fmt.Sprintf("%d-01-01 00:00:00", year))
	if err != nil {
		return p, err
	}

	p.End = p.Start.AddDate(1, 0, 0)

	return p, nil
}

// Create a Period object from a starting point and an interval.
func CreateFromDuration(start time.Time, duration time.Duration) (p Period) {

	p.Start = start
	p.End = p.Start.Add(duration)

	return p
}

func CreateFromDurationBeforeEnd(end time.Time, duration time.Duration) (p Period) {
	p.Start = end.Add(-1 * duration)
	p.End = end

	return p
}

func (p *Period) Contains(index time.Time) bool {
	return (-1 < compareDate(index, p.Start)) &&
		(-1 == compareDate(index, p.End))
}

func (p *Period) WithDuration(duration time.Duration) {
	p.End = p.Start.Add(duration)
}

func (p *Period) Add(duration time.Duration) {
	p.End = p.End.Add(duration)
}

func (p *Period) Sub(duration time.Duration) {
	p.End = p.End.Add(-1 * duration)
}

func (p *Period) Next() {
	clone := *p
	duration := clone.GetDurationInterval()
	p.Start = clone.End
	p.End = clone.End.Add(duration)
}

func (p *Period) Previous() {
	clone := *p
	duration := clone.GetDurationInterval()
	p.Start = clone.Start.Add(-1 * duration)
	p.End = clone.Start
}

func (p *Period) GetDurationInterval() time.Duration {
	return p.End.Sub(p.Start)
}

func (p *Period) Overlaps(period Period) bool {
	if abuts, _ := p.Abuts(period); abuts {
		return false
	}

	return (-1 == compareDate(p.Start, period.End)) &&
		(1 == compareDate(p.End, period.Start))
}

// Tells whether a Period is entirely after the specified index
func (p *Period) After(period Period) bool {
	return p.Start.After(period.End)
}

// Tells whether a Period is entirely before the specified index
func (p *Period) Before(period Period) bool {
	return p.End.Before(period.Start)
}

// Tells whether two Period object abuts
func (p *Period) Abuts(period Period) (bool, int) {
	if p.Start.Equal(period.End) || p.End.Equal(period.Start) {
		var pos int = -1

		if compareDate(p.Start, period.End) == 0 {
			pos = 0
		} else if compareDate(p.End, period.Start) == 0 {
			pos = 1
		}

		return true, pos
	}

	return false, -1
}

func (p *Period) Diff(period Period) ([]Period, error) {
	if p.Overlaps(period) == false {
		return nil, ShouldOverlapsError
	}

	var res = []Period{}
	var period1 Period = createFromDatepoints(p.Start, period.Start)
	var period2 Period = createFromDatepoints(p.End, period.End)

	if compareDate(period1.Start, period1.End) != 0 {
		res = append(res, period1)
	}

	if compareDate(period2.Start, period2.End) != 0 {
		res = append(res, period2)
	}

	return res, nil
}

// Merges one or more Period objects to return a new Period object.
// The resultant object englobes the largest duration possible.
func (p *Period) Merge(periods ...Period) {
	// allPeriods := []Period{}
	for _, period := range periods {
		if 1 == compareDate(p.Start, period.Start) {
			p.Start = period.Start
		}

		if -1 == compareDate(p.End, period.End) {
			p.End = period.End
		}
	}
}

// Computes the intersection between two Period objects.
func (p *Period) Intersect(period Period) (Period, error) {
	var newPeriod Period
	if abuts, _ := p.Abuts(period); abuts {
		return newPeriod, BothShouldNotAbuts
	}

	if newPeriod.Start = p.Start; 1 == compareDate(period.Start, p.Start) {
		newPeriod.Start = period.Start
	}

	if newPeriod.End = p.End; -1 == compareDate(period.End, p.End) {
		newPeriod.End = period.End
	}

	return newPeriod, nil
}

func (p *Period) Gap(period Period) Period {
	var newPeriod Period
	if 1 == compareDate(period.Start, p.Start) {
		newPeriod.Start = p.End
		newPeriod.End = period.Start
		return newPeriod
	}

	newPeriod.Start = period.End
	newPeriod.End = p.Start
	return newPeriod
}

// Compares two Period objects according to their duration.
func (p *Period) CompareDuration(period Period) int {
	return compareDate(p.End, period.End)
}

// Tells whether the current Period object duration
// is greater than the submitted one.
func (p *Period) DurationGreaterThan(period Period) bool {
	return 1 == p.CompareDuration(period)
}

// Tells whether the current Period object duration
// is less than the submitted one.
func (p *Period) DurationLessThan(period Period) bool {
	return -1 == p.CompareDuration(period)
}

// Tells whether the current Period object duration
// is equal to the submitted one
func (p *Period) SameDurationAs(period Period) bool {
	return p.GetDurationInterval() == period.GetDurationInterval()
}

// Create a Period object from a Year and a Quarter.
func (p *Period) DurationDiff(period Period) time.Duration {
	return time.Duration(p.TimestampDurationDiff(period)) * time.Nanosecond
}

func (p *Period) TimestampDurationDiff(period Period) int64 {
	return p.GetDurationInterval().Nanoseconds() - period.GetDurationInterval().Nanoseconds()
}

func createFromDatepoints(date1, date2 time.Time) Period {
	if 1 == compareDate(date1, date2) {
		return Period{date2, date1}
	}

	return Period{date1, date2}
}
