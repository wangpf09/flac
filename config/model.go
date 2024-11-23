package config

type Config struct {
	FlacInfo    FlacInfo `yaml:"flac"`
	WorkerCount int      `yaml:"worker-count"`
	SavePath    string   `yaml:"save-path"`
}

type FlacInfo struct {
	Baseurl    string   `yaml:"baseurl"`
	SearchApi  string   `yaml:"search-api"`
	UrlQq      string   `yaml:"url-qq"`
	Quality    []string `yaml:"quality"`
	Keywords   []string `yaml:"keywords"`
	PageSize   int      `yaml:"page-size"`
	UnlockCode string   `yaml:"unlock-code"`
}
