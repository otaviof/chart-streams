package config

type Config struct {
	RepoURL    string // git repository URL
	ListenAddr string // address to listen on
	Depth      int    // number of commits to load from history
}
