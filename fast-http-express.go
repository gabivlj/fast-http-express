package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/valyala/fasthttp"
)

const (
	// GET >> Enum
	GET = iota
	// POST >> Enum
	POST
	// PUT >> Enum
	PUT
	// DELETE >> Enum
	DELETE
)

const (
	// GETSTRING >>
	GETSTRING = "GET"
	// POSTSTRING >>
	POSTSTRING = "POST"
	// PUTSTRING >>
	PUTSTRING = "PUT_STRING"
	// DELETESTRING >>
	DELETESTRING = "DELETE_STRING"
)

var paramsForCurrentConnections = make(map[uint64]map[string]string)
var hasFinishedTheConection = make(map[uint64]bool)

// Route >>
type Route struct {
	route       string
	typeRequest int
	middlewares []func(*fasthttp.RequestCtx) error
}

// Route >>
func (r *Route) Route() string {
	return r.route
}

// Middleware >>
func (r *Route) Middleware() []func(*fasthttp.RequestCtx) error {
	return r.middlewares
}

// RouteInterface >> Implements a route
type RouteInterface interface {
	// What will handle the request.
	Middleware() []func(*fasthttp.RequestCtx) error
	// Shall return a string of the route requested
	Route() string
}

// Router >>
type Router struct {
	// Get handlers.
	GetHandlers []RouteInterface
	// Post Handlers
	PostHandlers []RouteInterface
	// Put Handlers
	PutHandlers []RouteInterface
	// Delete Handlers
	DeleteHandlers []RouteInterface
	// Middlewares
	Middleware []func(*fasthttp.RequestCtx) error
	prefix     string
}

// RouterInterface >>
type RouterInterface interface {
	Get(string, ...func(*fasthttp.RequestCtx) error)
	Post(string, ...func(*fasthttp.RequestCtx) error)
	Put(string, ...func(*fasthttp.RequestCtx) error)
	Delete(string, ...func(*fasthttp.RequestCtx) error)
	Middlewares() []func(*fasthttp.RequestCtx) error
}

// NewRouter -> Creates a new router
func NewRouter(prefix string) *Router {
	router := &Router{Middleware: make([]func(*fasthttp.RequestCtx) error, 0), prefix: prefix}
	return router
}

// Use -> Adds a middleware.
func (r *Router) Use(handler func(*fasthttp.RequestCtx) error) {
	r.Middleware = append(r.Middleware, handler)
}

// Middlewares -> Returns middleware handlers
func (r *Router) Middlewares() []func(*fasthttp.RequestCtx) error {
	return r.Middleware
}

// Get -> Adds a Get() handler for the specified route.
func (r *Router) Get(route string, handlers ...func(*fasthttp.RequestCtx) error) *Route {
	routeObject := &Route{route: route, typeRequest: GET, middlewares: handlers}
	r.GetHandlers = append(r.GetHandlers, routeObject)
	return routeObject
}

// Post -> Adds a Post() handler for the specified route.
func (r *Router) Post(route string, handlers ...func(*fasthttp.RequestCtx) error) *Route {
	routeObject := &Route{route: route, typeRequest: POST, middlewares: handlers}
	r.PostHandlers = append(r.PostHandlers, routeObject)
	return routeObject
}

// Put -> Adds a Put() handler for the specified route.
func (r *Router) Put(route string, handlers ...func(*fasthttp.RequestCtx) error) *Route {
	routeObject := &Route{route: route, typeRequest: PUT, middlewares: handlers}
	r.PutHandlers = append(r.PutHandlers, routeObject)
	return routeObject
}

// Delete -> Adds a Delete() handler for the specified route.
func (r *Router) Delete(route string, handlers ...func(*fasthttp.RequestCtx) error) *Route {
	routeObject := &Route{route: route, typeRequest: DELETE, middlewares: handlers}
	r.DeleteHandlers = append(r.DeleteHandlers, routeObject)
	return routeObject
}

