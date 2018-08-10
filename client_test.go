package jrpc2client

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/erikdubbelboer/fasthttp"
	"github.com/riftbit/jrpc2server"
	"github.com/stretchr/testify/assert"
)

// DemoAPI area
type DemoAPI struct{}

type TestArgs struct {
	ID string
}

type TestReply struct {
	LogID     string `json:"log_id"`
	UserAgent string `json:"user_agent"`
}

// Test Method to test
func (h *DemoAPI) Test(ctx *fasthttp.RequestCtx, args *TestArgs, reply *TestReply) error {
	if args.ID == "" {
		return &jrpc2server.Error{Code: jrpc2server.JErrorInvalidParams, Message: "ID should not be empty"}
	}
	reply.LogID = args.ID
	reply.UserAgent = string(ctx.Request.Header.UserAgent())
	return nil
}

// TestUserAgent Method to test user agent value
func (h *DemoAPI) TestUserAgent(ctx *fasthttp.RequestCtx, args *TestArgs, reply *TestReply) error {
	reply.UserAgent = string(ctx.Request.Header.UserAgent())
	return nil
}

func TestPrepare(t *testing.T) {
	api := jrpc2server.NewServer()
	err := api.RegisterService(new(DemoAPI), "demo")
	assert.Nil(t, err)
	reqHandler := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/api":
			api.APIHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}
	go fasthttp.ListenAndServe(":65001", reqHandler)
}

func TestBasicClientErrorOnWrongAddress(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:12345")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: ""}, dstP)
	assert.NotNil(t, err)
}

func TestBasicClientErrorOnAPIFormat(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://yandex.ru")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: ""}, dstP)
	assert.NotNil(t, err)
}

func TestBasicClientErrorOnAPIAnwser(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:65001")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: ""}, dstP)
	assert.NotNil(t, err)
}

func TestDefaultUserAgentClient(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:65001")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.TestUserAgent", TestArgs{ID: "TESTER_ID_TestDefaultUserAgentClient"}, dstP)
	assert.Nil(t, err)
	assert.Equal(t, userAgent, dstP.UserAgent)
}

func TestCustomUserAgentClient(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:65001")
	client.SetUserAgent("JsonRPC Test Client")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.TestUserAgent", TestArgs{ID: "TESTER_ID_TestCustomUserAgentClient"}, dstP)
	assert.Nil(t, err)
	assert.Equal(t, "JsonRPC Test Client", dstP.UserAgent)
}

func TestBasicAuthClient(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:65001")
	client.SetUserAgent("JsonRPC Test Client")
	client.SetBasicAuthHeader("user", "password")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: "TESTER_ID_TestBasicClient"}, dstP)
	assert.Nil(t, err)
	assert.Equal(t, "TESTER_ID_TestLoggingClient", dstP.LogID)
}

func TestLoggingDevNullClient(t *testing.T) {
	logger := &logrus.Logger{
		Out:       ioutil.Discard,
		Formatter: &logrus.JSONFormatter{DisableTimestamp: false},
		Level:     logrus.DebugLevel,
	}
	client := NewClientWithLogger(logger)
	client.SetBaseURL("http://127.0.0.1:65001")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: "TESTER_ID_TestLoggingDevNullClient"}, dstP)
	assert.Nil(t, err)
	assert.Equal(t, "TESTER_ID_TestLoggingClient", dstP.LogID)
}

func TestLoggingClient(t *testing.T) {
	logger := &logrus.Logger{
		Out:       os.Stdout,
		Formatter: &logrus.JSONFormatter{DisableTimestamp: false},
		Level:     logrus.DebugLevel,
	}
	client := NewClientWithLogger(logger)
	client.SetBaseURL("http://127.0.0.1:65001")
	dstP := &TestReply{}
	err := client.Call("/api", "demo.Test", TestArgs{ID: "TESTER_ID_TestLoggingClient"}, dstP)
	assert.Nil(t, err)
	assert.Equal(t, "TESTER_ID_TestLoggingClient", dstP.LogID)
}

func TestMapBasicClient(t *testing.T) {
	client := NewClient()
	client.SetBaseURL("http://127.0.0.1:65001")
	client.SetUserAgent("JsonRPC Test Client")
	client.SetBasicAuthHeader("user", "password")
	dstT, err := client.CallForMap("/api", "demo.Test", TestArgs{ID: "TESTER_ID_TestMapBasicClient"})
	assert.Nil(t, err)
	val, ok := dstT["log_id"]
	if assert.NotEqual(t, ok, false) {
		assert.Equal(t, "TESTER_ID_TestMapBasicClient", val)
	}
}