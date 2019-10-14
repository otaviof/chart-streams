package chartstreams

type ChartStreamService struct {
	config        *Config
	gitRepository *Git
}

func NewChartStreamService(config *Config) *ChartStreamService {
	g := NewGit(config)

	return &ChartStreamService{
		config:        config,
		gitRepository: g,
	}
}

func (gs *ChartStreamService) Initialize() error {
	return gs.gitRepository.Clone()
}

func (gs *ChartStreamService) GetHelmChart(name string, version string) error {
	return nil
}

func (gs *ChartStreamService) GetIndex() (string, error) {
	return "", nil
}
