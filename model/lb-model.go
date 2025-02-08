package model

type (
	Config struct {
		Port           string          `json:"port,omitempty" yaml:"port,omitempty"`
		ServerList     []string        `json:"serverList,omitempty" yaml:"serverList,omitempty"`
		HealthCheckCfg *HealthCheckCfg `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	}

	HealthCheckCfg struct {
		Endpoint          string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
		IntervalInSeconds int    `json:"intervalInSeconds,omitempty" yaml:"intervalInSeconds,omitempty"`
	}
)
