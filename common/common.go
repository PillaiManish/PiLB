package common

import (
	"fmt"
	"net/http"
	"net/url"
	"pi.com/lb/model"
)

func OnStartUpValidation(config *model.LoadBalancerConfig) error {
	if config == nil {
		return fmt.Errorf("load-balancer config cannot be nil")
	}

	if config.Strategy == "" {
		config.Strategy = model.DEFAULT_LOAD_BALANCER_STRATEGY
	}

	for name, upstream := range config.Upstream {
		if len(upstream) == 0 {
			return fmt.Errorf("load-balancer upstream:%s configuration is empty", name)
		}

		for _, server := range upstream {
			err := serverURLValidation(server.URL)

			if err != nil {
				return fmt.Errorf("upstream %s: %v", name, err)
			}
		}
	}

	for _, pathRoute := range config.PathRoutes {
		if pathRoute == nil {
			return fmt.Errorf("load-balancer path route cannot be nil")
		}

		if isStringEmpty(pathRoute.Path) || isStringEmpty(pathRoute.ProxyPass) {
			return fmt.Errorf("load-balancer path route/proxy pass must not be empty")
		}

		parseProxyPass, err := url.Parse(pathRoute.ProxyPass)
		if err != nil {
			return fmt.Errorf("error parsing proxy pass: %v url: %v", pathRoute.ProxyPass, err)
		}

		if isStringEmpty(parseProxyPass.Host) {
			return fmt.Errorf("proxy pass host cannot be empty")
		}

		if _, found := config.Upstream[parseProxyPass.Hostname()]; !found {
			err = serverURLValidation(parseProxyPass.String())
			if err != nil {
				return fmt.Errorf("path routes: %v", err)
			}
		}
	}

	return nil
}

func isStringEmpty(str string) bool {
	return len(str) == 0
}

func isStringNonEmpty(str string) bool {
	return !isStringEmpty(str)
}

func serverURLValidation(URL string) error {
	if isStringEmpty(URL) {
		return fmt.Errorf("URL cannot be empty")
	}

	httpResponse, err := http.Get(URL)
	if err != nil {
		return fmt.Errorf("error during http GET %v: %v", URL, err)
	}

	if httpResponse.StatusCode == http.StatusServiceUnavailable {
		return fmt.Errorf("server %s unavailable", URL)
	}
	return nil
}
