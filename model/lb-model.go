package model

const (
	ROUND_ROBIN                    LoadBalanceStrategy = "ROUND_ROBIN"
	LEAST_CONN                     LoadBalanceStrategy = "LEAST_CONN"
	RANDOM                         LoadBalanceStrategy = "RANDOM"
	DEFAULT_LOAD_BALANCER_STRATEGY                     = ROUND_ROBIN
)

type (
	LoadBalanceStrategy string

	LoadBalancerConfig struct {
		Listen        int                  `yaml:"listen,omitempty"`
		ServerName    string               `yaml:"server_name,omitempty"` // allowed client host
		Upstream      map[string][]*Server `yaml:"upstream,omitempty"`
		PathRoutes    []*PathRoute         `yaml:"path_routes,omitempty"`
		Strategy      LoadBalanceStrategy  `yaml:"strategy,omitempty"`
		StickySession bool                 `yaml:"sticky_session,omitempty"`
		SSLConfig     *SSL                 `yaml:"ssl,omitempty"`
		HealthCheck   *Health              `yaml:"health_check,omitempty"`
		RateLimit     *Rate                `yaml:"rate_limit,omitempty"`
		Timeouts      *Timeout             `yaml:"timeouts,omitempty"`
	}

	PathRoute struct {
		Path      string `yaml:"path,omitempty"`
		ProxyPass string `yaml:"proxyPass,omitempty"` // by default right now HTTP, will support for me in future
	}

	Server struct {
		URL         string `yaml:"url,omitempty"`
		Weight      int    `yaml:"weight,omitempty"`
		MaxFails    int    `yaml:"max_fails,omitempty"`
		FailTimeout int    `yaml:"fail_timeout,omitempty"`
	}

	SSL struct {
		Enabled        bool   `yaml:"enabled,omitempty"`
		Certificate    string `yaml:"certificate,omitempty"`
		CertificateKey string `yaml:"certificate_key,omitempty"`
	}

	Health struct {
		Endpoint          string `yaml:"endpoint,omitempty"`
		IntervalInSeconds int    `yaml:"interval,omitempty"`
	}

	Rate struct {
		Enabled bool   `yaml:"enabled,omitempty"`
		Zone    string `yaml:"zone,omitempty"`
		Rate    string `yaml:"rate,omitempty"` // e.g., "10r/s"
		Burst   int    `yaml:"burst,omitempty"`
	}

	Timeout struct {
		Connect int `yaml:"connect,omitempty"`
		Send    int `yaml:"send,omitempty"`
		Read    int `yaml:"read,omitempty"`
	}
)
