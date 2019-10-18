package chart

import "time"

// ChartBuilder provides a fluent interface to build charts.
type ChartBuilder interface {
	Build() (*Package, error)
	SetChartName(string) ChartBuilder
	SetChartPath(string) ChartBuilder
	SetCommitTime(time.Time) ChartBuilder
}
