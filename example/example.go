//
// Package example provide an example how to use the Trunks library from setup
// to creating targets.
//
package example

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	liberrors "github.com/shuLhan/share/lib/errors"
	libhttp "github.com/shuLhan/share/lib/http"
	"github.com/shuLhan/share/lib/mlog"
	"github.com/shuLhan/share/lib/websocket"
	vegeta "github.com/tsenart/vegeta/v12/lib"

	"git.sr.ht/~shulhan/trunks"
)

const (
	pathExample      = "/example"
	pathExampleError = "/example/error"
)

const (
	websocketAddress = "127.0.0.1:28240"
)

type Example struct {
	trunks   *trunks.Trunks
	wsServer *websocket.Server

	targetExampleErrorGet vegeta.Target
	targetExampleGet      vegeta.Target
	targetExamplePostForm vegeta.Target
}

//
// New create, initialize, and setup an example service.
//
func New() (ex *Example, err error) {
	env := &trunks.Environment{
		ResultsDir:    "example/testdata/",
		ResultsSuffix: "example",
	}

	ex = &Example{}

	ex.trunks, err = trunks.New(env)
	if err != nil {
		return nil, fmt.Errorf("example: New: %w", err)
	}

	err = ex.registerEndpoints()
	if err != nil {
		return nil, fmt.Errorf("example: New: %w", err)
	}

	// Create and register endpoint for WebSocket server.
	wsOpts := &websocket.ServerOptions{
		Address: websocketAddress,
	}

	ex.wsServer = websocket.NewServer(wsOpts)

	err = ex.registerWebSocketEndpoints()
	if err != nil {
		return nil, fmt.Errorf("example: New: %w", err)
	}

	// Register targets for testing HTTP and WebSocket endpoints.
	err = ex.registerTargets()
	if err != nil {
		return nil, fmt.Errorf("example: New: %w", err)
	}

	return ex, nil
}

func (ex *Example) Start() (err error) {
	go func() {
		err = ex.wsServer.Start()
		if err != nil {
			mlog.Errf("example.Start: %s\n", err)
		}
	}()

	return ex.trunks.Start()
}

func (ex *Example) Stop() {
	ex.wsServer.Stop()
	ex.trunks.Stop()
}

//
// registerEndpoints register HTTP endpoints for testing.
//
func (ex *Example) registerEndpoints() (err error) {
	err = ex.trunks.Server.RegisterEndpoint(&libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		Path:         pathExample,
		RequestType:  libhttp.RequestTypeQuery,
		ResponseType: libhttp.ResponseTypeJSON,
		Call:         ex.pathExampleGet,
	})
	if err != nil {
		return err
	}

	err = ex.trunks.Server.RegisterEndpoint(&libhttp.Endpoint{
		Method:       libhttp.RequestMethodGet,
		Path:         pathExampleError,
		RequestType:  libhttp.RequestTypeQuery,
		ResponseType: libhttp.ResponseTypeJSON,
		Call:         ex.pathExampleErrorGet,
	})
	if err != nil {
		return err
	}

	err = ex.trunks.Server.RegisterEndpoint(&libhttp.Endpoint{
		Method:       libhttp.RequestMethodPost,
		Path:         pathExample,
		RequestType:  libhttp.RequestTypeForm,
		ResponseType: libhttp.ResponseTypeJSON,
		Call:         ex.pathExamplePostForm,
	})

	return err
}

func (ex *Example) registerWebSocketEndpoints() (err error) {
	err = ex.wsServer.RegisterTextHandler(http.MethodGet, pathExample,
		ex.handleWSExampleGet)
	if err != nil {
		return err
	}
	return nil
}

