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
