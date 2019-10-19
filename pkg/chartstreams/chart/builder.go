package chart

import "time"

// Builder provides a fluent interface to build charts.
type Builder interface {
	// Build produces a chart Package.
	Build() (*Package, error)
	// SetChartName configures the name of the chart.
	SetChartName(string) Builder
	// SetChartPath configures the path of the chart.
	SetChartPath(string) Builder
	// SetCommitTime configures the chart's creation time.
	SetCommitTime(time.Time) Builder
}
