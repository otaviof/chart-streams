package chartstreams

// Config application central configuration instance.
type Config struct {
	RepoURL     string // git repository URL
	RelativeDir string // relative directory in repository
	CloneDepth  int    // number of commits to load from history
	ListenAddr  string // address to listen on (address:port)
	ForceClone  bool   // destroy the destination directory before cloning
}
