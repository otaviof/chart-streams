package chart

import "time"

// Builder provides a fluent interface to build charts.
type Builder interface {
	Build() (*Package, error)
	SetChartName(string) Builder
	SetChartPath(string) Builder
	SetCommitTime(time.Time) Builder
}
