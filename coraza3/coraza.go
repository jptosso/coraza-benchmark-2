package coraza3

import (
	"net/http"

	"github.com/corazawaf/coraza/v3"
	"github.com/jptosso/coraza-benchmark/waf"
)

type Corazav3 struct {
	waf coraza.WAF
}

func (c *Corazav3) Init() error {
	return nil
}

func (c *Corazav3) EvaluateString(data string) error {
	var err error
	cfg := coraza.NewWAFConfig().WithDirectives(data)
	c.waf, err = coraza.NewWAF(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Corazav3) EvaluateFile(data string) error {
	var err error
	cfg := coraza.NewWAFConfig().WithDirectivesFromFile(data)
	c.waf, err = coraza.NewWAF(cfg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Corazav3) Evaluate(req *http.Request, resp *http.Response) error {
	tx := c.waf.NewTransaction()
	defer tx.Close()
	tx.ProcessConnection("", 0, "", 0)
	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)
	for k, vv := range req.Header {
		for _, v := range vv {
			tx.AddRequestHeader(k, v)
		}
	}
	tx.ProcessRequestHeaders()
	if req.Body != nil {
		tx.ReadRequestBodyFrom(req.Body)
	}
	if _, err := tx.ProcessRequestBody(); err != nil {
		return err
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			tx.AddResponseHeader(k, v)
		}
	}
	tx.ProcessResponseHeaders(resp.StatusCode, resp.Proto)
	tx.ReadResponseBodyFrom(resp.Body)
	if _, err := tx.ProcessResponseBody(); err != nil {
		return err
	}
	tx.ProcessLogging()
	return nil
}

func (c *Corazav3) Cleanup() error {
	return nil
}

func (c *Corazav3) Name() string {
	return "coraza_v3"
}

var _ waf.WafRunner = &Corazav3{}

func NewCoraza() waf.WafRunner {
	return &Corazav3{}
}
