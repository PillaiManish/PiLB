package common

import (
	"fmt"
	"net/http"
	"net/url"
	"pi.com/lb/constants"
	"pi.com/lb/model"
)

func OnStartUpValidation(config *model.LoadBalancerConfig) error {
	if config == nil {
		return fmt.Errorf("load-balancer config cannot be nil")
	}

	if config.Strategy == "" {
		config.Strategy = constants.DEFAULT_LOAD_BALANCER_STRATEGY
	}

	healthCheckPath := ""
	if config.HealthCheck != nil && IsStringNonEmpty(config.HealthCheck.Endpoint) {
		healthCheckPath = config.HealthCheck.Endpoint
	}

	for name, upstream := range config.Upstream {
		if len(upstream) == 0 {
			return fmt.Errorf("load-balancer upstream:%s configuration is empty", name)
		}

		for _, server := range upstream {
			err := ServerURLValidation(server.URL, healthCheckPath, server.MaxFails)

			if err != nil {
				return fmt.Errorf("upstream %s: %v", name, err)
			}
		}
	}

	for _, location := range config.Location {
		if location == nil {
			return fmt.Errorf("load-balancer path route cannot be nil")
		}

		if IsStringEmpty(location.Path) || IsStringEmpty(location.ProxyPass) {
			return fmt.Errorf("load-balancer path route/proxy pass must not be empty")
		}

		parseProxyPass, err := url.Parse(location.ProxyPass)
		if err != nil {
			return fmt.Errorf("error parsing proxy pass: %v url: %v", location.ProxyPass, err)
		}

		if IsStringEmpty(parseProxyPass.Host) {
			return fmt.Errorf("proxy pass host cannot be empty")
		}

		if _, found := config.Upstream[parseProxyPass.Hostname()]; !found {
			err = ServerURLValidation(parseProxyPass.String(), healthCheckPath, 0)
			if err != nil {
				return fmt.Errorf("path routes: %v", err)
			}
		}
	}

	return nil
}

func IsStringEmpty(str string) bool {
	return len(str) == 0
}

func IsStringNonEmpty(str string) bool {
	return !IsStringEmpty(str)
}

func ServerURLValidation(URL, healthCheckPath string, retryAttempt int) error {
	var err error
	if IsStringEmpty(URL) {
		return fmt.Errorf("URL cannot be empty")
	}

	if retryAttempt == 0 {
		retryAttempt = constants.DEFAULT_SERVER_RETRY_ATTEMPTS
	}

	if IsStringNonEmpty(healthCheckPath) {
		URL, err = url.JoinPath(URL, healthCheckPath)
		if err != nil {
			return err
		}
	}

	var httpResponse *http.Response
	for _ = range retryAttempt {
		httpResponse, err = http.Get(URL)
		if err != nil {
			continue
		}

		if httpResponse.StatusCode == http.StatusServiceUnavailable {
			err = fmt.Errorf("server %s unavailable", URL)
			continue
		}
	}
	return err
}
