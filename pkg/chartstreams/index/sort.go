package index

type byChartNameAndVersion []ChartNameVersion

func (b byChartNameAndVersion) Len() int {
	return len(b)
}

func (b byChartNameAndVersion) Less(i, j int) bool {
	if b[i].name < b[j].name {
		return true
	}
	if b[i].name == b[j].name {
		return b[i].version < b[j].version
	}
	return false
}

func (b byChartNameAndVersion) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}
