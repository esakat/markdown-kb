package config

// ServeConfig holds configuration for the serve command.
type ServeConfig struct {
	RootDir string
	Port    int
	Open    bool
	Repo    RepoConfig
}

// IndexConfig holds configuration for the index command.
type IndexConfig struct {
	RootDir string
	Format  string
}