func (ex *Example) registerTargets() (err error) {
	targetHttp := &trunks.Target{
		Name:    "Example HTTP",
		Hint:    "This section provide an example of HTTP endpoints that can be tested and attacked.",
		BaseUrl: fmt.Sprintf("http://%s", ex.trunks.Env.ListenAddress),
		Opts: &trunks.AttackOptions{
			Duration:      300 * time.Second,
			RatePerSecond: 1,
		},
		Vars: trunks.KeyFormInput{
			"A": trunks.FormInput{
				Label: "A",
				Kind:  trunks.FormInputKindNumber,
				Value: "1",
			},
		},
		HttpTargets: []*trunks.HttpTarget{{
			Name:        "HTTP Get",
			Hint:        fmt.Sprintf("Test or attack endpoint %q using HTTP GET.", pathExample),
			Method:      libhttp.RequestMethodGet,
			Path:        pathExample,
			RequestType: libhttp.RequestTypeQuery,
			Headers: trunks.KeyFormInput{
				"X-Get": trunks.FormInput{
					Label: "X-Get",
					Hint:  "Custom HTTP header to be send.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1.1",
				},
			},
			Params: trunks.KeyFormInput{
				"Param1": trunks.FormInput{
					Label: "Param1",
					Hint:  "Parameter with number.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1",
				},
			},
			Run:         ex.runExampleGet,
			AllowAttack: true,
			Attack:      ex.attackExampleGet,
			PreAttack:   ex.preattackExampleGet,
		}, {
			Name:        "HTTP Error Get",
			Hint:        fmt.Sprintf("Test error on endpoint %q using HTTP GET.", pathExampleError),
			Method:      libhttp.RequestMethodGet,
			Path:        pathExampleError,
			RequestType: libhttp.RequestTypeQuery,
			Headers: trunks.KeyFormInput{
				"X-Get": trunks.FormInput{
					Label: "X-Get",
					Hint:  "Custom HTTP header to be send.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1.1",
				},
			},
			Params: trunks.KeyFormInput{
				"Param1": trunks.FormInput{
					Label: "Param1",
					Hint:  "Parameter with number.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1",
				},
			},
			Run:         ex.runExampleGet,
			AllowAttack: true,
			Attack:      ex.attackExampleErrorGet,
			PreAttack:   ex.preattackExampleErrorGet,
		}, {
			Name:        "HTTP Post Form",
			Hint:        fmt.Sprintf("Test or attack endpoint %q using HTTP POST.", pathExample),
			Method:      libhttp.RequestMethodPost,
			Path:        pathExample,
			RequestType: libhttp.RequestTypeForm,
			Headers: trunks.KeyFormInput{
				"X-PostForm": trunks.FormInput{
					Label: "X-PostForm",
					Hint:  "Custom HTTP header to be send.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1",
				},
			},
			Params: trunks.KeyFormInput{
				"Param1": trunks.FormInput{
					Label: "Param1",
					Hint:  "Parameter with number.",
					Kind:  trunks.FormInputKindNumber,
					Value: "1",
				},
				"Param2": trunks.FormInput{
					Label: "Param2",
					Hint:  "Parameter with string.",
					Kind:  trunks.FormInputKindString,
					Value: "a string",
				},
			},
			Run:         ex.runExamplePostForm,
			AllowAttack: true,
			PreAttack:   ex.preattackExamplePostForm,
			Attack:      ex.attackExamplePostForm,
		}, {
			Name:        "HTTP free form",
			Hint:        fmt.Sprintf("Test endpoint %q using custom HTTP method and/or content type.", pathExample),
			Method:      libhttp.RequestMethodGet,
			Path:        pathExample,
			RequestType: libhttp.RequestTypeForm,
			Headers: trunks.KeyFormInput{
				"X-FreeForm": trunks.FormInput{
					Label: "X-FreeForm",
					Hint:  "Custom HTTP header to be send.",
					Kind:  trunks.FormInputKindString,
					Value: "1",
				},
			},
			Params: trunks.KeyFormInput{
				"Param1": trunks.FormInput{
					Label: "X-FreeForm",
					Hint:  "Parameter with number.",
					Kind:  trunks.FormInputKindNumber,
					Value: "123",
				},
			},
			IsCustomizable: true,
		}},
	}

	err = ex.trunks.RegisterTarget(targetHttp)
	if err != nil {
		return err
	}

	targetWebSocket := &trunks.Target{
		Name:    "Example WebSocket",
		Hint:    "This section provide an example of WebSocket endpoints that can be tested.",
		BaseUrl: fmt.Sprintf("ws://%s", websocketAddress),
		Opts:    &trunks.AttackOptions{},
		Vars: trunks.KeyFormInput{
			"WebSocketVar": trunks.FormInput{
				Label: "WebSocketVar",
				Kind:  trunks.FormInputKindString,
				Value: "hello",
			},
		},
		WebSocketTargets: []*trunks.WebSocketTarget{{
			Name: "Similar to HTTP GET",
			Hint: "Test WebSocket endpoint with parameters.",
			Params: trunks.KeyFormInput{
				"Param1": trunks.FormInput{
					Label: "Param1",
					Hint:  "Parameter with kind is number.",
					Kind:  "number",
					Value: "123",
				},
			},
			Run: ex.runWebSocketGet,
		}},
	}

	err = ex.trunks.RegisterTarget(targetWebSocket)
	if err != nil {
		return err
	}

	return nil
}

