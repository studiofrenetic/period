package period

import (
	"fmt"
	"reflect"
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
// @param time.Time $date1
// @param time.Time $date2
// @return int
func compareDate(date1, date2 time.Time) int {
	if date1.UnixNano() > date2.UnixNano() {
		return 1
	}

	if date1.UnixNano() < date2.UnixNano() {
		return -1
	}

	return 0
}

// Create a Period object from a Year and a Week.
// @param int64 $year
// @param int64 $week index from 1 to 53
// @return self A new instance
func CreateFromWeek(year, week int) (p Period, err []error) {
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

func validateRange(value, min, max int) (int, []error) {
	if value < min || value > max {
		return value, []error{OutOfRangeError}
	}
	return value, nil
}

// Create a Period object from a Year and a Month.
// @param int $year
// @param int $month Month index from 1 to 12
// @return self A new instance
func CreateFromMonth(year, month int) (p Period, err []error) {
	if month, err = validateRange(month, 1, 12); err != nil {
		return p, err
	}

	p.Start, _ = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, month))
	p.End = p.Start.AddDate(0, 1, 0)

	return p, nil
}

// Create a Period object from a Year and a Quarter.
// @param int $year
// @param int $quarter Quarter Index from 1 to 4
// @return self A new instance
func CreateFromQuarter(year, quarter int) (p Period, err []error) {
	if quarter, err = validateRange(quarter, 1, 4); err != nil {
		return p, err
	}

	p.Start, _ = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, ((quarter-1)*3)+1))
	p.End = p.Start.AddDate(0, 3, 0)

	return p, nil
}

// Create a Period object from a Year and a Quarter.
// @param int $year
// @param int $semester Semester Index from 1 to 2
// @return self A new instance
func CreateFromSemester(year, semester int) (p Period, err []error) {
	if semester, err = validateRange(semester, 1, 2); err != nil {
		return p, err
	}

	p.Start, _ = time.Parse(YMDHIS, fmt.Sprintf("%d-%02d-01 00:00:00", year, ((semester-1)*6)+1))
	p.End = p.Start.AddDate(0, 6, 0)

	return p, nil
}

// Create a Period object from a Year
// @param int $year
// @return self A new instance
func CreateFromYear(year int) (p Period, err []error) {
	p.Start, _ = time.Parse(YMDHIS, fmt.Sprintf("%d-01-01 00:00:00", year))
	p.End = p.Start.AddDate(1, 0, 0)

	return p, nil
}

// Create a Period object from a starting point and an interval.
// @param string|\DateTimeInterface  $startDate start datepoint
// @param \Duration|float|string $duration  The duration. If an numeric is passed, it is
//                                              interpreted as the duration expressed in seconds.
//                                              If a string is passed, it must be parsable by
//                                              `Duration::createFromDateString`
// @return self A new instance
func CreateFromDuration(start time.Time, duration time.Duration) (p Period, err []error) {

	p.Start = start
	p.End = addDuration(p.Start, duration)

	return p, nil
}

func CreateFromDurationBeforeEnd(end time.Time, duration time.Duration) (p Period, err []error) {
	p.Start = subDuration(end, duration)
	p.End = end

	return p, nil
}

func (p *Period) Contains(index time.Time) bool {
	return (-1 < compareDate(index, p.Start)) &&
		(-1 == compareDate(index, p.End))
}

func addDuration(date time.Time, duration time.Duration) time.Time {
	return date.Add(duration)
}

func subDuration(date time.Time, duration time.Duration) time.Time {
	sub := time.Duration(-duration.Nanoseconds()) * time.Nanosecond
	date = date.Add(sub)

	return date
}

func (p *Period) StartingOn(start time.Time) {
	p.Start = start
}
func (p *Period) EndingOn(end time.Time) {
	p.End = end
}

func (p *Period) WithDuration(duration time.Duration) {
	p.End = addDuration(p.Start, duration)
}

func (p *Period) Add(duration time.Duration) {
	p.End = addDuration(p.End, duration)
}

func (p *Period) Sub(duration time.Duration) {
	p.End = subDuration(p.End, duration)
}

