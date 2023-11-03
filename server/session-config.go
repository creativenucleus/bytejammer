package server

type SessionConfig struct {
	Port int
	Name string
	Slug string
}

// JamSessionConfig should be enough to save to disk and restart a JamSession if it crashes
func getSessionConfig(js Session) SessionConfig {
	return SessionConfig{
		Port: js.port,
		Name: js.name,
		Slug: js.slug,
	}
}
