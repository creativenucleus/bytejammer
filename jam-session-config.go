package main

type JamSessionConfig struct {
	Port int
	Name string
	Slug string
}

// JamSessionConfig should be enough to save to disk and restart a JamSession if it crashes
func getJamSessionConfig(js JamSession) JamSessionConfig {
	return JamSessionConfig{
		Port: js.port,
		Name: js.name,
		Slug: js.slug,
	}
}
