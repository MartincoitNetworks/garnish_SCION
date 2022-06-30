package garnish

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

const Xcache = "X-Cache"
const XcacheHit = "HIT"
const XcacheMiss = "MISS"

/**
  Cache .-----request---->> Garnish (Cache GET requests) .-------request------>> original Server
      <<-----response-----.                             <<-------response------.
*/

type garnish struct {
	c     *cache
	proxy *httputil.ReverseProxy
}

// process: requested url --> garnish
func New(url url.URL) *garnish {
	director := func(req *http.Request) {
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
	}
	// director to modify the request
	reverseProxy := &httputil.ReverseProxy{Director: director}
	return &garnish{c: newCache(), proxy: reverseProxy}
}

// create a handler object--> once server get the request, it will handle the request and construct response
func (g *garnish) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	// only GET requests should be cached
	if r.Method != http.MethodGet {
		rw.Header().Set(Xcache, XcacheMiss)
		g.proxy.ServeHTTP(rw, r)
		return
	}

	u := r.URL.String()
	cached := g.c.get(u)
	// handle Xcache -response
	//if cached, return the cached data
	if cached != nil {
		rw.Header().Set(Xcache, XcacheHit)
		_, _ = rw.Write(cached)
		return
	}
	//instantiate the interface(struct responseWriter)
	proxyRW := &responseWriter{
		proxied: rw,
	}

	proxyRW.Header().Set(Xcache, XcacheMiss)
	g.proxy.ServeHTTP(proxyRW, r)

	cc := rw.Header().Get(cacheControl)
	toCache, duration := parseCacheControl(cc)
	//check if it needs cache
	if toCache {
		g.c.store(u, proxyRW.body, duration)
	}
}
