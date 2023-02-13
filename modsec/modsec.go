package modsec

import (
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/anuraaga/go-modsecurity"
	"github.com/jptosso/coraza-benchmark/waf"
)

type Modsec struct {
	waf *modsecurity.Modsecurity
	rs  *modsecurity.RuleSet
}

func (c *Modsec) Init() error {
	ms, err := modsecurity.NewModsecurity()
	if err != nil {
		return err
	}
	ms.SetServerLogCallback(func(msg string) {

	})
	c.waf = ms
	c.rs = ms.NewRuleSet()
	return nil
}

func (c *Modsec) EvaluateString(data string) error {
	return c.rs.AddRules(data)
}

func (c *Modsec) EvaluateFile(data string) error {
	if strings.Contains(data, "*") {
		files, err := filepath.Glob(data)
		if err != nil {
			return err
		}
		for _, f := range files {
			if err := c.rs.AddFile(f); err != nil {
				return err
			}
		}
		return nil
	}
	return c.rs.AddFile(data)
}

func (c *Modsec) Evaluate(req *http.Request, resp *http.Response) error {
	tx, err := c.rs.NewTransaction("127.0.0.1:8080", 0, "127.0.0.1:8080", 0)
	tx.ProcessUri(req.URL.String(), req.Method, req.Proto)
	for k, vv := range req.Header {
		for _, v := range vv {
			if err := tx.AddRequestHeader(k, v); err != nil {
				return err
			}
		}
	}
	tx.ProcessRequestHeaders()
	if req.Body != nil {
		rbody, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		if len(rbody) > 0 {
			tx.AppendRequestBody(rbody)
		}
	}
	if err := tx.ProcessRequestBody(); err != nil {
		return err
	}
	for k, vv := range resp.Header {
		for _, v := range vv {
			tx.AddResponseHeader(k, v)
		}
	}
	respbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(respbody) > 0 {
		tx.AppendResponseBody(respbody)
	}
	if err := tx.ProcessResponseBody(); err != nil {
		return err
	}
	if err := tx.ProcessLogging(); err != nil {
		return err
	}
	tx.Cleanup()
	return nil
}

func (c *Modsec) Cleanup() error {
	return nil
}

func (c *Modsec) Name() string {
	return "modsecurity"
}

var _ waf.WafRunner = &Modsec{}

func NewModsec() waf.WafRunner {
	return &Modsec{}
}