func (p *Period) Next() {
	clone := *p
	duration := clone.GetDurationInterval()
	p.Start = clone.End
	p.End = addDuration(clone.End, duration)
}

func (p *Period) Previous() {
	clone := *p
	duration := clone.GetDurationInterval()
	p.Start = subDuration(clone.Start, duration)
	p.End = clone.Start
}

func (p *Period) GetDurationInterval() time.Duration {
	end := p.End.UnixNano()
	start := p.Start.UnixNano()
	return time.Duration(end-start) * time.Nanosecond
}

func (p *Period) Overlaps(period Period) bool {
	if abuts, _ := p.Abuts(period); abuts {
		return false
	}

	return (-1 == compareDate(p.Start, period.End)) &&
		(1 == compareDate(p.End, period.Start))
}

// Tells whether two Period share the same datepoints.
// @param Period $period
// @return bool
func (p *Period) SameValueAs(period Period) bool {
	return 0 == compareDate(p.Start, period.Start) &&
		0 == compareDate(p.End, period.End)
}

// Tells whether a Period is entirely after the specified index
// @param Period|\DateTimeInterface $index
// @return bool
func (p *Period) IsAfter(period Period) bool {
	return -1 < compareDate(p.Start, period.Start)
}

// Tells whether a Period is entirely before the specified index
// @param Period|\DateTimeInterface $index
// @return bool
func (p *Period) IsBefore(period Period) bool {
	return 1 > compareDate(p.End, period.End)
}

// Tells whether two Period object abuts
// @param Period $period
// @return bool
func (p *Period) Abuts(period Period) (bool, int) {
	found, pos := in_array(0, []int{
		compareDate(p.Start, period.End),
		compareDate(p.End, period.Start),
	})

	return found, pos
}

func (p *Period) Diff(period Period) ([]Period, []error) {
	if p.Overlaps(period) == false {
		return nil, []error{ShouldOverlapsError}
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
// @param Period $period
// @param Period ...$periods one or more Period objects
// @return self A new instance
func (p *Period) Merge(periods ...Period) {
	// allPeriods := []Period{}
	for _, period := range periods {
		if 1 == compareDate(p.Start, period.Start) {
			p.StartingOn(period.Start)
		}

		if -1 == compareDate(p.End, period.End) {
			p.EndingOn(period.End)
		}
	}
}

// Computes the intersection between two Period objects.
// @param Period $period
// @return self A new instance
func (p *Period) Intersect(period Period) (Period, []error) {
	var newPeriod Period
	if abuts, _ := p.Abuts(period); abuts {
		return newPeriod, []error{BothShouldNotAbuts}
	}

	if newPeriod.Start = p.Start; 1 == compareDate(period.Start, p.Start) {
		newPeriod.Start = period.Start
	}

	if newPeriod.End = p.End; -1 == compareDate(period.End, p.End) {
		newPeriod.End = period.End
	}

	return newPeriod, nil
}

func (p *Period) Gap(period Period) (newPeriod Period) {

	if 1 == compareDate(period.Start, p.Start) {
		newPeriod.Start = p.End
		newPeriod.End = period.Start
		return
	}

	newPeriod.Start = period.End
	newPeriod.End = p.Start
	return
}

// Compares two Period objects according to their duration.
// @param Period $period
// @return int
func (p *Period) CompareDuration(period Period) int {
	return compareDate(p.End, period.End)
}

// Tells whether the current Period object duration
// is greater than the submitted one.
//
// @param Period $period
//
// @return bool
func (p *Period) DurationGreaterThan(period Period) bool {
	return 1 == p.CompareDuration(period)
}

// Tells whether the current Period object duration
// is less than the submitted one.
// @param Period $period
// @return bool
func (p *Period) DurationLessThan(period Period) bool {
	return -1 == p.CompareDuration(period)
}

// Tells whether the current Period object duration
// is equal to the submitted one
// @param Period $period
// @return bool
func (p *Period) SameDurationAs(period Period) bool {
	return 0 == p.CompareDuration(period)
}

// Create a Period object from a Year and a Quarter.
// @param Period $period
// @return \Duration
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

// http://codereview.stackexchange.com/questions/60074/in-array-in-go
func in_array(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}
