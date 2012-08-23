package filter

import (
	"net/http"
)

// share common logger interface with core falcore
type Logger interface {
	// Matches the log4go interface
	Finest(arg0 interface{}, args ...interface{})
	Fine(arg0 interface{}, args ...interface{})
	Debug(arg0 interface{}, args ...interface{})
	Trace(arg0 interface{}, args ...interface{})
	Info(arg0 interface{}, args ...interface{})
	Warn(arg0 interface{}, args ...interface{}) error
	Error(arg0 interface{}, args ...interface{}) error
	Critical(arg0 interface{}, args ...interface{}) error
}

func Finest(arg0 interface{}, args ...interface{}) {
	log.Finest(arg0, args...)
}

func Fine(arg0 interface{}, args ...interface{}) {
	log.Fine(arg0, args...)
}

func Debug(arg0 interface{}, args ...interface{}) {
	log.Debug(arg0, args...)
}

func Trace(arg0 interface{}, args ...interface{}) {
	log.Trace(arg0, args...)
}

func Info(arg0 interface{}, args ...interface{}) {
	log.Info(arg0, args...)
}

func Warn(arg0 interface{}, args ...interface{}) error {
	return log.Warn(arg0, args...)
}

func Error(arg0 interface{}, args ...interface{}) error {
	return log.Error(arg0, args...)
}

func Critical(arg0 interface{}, args ...interface{}) error {
	return log.Critical(arg0, args...)
}

var log Logger

func SetLogger(newLog Logger) {
	log = newLog
}

// Filter incomming requests and optionally return a response or nil.  
// Filters are chained together into a flow (the Pipeline) which will terminate
// if the Filter returns a response.  
type RequestFilter interface {
	FilterRequest(req *Request) *http.Response
}

// Helper to create a Filter by just passing in a func
//    filter = NewRequestFilter(func(req *Request) *http.Response { 
//			req.Headers.Add("X-Falcore", "is_cool")
//			return
//		})
func NewRequestFilter(f func(req *Request) *http.Response) RequestFilter {
	rf := new(genericRequestFilter)
	rf.f = f
	return rf
}

type genericRequestFilter struct {
	f func(req *Request) *http.Response
}

func (f *genericRequestFilter) FilterRequest(req *Request) *http.Response {
	return f.f(req)
}

// Filter outgoing responses. This can be used to modify the response
// before it is sent.  Modifying the request at this point will have no
// effect. 
type ResponseFilter interface {
	FilterResponse(req *Request, res *http.Response)
}

// Helper to create a Filter by just passing in a func
//    filter = NewResponseFilter(func(req *Request, res *http.Response) { 
//			// some crazy response magic
//			return
//		})
func NewResponseFilter(f func(req *Request, res *http.Response)) ResponseFilter {
	rf := new(genericResponseFilter)
	rf.f = f
	return rf
}

type genericResponseFilter struct {
	f func(req *Request, res *http.Response)
}

func (f *genericResponseFilter) FilterResponse(req *Request, res *http.Response) {
	f.f(req, res)
}
