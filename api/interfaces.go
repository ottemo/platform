package api

import (
	"io"
	"net/http"
)

// InterfaceSessionService is an interface to access session managing service
type InterfaceSessionService interface {
	GetName() string

	GC() error

	New() (InterfaceSession, error)
	Get(sessionID string, create bool) (InterfaceSession, error)

	IsEmpty(sessionID string) bool

	Touch(sessionID string) error
	Close(sessionID string) error

	GetKey(sessionID string, key string) interface{}
	SetKey(sessionID string, key string, value interface{})
}

// InterfaceSession is an interface represents private storage for particular API request
type InterfaceSession interface {
	GetID() string

	Get(key string) interface{}
	Set(key string, value interface{})

	IsEmpty() bool

	Touch() error
	Close() error
}

// InterfaceRestService is an interface to interact with RESTFul API service
type InterfaceRestService interface {
	GetName() string

	Run() error
	RegisterAPI(resource string, operation string, handler FuncAPIHandler) error

	http.Handler
}

// InterfaceApplicationContextSupport is an interface to assign/get application context to object
type InterfaceApplicationContextSupport interface {
	GetApplicationContext() InterfaceApplicationContext
	SetApplicationContext(context InterfaceApplicationContext) error
}

// InterfaceApplicationContext is an interface representing context where current execution happens
type InterfaceApplicationContext interface {
	GetRequest() interface{}
	GetResponse() interface{}

	GetSession() InterfaceSession
	SetSession(session InterfaceSession) error

	GetContextValues() map[string]interface{}
	GetContextValue(key string) interface{}
	SetContextValue(key string, value interface{})

	GetResponseWriter() io.Writer

	GetRequestArguments() map[string]string
	GetRequestArgument(name string) string
	GetRequestFiles() map[string]io.Reader
	GetRequestFile(name string) io.Reader
	GetRequestSettings() map[string]interface{}
	GetRequestSetting(name string) interface{}
	GetRequestContent() interface{}
	GetRequestContentType() string

	GetResponseContentType() string
	SetResponseContentType(mimeType string) error
	GetResponseSetting(name string) interface{}
	SetResponseSetting(name string, value interface{}) error
	GetResponseResult() interface{}
	SetResponseResult(value interface{}) error
}

// InterfaceJSONLogger is an interface to log mapped data
type InterfaceJSONLogger interface {
	GetName() string
	Log(storage string, data map[string]interface{}) error
}
