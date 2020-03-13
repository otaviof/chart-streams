package chart

// Builder provides a fluent interface to build charts.
type Builder interface {
	// Build produces a chart Package.
	Build() (*Package, error)
}
