package light

import (
	"context"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/Akagi201/utilgo/chain"
	"github.com/julienschmidt/httprouter"
)

// Handler is the http request multiplexer backed by httprouter.Router.
type Handler struct {
	prefix string            // prefix for all paths
	routes map[string]*route // map of pattern to route for subtrees
	router *httprouter.Router
	chain  chain.Chain
}

type route struct {
	Method  string
	Handler http.HandlerFunc
}

// Config is the Handler configuration.
type Config struct {
	// Prefix is the prefix for all paths in the handler.
	// Empty value is allowed and defaults to "/".
	Prefix string

	// Enables automatic redirection if the current route can't be matched but a
	// handler for the path with (without) the trailing slash exists.
	// For example if /foo/ is requested but a route only exists for /foo, the
	// client is redirected to /foo with http status code 301 for GET requests
	// and 307 for all other request methods.
	RedirectTrailingSlash bool

	// If enabled, the router tries to fix the current request path, if no
	// handle is registered for it.
	// First superfluous path elements like ../ or // are removed.
	// Afterwards the router does a case-insensitive lookup of the cleaned path.
	// If a handle can be found for this route, the router makes a redirection
	// to the corrected path with status code 301 for GET requests and 307 for
	// all other request methods.
	// For example /FOO and /..//Foo could be redirected to /foo.
	// RedirectTrailingSlash is independent of this option.
	RedirectFixedPath bool

	// If enabled, the router checks if another method is allowed for the
	// current route, if the current request can not be routed.
	// If this is the case, the request is answered with 'Method Not Allowed'
	// and HTTP status code 405.
	// If no other Method is allowed, the request is delegated to the NotFound
	// handler.
	HandleMethodNotAllowed bool

	// Configurable http.Handler which is called when no matching route is
	// found. If it is not set, http.NotFound is used.
	NotFound http.Handler

	// Configurable http.Handler which is called when a request
	// cannot be routed and HandleMethodNotAllowed is true.
	// If it is not set, http.Error with http.StatusMethodNotAllowed is used.
	MethodNotAllowed http.Handler

	// Function to handle panics recovered from http handlers.
	// It should be used to generate a error page and return the http error code
	// 500 (Internal Server Error).
	// The handler can be used to keep your server from crashing because of
	// unrecovered panics.
	//
	// No middleware is applied to the PanicHandler.
	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

// DefaultConfig is the default Handler configuration used by New.
var DefaultConfig = Config{
	RedirectTrailingSlash:  true,
	RedirectFixedPath:      true,
	HandleMethodNotAllowed: true,
}

// New creates and initializes a new Handler using default settings
// and the given options.
func New(opts ...ConfigOption) *Handler {
	c := DefaultConfig
	for _, o := range opts {
		o.Set(&c)
	}
	return NewHandler(&c)
}

// NewHandler creates and initializes a new Handler with the given config.
func NewHandler(c *Config) *Handler {
	h := &Handler{
		prefix: c.Prefix,
		routes: make(map[string]*route),
	}
	router := httprouter.New()
	router.RedirectTrailingSlash = c.RedirectTrailingSlash
	router.RedirectFixedPath = c.RedirectFixedPath
	router.HandleMethodNotAllowed = c.HandleMethodNotAllowed
	if c.NotFound != nil {
		router.NotFound = h.chain.Then(c.NotFound)
	}
	if c.MethodNotAllowed != nil {
		router.MethodNotAllowed = h.chain.Then(c.MethodNotAllowed)
	}
	router.PanicHandler = c.PanicHandler
	h.router = router
	return h
}

// Use the given middleware.
func (h *Handler) Use(mw ...chain.Constructor) {
	h.chain = h.chain.Append(mw...)
}

// Append appends a handler to this handler, under the given pattern. All
// middleware from the root tree propagates to the subtree. However,
// the subtree's configuration such as prefix and fallback handlers,
// like NotFound and MethodNotAllowed, are ignored by the root tree
// in favor of its own configuration.
func (h *Handler) Append(pattern string, subtree *Handler) {
	for pp, route := range subtree.routes {
		pp = path.Join(h.prefix, pattern, pp)
		f := subtree.chain.ThenFunc(route.Handler)
		h.router.Handle(route.Method, pp, h.wrap(f.ServeHTTP))
	}
}

// Delete is a shortcut for mux.Handle("DELETE", path, handle)
func (h *Handler) Delete(pattern string, f http.HandlerFunc) {
	h.Handle("DELETE", pattern, f)
}

// Get is a shortcut for mux.Handle("GET", path, handle)
func (h *Handler) Get(pattern string, f http.HandlerFunc) {
	h.Handle("GET", pattern, f)
}

// Head is a shortcut for mux.Handle("HEAD", path, handle)
func (h *Handler) Head(pattern string, f http.HandlerFunc) {
	h.Handle("HEAD", pattern, f)
}

// Options is a shortcut for mux.Handle("OPTIONS", path, handle)
func (h *Handler) Options(pattern string, f http.HandlerFunc) {
	h.Handle("OPTIONS", pattern, f)
}

// Patch is a shortcut for mux.Handle("PATCH", path, handle)
func (h *Handler) Patch(pattern string, f http.HandlerFunc) {
	h.Handle("PATCH", pattern, f)
}

// Post is a shortcut for mux.Handle("POST", path, handle)
func (h *Handler) Post(pattern string, f http.HandlerFunc) {
	h.Handle("POST", pattern, f)
}

// Put is a shortcut for mux.Handle("PUT", path, handle)
func (h *Handler) Put(pattern string, f http.HandlerFunc) {
	h.Handle("PUT", pattern, f)
}

// HandleAll will register a pattern with a handler for All requests.
func (h *Handler) HandleAll(pattern string, f http.HandlerFunc) {
	methods := []string{"HEAD", "GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	for _, method := range methods {
		h.Handle(method, pattern, f)
	}
}

// Handle registers a new request handler with the given method and pattern.
func (h *Handler) Handle(method, pattern string, f http.Handler) {
	h.HandleFunc(method, pattern, f.ServeHTTP)
}

// HandleFunc registers a new request handler with the given method and pattern.
func (h *Handler) HandleFunc(method, pattern string, f http.HandlerFunc) {
	p := path.Join(h.prefix, pattern)
	if len(pattern) > 1 && pattern[len(pattern)-1] == '/' {
		p += "/"
	}
	h.routes[pattern] = &route{Method: method, Handler: f}
	h.router.Handle(method, p, h.wrap(f.ServeHTTP))
}

func (h *Handler) wrap(next http.HandlerFunc) httprouter.Handle {
	nextHandler := h.chain.ThenFunc(next)
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		if c, ok := r.Body.(*ctxBody); !ok {
			c = &ctxBody{
				ReadCloser: r.Body,
				ctx:        context.Background(),
			}
			r.Body = c
			defer func() {
				r.Body = c.ReadCloser
			}()
		}
		ctx := context.WithValue(Context(r), paramsID, p)
		SetContext(ctx, r)
		nextHandler.ServeHTTP(w, r)
	}
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// ServeFiles serves files from the given file system root.
//
// The pattern must end with "/*filepath" to have files served from
// the local path /path/to/dir/*filepath.
//
// For example, if root is "/etc" and *filepath is "passwd", the local
// file "/etc/passwd" is served. Because an http.FileServer is used
// internally it's possible that http.NotFound is called instead
// of httpmux's NotFound handler.
//
// To use the operating system's file system implementation, use
// http.Dir: mux.ServeFiles("/src/*filepath", http.Dir("/var/www")).
func (h *Handler) ServeFiles(pattern string, root http.FileSystem) {
	if !strings.HasSuffix(pattern, "/*filepath") {
		panic("pattern must end with /*filepath in path '" + pattern + "'")
	}
	fs := http.FileServer(root)
	h.Get(pattern, func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = Params(r).ByName("filepath")
		fs.ServeHTTP(w, r)
	})
}

// ctxBody is the object we save in the request's Body field.
type ctxBody struct {
	io.ReadCloser
	ctx context.Context
}

// Context returns the context from the given request, or a new
// context.Background if it doesn't exist.
func Context(r *http.Request) context.Context {
	if c, ok := r.Body.(*ctxBody); ok {
		return c.ctx
	}
	return context.Background()
}

// SetContext updates the given context in the request if the request
// has been previously instrumented by httpmux.
func SetContext(ctx context.Context, r *http.Request) {
	if c, ok := r.Body.(*ctxBody); ok {
		c.ctx = ctx
	}
}

type paramsType int

var paramsID paramsType

// Params returns the httprouter.Params from the request context.
func Params(r *http.Request) httprouter.Params {
	if p, ok := Context(r).Value(paramsID).(httprouter.Params); ok {
		return p
	}
	return httprouter.Params{}
}
