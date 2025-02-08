package model

type (
	Config struct {
		Port           string          `json:"port,omitempty" yaml:"port,omitempty"`
		ServerList     []string        `json:"serverList,omitempty" yaml:"serverList,omitempty"`
		HealthCheckCfg *HealthCheckCfg `json:"healthCheck,omitempty" yaml:"healthCheck,omitempty"`
	}

	HealthCheckCfg struct {
		Endpoints         string `json:"endpoints,omitempty" yaml:"endpoints,omitempty"`
		IntervalInSeconds int    `json:"intervalInSeconds,omitempty" yaml:"intervalInSeconds,omitempty"`
	}
)
