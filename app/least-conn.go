package app

import (
	"net/http"
	"net/http/httputil"
	"sort"
	"sync/atomic"
)

func (lbApp *lbApp) leastConn(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path

	parseURL, err := lbApp.mostMatchingLocation(requestPath, w)
	if err != nil {
		lbApp.logger.Errorf("failed due to err: %v", err)
		return
	}

	upstream, found := lbApp.upstream[parseURL.Host]

	if !found || len(upstream.serverInfo) == 0 {
		reverseProxyHost := httputil.NewSingleHostReverseProxy(parseURL)
		reverseProxyHost.ServeHTTP(w, r)
		return
	}

	server := lbApp.nextServer(upstream)
	atomic.AddInt32(&server.currConnection, 1)
	defer atomic.AddInt32(&server.currConnection, -1)
	server.reverseProxy.ServeHTTP(w, r)

}

func (lbApp *lbApp) nextServer(upstream *upstreamInfo) *serverInfo {
	upstream.mux.Lock()
	defer upstream.mux.Unlock()
	sort.Slice(upstream.serverInfo, func(i, j int) bool {
		return upstream.serverInfo[i].currConnection < upstream.serverInfo[j].currConnection
	})
	return upstream.serverInfo[0]
}
