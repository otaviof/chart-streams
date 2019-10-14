package chartstreams

type ChartStreamService struct {
	config        *Config
	gitRepository *Git
	index         map[string]interface{}
}

func NewChartStreamService(config *Config) *ChartStreamService {
	g := NewGit(config)

	return &ChartStreamService{
		config:        config,
		gitRepository: g,
	}
}

func (gs *ChartStreamService) Initialize() error {
	gs.index = make(map[string]interface{})

	return gs.gitRepository.Clone()
}

func (gs *ChartStreamService) GetHelmChart(name string, version string) error {
	return nil
}

func (gs *ChartStreamService) GetIndex() (map[string]interface{}, error) {
	return gs.index, nil
}