func (ex *Example) pathExampleGet(epr *libhttp.EndpointRequest) ([]byte, error) {
	res := libhttp.EndpointResponse{}
	res.Code = http.StatusOK
	res.Message = pathExample
	res.Data = epr.HttpRequest.Form

	return json.Marshal(&res)
}

func (ex *Example) pathExampleErrorGet(epr *libhttp.EndpointRequest) ([]byte, error) {
	return nil, liberrors.Internal(fmt.Errorf("server error"))
}

func (ex *Example) pathExamplePostForm(epr *libhttp.EndpointRequest) ([]byte, error) {
	res := libhttp.EndpointResponse{}
	res.Code = http.StatusOK
	res.Message = pathExample
	res.Data = epr.HttpRequest.Form

	return json.Marshal(&res)
}

func (ex *Example) runExampleGet(req *trunks.RunRequest) (res *trunks.RunResponse, err error) {
	if req.Target.HttpClient == nil {
		req.Target.HttpClient = libhttp.NewClient(req.Target.BaseUrl, nil, true)
	}

	res = &trunks.RunResponse{}

	headers := req.HttpTarget.Headers.ToHttpHeader()
	params := req.HttpTarget.Params.ToUrlValues()

	httpRequest, err := req.Target.HttpClient.GenerateHttpRequest(
		req.HttpTarget.Method,
		req.HttpTarget.Path,
		req.HttpTarget.RequestType,
		headers,
		params,
	)
	if err != nil {
		return nil, err
	}

	err = res.SetHttpRequest(httpRequest)
	if err != nil {
		return nil, err
	}

	httpResponse, _, err := req.Target.HttpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	err = res.SetHttpResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ex *Example) preattackExampleErrorGet(rr *trunks.RunRequest) {
	ex.targetExampleErrorGet = vegeta.Target{
		Method: rr.HttpTarget.Method.String(),
		URL:    fmt.Sprintf("%s%s", rr.Target.BaseUrl, rr.HttpTarget.Path),
		Header: rr.HttpTarget.Headers.ToHttpHeader(),
	}

	q := rr.HttpTarget.Params.ToUrlValues().Encode()
	if len(q) > 0 {
		ex.targetExampleErrorGet.URL += "?" + q
	}

	fmt.Printf("preattackExampleErrorGet: %+v\n", ex.targetExampleErrorGet)
}

