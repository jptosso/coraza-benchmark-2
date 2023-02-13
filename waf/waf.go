package waf

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type HttpObject struct {
	reqHeaders  http.Header
	respHeaders io.ReadCloser
	ReqBody     string
	RespBody    string
}

type WafRunner interface {
	Init() error
	EvaluateString(string) error
	EvaluateFile(string) error
	Evaluate(*http.Request, *http.Response) error
	Cleanup() error
	Name() string
}

type Case struct {
	Name      string
	Variables map[string]string
	Request   string
	Response  string
}

func (c *Case) HTTPRequest() (*http.Request, error) {
	var err error
	resp := strings.ReplaceAll(c.Request, "\n", "\r\n")
	r := &http.Request{
		Header: make(http.Header),
	}
	inHeaders := false
	inBody := false
	for i, l := range strings.Split(resp, "\r\n") {
		if i == 0 {
			// request line
			spl := strings.SplitN(l, " ", 3)
			if len(spl) != 3 {
				return nil, fmt.Errorf("invalid request line: %s", l)
			}
			r.Method = spl[0]
			r.URL, err = url.Parse(spl[1])
			r.Proto = spl[2]
			if err != nil {
				return nil, err
			}
			inHeaders = true
		} else if inHeaders {
			// headers
			if l == "" {
				inHeaders = false
				inBody = true
				continue
			}
			spl := strings.Split(l, ":")
			r.Header.Add(strings.TrimSpace(spl[0]), strings.TrimSpace(spl[1]))
		} else if inBody {
			// body is l to io.ReadCloser
			r.Body = io.NopCloser(strings.NewReader(l))
		}
	}
	return r, nil
}

func (c *Case) HTTPResponse() (*http.Response, error) {
	resp := strings.ReplaceAll(c.Response, "\n", "\r\n")
	r := &http.Response{
		Header: make(http.Header),
	}
	inHeaders := false
	inBody := false
	for i, l := range strings.Split(resp, "\r\n") {
		if i == 0 {
			// response line
			spl := strings.SplitN(l, " ", 3)
			if len(spl) != 3 {
				return nil, fmt.Errorf("invalid response line: %s", l)
			}
			r.StatusCode, _ = strconv.Atoi(spl[1])
			r.Status = fmt.Sprintf("%s %s", spl[1], spl[2])
			proto := strings.SplitN(spl[0], "/", 2)
			r.Proto = spl[0]
			r.ProtoMajor, _ = strconv.Atoi(proto[0])
			r.ProtoMinor, _ = strconv.Atoi(proto[1])
			inHeaders = true
		} else if inHeaders {
			// headers
			if l == "" {
				inHeaders = false
				inBody = true
				continue
			}
			spl := strings.Split(l, ":")
			r.Header.Add(strings.TrimSpace(spl[0]), strings.TrimSpace(spl[1]))
		} else if inBody {
			// body is l to io.ReadCloser
			r.Body = io.NopCloser(strings.NewReader(l))
		}
	}
	return r, nil
}

func NewCase(data []byte) (*Case, error) {
	c := &Case{}
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return c, nil
}
