package sux

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

/*************************************************************
 * Context
 *************************************************************/

const (
	// abortIndex int8 = math.MaxInt8 / 2
	abortIndex int8 = 63
)

// Context for http server
type Context struct {
	Req  *http.Request
	Resp http.ResponseWriter
	// current route Params, if route has var Params
	Params Params

	index int8
	// current router instance
	router *Router
	// context data, you can save some custom data.
	values map[string]interface{}
	// all handlers for current request
	handlers HandlersChain
}

// Init a context
func (c *Context) InitRequest(w http.ResponseWriter, req *http.Request, handlers HandlersChain) {
	c.Req = req
	c.Resp = w
	c.values = make(map[string]interface{})
	c.handlers = handlers
}

// HandlerName get the main handler name
func (c *Context) HandlerName() string {
	return nameOfFunction(c.handlers.Last())
}

// Handler returns the main handler.
func (c *Context) Handler() HandlerFunc {
	return c.handlers.Last()
}

// Values get all values
func (c *Context) Values() map[string]interface{} {
	return c.values
}

// Router get router instance
func (c *Context) Router() *Router {
	return c.router
}

// Set a value to context by key
func (c *Context) Set(key string, val interface{}) {
	c.values[key] = val
}

// Get a value from context
func (c *Context) Get(key string) interface{} {
	return c.values[key]
}

// Abort will abort at the end of this middleware run
func (c *Context) Abort() {
	c.index = abortIndex
}

// Next run next handler
func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))

	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// Reset context data
func (c *Context) Reset() {
	// c.Writer = &c.writermem
	c.index = -1
	c.Params = nil
	c.values = nil
	c.handlers = nil
	// c.Errors = c.Errors[0:0]
	// c.Accepted = nil
}

// Copy a new context
func (c *Context) Copy() *Context {
	var ctx = *c
	ctx.handlers = nil
	ctx.index = abortIndex

	return &ctx
}

/*************************************************************
 * Context: request data
 *************************************************************/

// URL get URL instance from request
func (c *Context) URL() *url.URL {
	return c.Req.URL
}

// Query return query value by key
func (c *Context) Query(key string) string {
	if vs, ok := c.Req.URL.Query()[key]; ok && len(vs) > 0 {
		return vs[0]
	}

	return ""
}

// Param returns the value of the URL param.
// 		router.GET("/user/{id}", func(c *sux.Context) {
// 			// a GET request to /user/john
// 			id := c.Param("id") // id == "john"
// 		})
func (c *Context) Param(key string) string {
	return c.Params.String(key)
}

// Header return header value by key
func (c *Context) Header(key string) string {
	if values, _ := c.Req.Header[key]; len(values) > 0 {
		return values[0]
	}

	return ""
}

// RawData return stream data
func (c *Context) RawData() ([]byte, error) {
	return ioutil.ReadAll(c.Req.Body)
}

// IsWebSocket returns true if the request headers indicate that a webSocket
// handshake is being initiated by the client.
func (c *Context) IsWebSocket() bool {
	if strings.Contains(strings.ToLower(c.Header("Connection")), "upgrade") &&
		strings.ToLower(c.Header("Upgrade")) == "websocket" {
		return true
	}
	return false
}

// ClientIP implements a best effort algorithm to return the real client IP
func (c *Context) ClientIP() string {
	clientIP := c.Header("X-Forwarded-For")
	if index := strings.IndexByte(clientIP, ','); index >= 0 {
		clientIP = clientIP[0:index]
	}

	clientIP = strings.TrimSpace(clientIP)
	if len(clientIP) > 0 {
		return clientIP
	}

	clientIP = strings.TrimSpace(c.Header("X-Real-Ip"))
	if len(clientIP) > 0 {
		return clientIP
	}

	// if c.AppEngine {
	// 	if addr := c.Req.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
	// 		return addr
	// 	}
	// }

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Req.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

/*************************************************************
 * Context: response data
 *************************************************************/

// Write byte data to response
func (c *Context) Write(bt []byte) (n int, err error) {
	return c.Resp.Write(bt)
}

// WriteString to response
func (c *Context) WriteString(str string) (n int, err error) {
	return c.Resp.Write([]byte(str))
}

// Text writes out a string as plain text.
func (c *Context) Text(status int, str string) (err error) {
	c.Resp.WriteHeader(status)
	c.Resp.Header().Set("Content-Type", "text/plain; charset=UTF-8")

	_, err = c.Resp.Write([]byte(str))
	return
}

// JSONBytes writes out a string as json data.
func (c *Context) JSONBytes(status int, bs []byte) (err error) {
	c.Resp.WriteHeader(status)
	c.Resp.Header().Set("Content-Type", "application/json; charset=UTF-8")

	_, err = c.Resp.Write(bs)
	return
}

// NoContent serve success but no content response
func (c *Context) NoContent() error {
	c.Resp.WriteHeader(http.StatusNoContent)
	return nil
}

// SetHeader for the response
func (c *Context) SetHeader(key, value string) {
	c.Resp.Header().Set(key, value)
}

// SetStatus code for the response
func (c *Context) SetStatus(status int) {
	c.Resp.WriteHeader(status)
}