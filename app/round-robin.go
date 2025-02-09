package app

import (
	"net/http"
	"net/http/httputil"
)

func (lbApp *lbApp) roundRobinImpl(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	parseURL, err := lbApp.mostMatchingLocation(requestPath, w)
	if err != nil {
		lbApp.logger.Errorf("failed due to err: %v", err)
		return
	}

	servers, found := lbApp.upstream[parseURL.Host]

	if !found || len(servers.serverInfo) == 0 {
		reverseProxyHost := httputil.NewSingleHostReverseProxy(parseURL)
		reverseProxyHost.ServeHTTP(w, r)
		return
	}

	servers.mux.Lock()
	server := servers.serverInfo[servers.currServerId%len(servers.serverInfo)]
	servers.currServerId = (servers.currServerId + 1) % len(servers.serverInfo)
	lbApp.logger.Debugf("id: %v: %s", servers.currServerId, parseURL.Host)
	servers.mux.Unlock()
	server.reverseProxy.ServeHTTP(w, r)
}
