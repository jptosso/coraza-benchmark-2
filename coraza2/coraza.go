package coraza2

import (
	"io"
	"net/http"

	"github.com/corazawaf/coraza/v2"
	"github.com/corazawaf/coraza/v2/seclang"
	"github.com/jptosso/coraza-benchmark/waf"
)

type Corazav2 struct {
	waf *coraza.Waf
}

func (c *Corazav2) Init() error {
	c.waf = coraza.NewWaf()
	return nil
}

func (c *Corazav2) EvaluateString(data string) error {
	parser, _ := seclang.NewParser(c.waf)
	return parser.FromString(data)
}

func (c *Corazav2) EvaluateFile(data string) error {
	parser, _ := seclang.NewParser(c.waf)
	return parser.FromFile(data)
}

func (c *Corazav2) Evaluate(req *http.Request, resp *http.Response) error {
	tx := c.waf.NewTransaction()
	defer tx.Clean()
	tx.ProcessConnection("", 0, "", 0)
	tx.ProcessURI(req.URL.String(), req.Method, req.Proto)
	for k, vv := range req.Header {
		for _, v := range vv {
			tx.AddRequestHeader(k, v)
		}
	}
	tx.ProcessRequestHeaders()
	if req.Body != nil {
		rbody, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}
		tx.RequestBodyBuffer.Write(rbody)
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
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	tx.ResponseBodyBuffer.Write(rbody)
	if _, err := tx.ProcessResponseBody(); err != nil {
		return err
	}
	tx.ProcessLogging()
	return nil
}

func (c *Corazav2) Cleanup() error {
	return nil
}

func (c *Corazav2) Name() string {
	return "coraza_v2"
}

var _ waf.WafRunner = &Corazav2{}

func NewCoraza() waf.WafRunner {
	return &Corazav2{}
}