func executeMiddleware(request *fasthttp.RequestCtx, handlers []func(*fasthttp.RequestCtx) error) error {
	for _, handler := range handlers {
		err := handler(request)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *Router) executeHandlers(request *fasthttp.RequestCtx, handlers []RouteInterface) error {
	for _, get := range handlers {
		route := r.prefix + get.Route()
		if checkEntireRouteHTTP(string(request.RequestURI()), route, request.ConnRequestNum()) {
			// Loop through middlewares
			for _, handler := range get.Middleware() {
				err := handler(request)
				if err != nil {
					return err
				}
				if hasFinishedTheConection[request.ConnRequestNum()] == true {
					return nil
				}
			}
		}
	}
	return nil
}

// ProcessRequest Processes requests
func (r *Router) ProcessRequest(request *fasthttp.RequestCtx) error {
	defer func() {
		delete(hasFinishedTheConection, request.ConnRequestNum())
		delete(paramsForCurrentConnections, request.ConnRequestNum())
	}()
	hasFinishedTheConection[request.ConnRequestNum()] = false
	errMiddleware := executeMiddleware(request, r.Middlewares())
	if errMiddleware != nil {
		return errMiddleware
	}
	if hasFinishedTheConection[request.ConnRequestNum()] == true {
		return nil
	}
	switch true {
	case request.IsGet():
		err := r.executeHandlers(request, r.GetHandlers)
		if err != nil {
			return err
		}
		break
	case request.IsPost():
		err := r.executeHandlers(request, r.PostHandlers)
		if err != nil {
			return err
		}
		break
	case request.IsPut():
		err := r.executeHandlers(request, r.PutHandlers)
		if err != nil {
			return err
		}
		break
	case request.IsDelete():
		err := r.executeHandlers(request, r.DeleteHandlers)
		if err != nil {
			return err
		}
		break
	default:
		return errors.New("Non existing request method")
	}
	return nil
}

// FastModifiedHttp : fast-http, but with an abstraction layer where you can add middleware and control your routes better.
type FastModifiedHttp struct {
	routers []Router
}

// StartApp : Starts the application
func (f *FastModifiedHttp) StartApp(port string) {
	flag.Parse()
	h := f.processRouters
	if err := fasthttp.ListenAndServe(*addr, h); err != nil {
		log.Fatalf("Error in ListenAndServe: %s", err)
	}
}

// NewApp : Returns a FastModifiedHttp where you can add Routers
func NewApp(routers ...Router) *FastModifiedHttp {
	app := &FastModifiedHttp{routers: routers}
	return app
}

// Processes the routers
func (f *FastModifiedHttp) processRouters(request *fasthttp.RequestCtx) {
	for _, r := range f.routers {
		if checkPrefixHTTP(string(request.RequestURI()), r.prefix, request.ConnRequestNum()) {
			err := r.ProcessRequest(request)
			if err != nil {
				panic(err)
			}
			return
		}
	}
}

// Checks the entire route if it is good
func checkEntireRouteHTTP(route string, route2 string, currentNumberConnection uint64) bool {
	route2 = strings.Trim(route2, "/")
	route = strings.Trim(route, "/")
	splittedRoute := strings.Split(route, "/")
	splittedPrefix := strings.Split(route2, "/")

	if len(splittedPrefix) != len(splittedRoute) {
		return false
	}
	if len(splittedPrefix) == 0 && len(splittedRoute) == 0 {
		return true
	}
	if len(splittedPrefix) == 0 || len(splittedRoute) == 0 {
		return false
	}
	for i := 0; i < len(splittedRoute) && i < len(splittedPrefix); i++ {
		elementRoute, elementPrefix := splittedRoute[i], splittedPrefix[i]
		if len(elementRoute) == 0 && len(elementPrefix) == 0 {
			continue
		}
		if elementPrefix == "*" {
			return true
		}
		if len(elementPrefix) > 0 && elementPrefix[0] == byte(':') {
			theMap := paramsForCurrentConnections[currentNumberConnection]
			if theMap == nil {
				paramsForCurrentConnections[currentNumberConnection] = make(map[string]string)
			}
			paramsForCurrentConnections[currentNumberConnection][elementPrefix[1:]] = elementRoute
			continue
		}
		if elementRoute != elementPrefix {
			return false
		}
	}

	return true
}

// Checks the prefix for routers
func checkPrefixHTTP(route string, prefix string, currentNumberConnection uint64) bool {
	prefix = strings.Trim(prefix, "/")
	route = strings.Trim(route, "/")
	splittedRoute := strings.Split(route, "/")
	splittedPrefix := strings.Split(prefix, "/")
	if len(splittedPrefix) == 0 || splittedPrefix[0] == "" {
		return true
	}
	if len(splittedRoute) == 0 || splittedPrefix[0] == "" {
		return false
	}

	for i := 0; i < len(splittedRoute) && i < len(splittedPrefix); i++ {
		elementRoute, elementPrefix := splittedRoute[i], splittedPrefix[i]
		if len(elementRoute) == 0 && len(elementPrefix) == 0 {
			continue
		}
		if elementPrefix == "*" {
			return true
		}
		if len(elementPrefix) > 0 && elementPrefix[0] == byte(':') {
			theMap := paramsForCurrentConnections[currentNumberConnection]
			if theMap == nil {
				paramsForCurrentConnections[currentNumberConnection] = make(map[string]string)
			}
			paramsForCurrentConnections[currentNumberConnection][elementPrefix[1:]] = elementRoute
			continue
		}
		if elementRoute != elementPrefix {
			return false
		}
	}
	return true
}

// GetParams :: Return a map[string]string where you have mapped all the params that are passed to the route.
func GetParams(ctx *fasthttp.RequestCtx) map[string]string {
	return paramsForCurrentConnections[ctx.ConnRequestNum()]
}

// AddToRequestValue > Adds a key to the request body.
func AddToRequestValue(ctx *fasthttp.RequestCtx, key string, value interface{}) {
	ctx.SetUserValue(key, value)
}

// RequestKeyValue > Returns the specified key.
func RequestKeyValue(ctx *fasthttp.RequestCtx, key string) interface{} {
	return ctx.UserValue(key)
}

// RequestKeyValueBytes > Returns bytes of the key - value relation that you've added with AddToRequestValue. It uses json.Marshal to parse the interface.
func RequestKeyValueBytes(ctx *fasthttp.RequestCtx, key string) []byte {
	j, err := json.Marshal(ctx.UserValue(key))
	if err != nil {
		return nil
	}
	return j
}

// RequestKeyValueString > Returns string of the key - value relation that you've added with AddToRequestValue. It uses json.Marshal to parse the interface. Returns empty string if error or not found.
func RequestKeyValueString(ctx *fasthttp.RequestCtx, key string) string {
	j, err := json.Marshal(ctx.UserValue(key))
	if err != nil {
		return ""
	}
	return string(j)
}

// RespondJSON > Responds a json object, handles the content type and data parsing to json.
func RespondJSON(ctx *fasthttp.RequestCtx, data interface{}) error {
	// data := []byte{}
	hasFinishedTheConection[ctx.ConnRequestNum()] = true
	ctx.SetContentType("application/json")
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Fprint(ctx, string(js))
	return nil
}

// RespondBytes > Responds whatever you pass, it doesn't handle the content type.
func RespondBytes(ctx *fasthttp.RequestCtx, bytes []byte) {
	hasFinishedTheConection[ctx.ConnRequestNum()] = true
	fmt.Fprint(ctx, string(bytes))
}

// RespondText > Responds text that you pass
func RespondText(ctx *fasthttp.RequestCtx, str string) {
	hasFinishedTheConection[ctx.ConnRequestNum()] = true
	fmt.Fprintln(ctx, str)
}

// SendFile > Responds a file, contentType is for the content-type header value
func SendFile(ctx *fasthttp.RequestCtx, path string, contentType string) {
	hasFinishedTheConection[ctx.ConnRequestNum()] = true
	ctx.SetContentType(contentType)
	ctx.SendFile(path)
}

// Finished > If you've used the ctx response directly and not a response function of fast-http-express then call this function so it will stop going through handlers/middleware.
func Finished(ctx *fasthttp.RequestCtx) {
	hasFinishedTheConection[ctx.ConnRequestNum()] = true
}

// RequestKeyValueBytes > Returns the specified key (bytes).
// func RequestKeyValueBytes(ctx *fasthttp.RequestCtx, key string) []byte {
// 	switch
// 	return []byte(ctx.UserValue(key))
// }
