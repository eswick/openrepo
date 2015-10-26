package main

import(
	"net/http"
	"sync"
	"strings"
)

/* ======= Clean Path ======= */

type lazybuf struct {
	s   string
	buf []byte
	w   int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}
	return b.s[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.s) && b.s[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.s))
		copy(b.buf, b.s[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) string() string {
	if b.buf == nil {
		return b.s[:b.w]
	}
	return string(b.buf[:b.w])
}

/* From package "path"
	- Modified to keep single period (.)
*/
func CleanPath(path string) string {
	if path == "" {
		return "."
	}

	rooted := path[0] == '/'
	n := len(path)

	// Invariants:
	//	reading from path; r is index of next byte to process.
	//	writing to buf; w is index of next byte to write.
	//	dotdot is index in buf where .. must stop, either because
	//		it is the leading slash or it is a leading ../../.. prefix.
	out := lazybuf{s: path}
	r, dotdot := 0, 0
	if rooted {
		out.append('/')
		r, dotdot = 1, 1
	}

	for r < n {
		switch {
		case path[r] == '/':
			// empty path element
			r++
		/*case path[r] == '.' && (r+1 == n || path[r+1] == '/'):
			// . element (now disabled)
			r++*/
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || path[r+2] == '/'):
			// .. element: remove to last /
			r += 2
			switch {
			case out.w > dotdot:
				// can backtrack
				out.w--
				for out.w > dotdot && out.index(out.w) != '/' {
					out.w--
				}
			case !rooted:
				// cannot backtrack, but not rooted, so append .. element.
				if out.w > 0 {
					out.append('/')
				}
				out.append('.')
				out.append('.')
				dotdot = out.w
			}
		default:
			// real path element.
			// add slash if needed
			if rooted && out.w != 1 || !rooted && out.w != 0 {
				out.append('/')
			}
			// copy element
			for ; r < n && path[r] != '/'; r++ {
				out.append(path[r])
			}
		}
	}

	// Turn empty string into "."
	if out.w == 0 {
		return "."
	}

	return out.string()
}

/* ============== */









/* ======= RepoMux ======= */

/* Modified from net/http to allow single dot (.) in URL.
	- Otherwise, it sends a 301 redirect, which APT doesn't like.
*/

type RepoMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry
	hosts bool // whether any patterns contain hostnames
}

type muxEntry struct {
	explicit bool
	h        http.Handler
	pattern  string
}

// NewRepoMux allocates and returns a new RepoMux.
func NewRepoMux() *RepoMux { return &RepoMux{m: make(map[string]muxEntry)} }

var DefaultRepoMux = NewRepoMux()

func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := CleanPath(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
func (mux *RepoMux) match(path string) (h http.Handler, pattern string) {
	var n = 0
	for k, v := range mux.m {
		if !pathMatch(k, path) {
			continue
		}
		if h == nil || len(k) > n {
			n = len(k)
			h = v.h
			pattern = v.pattern
		}
	}
	return
}

func (mux *RepoMux) Handler(r *http.Request) (h http.Handler, pattern string) {
	if r.Method != "CONNECT" {
		if p := cleanPath(r.URL.Path); p != r.URL.Path {
			_, pattern = mux.handler(r.Host, p)
			url := *r.URL
			url.Path = p
			return http.RedirectHandler(url.String(), http.StatusMovedPermanently), pattern
		}
	}

	return mux.handler(r.Host, r.URL.Path)
}

// handler is the main implementation of Handler.
// The path is known to be in canonical form, except for CONNECT methods.
func (mux *RepoMux) handler(host, path string) (h http.Handler, pattern string) {
	mux.mu.RLock()
	defer mux.mu.RUnlock()

	// Host-specific pattern takes precedence over generic ones
	if mux.hosts {
		h, pattern = mux.match(host + path)
	}
	if h == nil {
		h, pattern = mux.match(path)
	}
	if h == nil {
		h, pattern = http.NotFoundHandler(), ""
	}
	return
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *RepoMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h, _ := mux.Handler(r)
	h.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *RepoMux) Handle(pattern string, handler http.Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if pattern == "" {
		panic("http: invalid pattern " + pattern)
	}
	if handler == nil {
		panic("http: nil handler")
	}
	if mux.m[pattern].explicit {
		panic("http: multiple registrations for " + pattern)
	}

	mux.m[pattern] = muxEntry{explicit: true, h: handler, pattern: pattern}

	if pattern[0] != '/' {
		mux.hosts = true
	}

	// Helpful behavior:
	// If pattern is /tree/, insert an implicit permanent redirect for /tree.
	// It can be overridden by an explicit registration.
	n := len(pattern)
	if n > 0 && pattern[n-1] == '/' && !mux.m[pattern[0:n-1]].explicit {
		// If pattern contains a host name, strip it and use remaining
		// path for redirect.
		path := pattern
		if pattern[0] != '/' {
			// In pattern, at least the last character is a '/', so
			// strings.Index can't be -1.
			path = pattern[strings.Index(pattern, "/"):]
		}
		mux.m[pattern[0:n-1]] = muxEntry{h: http.RedirectHandler(path, http.StatusMovedPermanently), pattern: pattern}
	}
}

// HandleFunc registers the handler function for the given pattern.
func (mux *RepoMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	mux.Handle(pattern, http.HandlerFunc(handler))
}

// Handle registers the handler for the given pattern
// in the DefaultRepoMux.
// The documentation for RepoMux explains how patterns are matched.
func Handle(pattern string, handler http.Handler) { DefaultRepoMux.Handle(pattern, handler) }

// HandleFunc registers the handler function for the given pattern
// in the DefaultRepoMux.
// The documentation for RepoMux explains how patterns are matched.
func HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	DefaultRepoMux.HandleFunc(pattern, handler)
}


/* ============== */
