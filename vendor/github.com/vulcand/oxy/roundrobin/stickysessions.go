// package stickysession is a mixin for load balancers that implements layer 7 (http cookie) session affinity
package roundrobin

import (
	"net/http"
	"net/url"
	"time"
)

type StickySession struct {
	cookiename string
}

func NewStickySession(c string) *StickySession {
	return &StickySession{c}
}

// GetBackend returns the backend URL stored in the sticky cookie, iff the backend is still in the valid list of servers.
func (s *StickySession) GetBackend(req *http.Request, servers []*url.URL) (*url.URL, bool, error) {
	cookie, err := req.Cookie(s.cookiename)
	switch err {
	case nil:
	case http.ErrNoCookie:
		return nil, false, nil
	default:
		return nil, false, err
	}

	s_url, err := url.Parse(cookie.Value)
	if err != nil {
		return nil, false, err
	}

	if s.isBackendAlive(s_url, servers) {
		return s_url, true, nil
	} else {
		return nil, false, nil
	}
}

func (s *StickySession) StickBackend(backend *url.URL, w *http.ResponseWriter) {
	// set cookie to expire one day in the future
	// set cookie path to the root path so it is sent back to the server for all requests
	now := time.Now()
	tomorrow := now.AddDate(0, 1, 0)
	c := &http.Cookie{Name: s.cookiename, Value: backend.String(), Path: "/", Expires: tomorrow}
	http.SetCookie(*w, c)
	return
}

func (s *StickySession) isBackendAlive(needle *url.URL, haystack []*url.URL) bool {
	if len(haystack) == 0 {
		return false
	}

	for _, s := range haystack {
		if sameURL(needle, s) {
			return true
		}
	}
	return false
}