func (ex *Example) preattackExampleGet(rr *trunks.RunRequest) {
	ex.targetExampleGet = vegeta.Target{
		Method: rr.HttpTarget.Method.String(),
		URL:    fmt.Sprintf("%s%s", rr.Target.BaseUrl, rr.HttpTarget.Path),
		Header: rr.HttpTarget.Headers.ToHttpHeader(),
	}

	q := rr.HttpTarget.Params.ToUrlValues().Encode()
	if len(q) > 0 {
		ex.targetExampleGet.URL += "?" + q
	}

	fmt.Printf("preattackExampleGet: %+v\n", ex.targetExampleGet)
}

func (ex *Example) attackExampleErrorGet(rr *trunks.RunRequest) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		rr.HttpTarget.AttackLocker.Lock()
		*tgt = ex.targetExampleErrorGet
		rr.HttpTarget.AttackLocker.Unlock()
		return nil
	}
}

func (ex *Example) attackExampleGet(rr *trunks.RunRequest) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		rr.HttpTarget.AttackLocker.Lock()
		*tgt = ex.targetExampleGet
		rr.HttpTarget.AttackLocker.Unlock()
		return nil
	}
}

func (ex *Example) runExamplePostForm(req *trunks.RunRequest) (res *trunks.RunResponse, err error) {
	if req.Target.HttpClient == nil {
		req.Target.HttpClient = libhttp.NewClient(req.Target.BaseUrl, nil, true)
	}

	res = &trunks.RunResponse{}

	headers := req.HttpTarget.Headers.ToHttpHeader()
	params := req.HttpTarget.Params.ToUrlValues()

	httpRequest, err := req.Target.HttpClient.GenerateHttpRequest(
		req.HttpTarget.Method,
		req.HttpTarget.Path,
		req.HttpTarget.RequestType,
		headers,
		params,
	)
	if err != nil {
		return nil, err
	}

	err = res.SetHttpRequest(httpRequest)
	if err != nil {
		return nil, err
	}

	httpResponse, _, err := req.Target.HttpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	err = res.SetHttpResponse(httpResponse)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (ex *Example) preattackExamplePostForm(rr *trunks.RunRequest) {
	ex.targetExamplePostForm = vegeta.Target{
		Method: rr.HttpTarget.Method.String(),
		URL:    fmt.Sprintf("%s%s", rr.Target.BaseUrl, rr.HttpTarget.Path),
		Header: rr.HttpTarget.Headers.ToHttpHeader(),
	}

	q := rr.HttpTarget.Params.ToUrlValues().Encode()
	if len(q) > 0 {
		ex.targetExamplePostForm.Body = []byte(q)
	}

	fmt.Printf("preattackExamplePostForm: %+v\n", ex.targetExamplePostForm)
}

func (ex *Example) attackExamplePostForm(rr *trunks.RunRequest) vegeta.Targeter {
	return func(tgt *vegeta.Target) error {
		rr.HttpTarget.AttackLocker.Lock()
		*tgt = ex.targetExamplePostForm
		rr.HttpTarget.AttackLocker.Unlock()
		return nil
	}
}

func (ex *Example) handleWSExampleGet(ctx context.Context, req *websocket.Request) (res websocket.Response) {
	res.ID = req.ID
	res.Code = http.StatusOK
	res.Body = req.Body
	return res
}

func (ex *Example) runWebSocketGet(rr *trunks.RunRequest) (res interface{}, err error) {
	var wg sync.WaitGroup

	wsc := &websocket.Client{
		Endpoint: "ws://" + websocketAddress,
		HandleText: func(cl *websocket.Client, frame *websocket.Frame) error {
			res = frame.Payload()
			wg.Done()
			return nil
		},
	}

	err = wsc.Connect()
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(rr.WebSocketTarget.Params)
	if err != nil {
		return nil, err
	}

	req := websocket.Request{
		ID:     uint64(time.Now().UnixNano()),
		Method: http.MethodGet,
		Target: pathExample,
		Body:   string(body),
	}

	reqtext, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	err = wsc.SendText(reqtext)
	if err != nil {
		return nil, err
	}
	wg.Add(1)
	wg.Wait()

	_ = wsc.Close()

	return res, nil
}
