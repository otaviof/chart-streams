package chartstreams

type GitService struct {
	config        *Config
	gitRepository *Git
}

func NewGitService(config *Config) *GitService {
	g := NewGit(config)

	return &GitService{
		config:        config,
		gitRepository: g,
	}
}

func (gs *GitService) Initialize() error {
	return gs.gitRepository.Clone()
}

func (gs *GitService) GetHelmChart(name string, version string) error {
	return nil
}

func (gs *GitService) GetIndex() (string, error) {
	return "", nil
}
