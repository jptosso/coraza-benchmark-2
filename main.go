package main

import (
	"os"

	"github.com/jptosso/coraza-benchmark/coraza2"
	"github.com/jptosso/coraza-benchmark/coraza3"
	"github.com/jptosso/coraza-benchmark/modsec"
	"github.com/jptosso/coraza-benchmark/waf"
)

func main() {
}

type corazaWrapper func() waf.WafRunner

func getWafRunners() ([]corazaWrapper, error) {
	return []corazaWrapper{
		coraza3.NewCoraza,
		coraza2.NewCoraza,
		modsec.NewModsec,
	}, nil
}

func getCases(path string) ([]*waf.Case, error) {
	// load all files in path
	d, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	cases := make([]*waf.Case, 0, len(d))
	for _, f := range d {
		if !f.IsDir() {
			f, err := os.ReadFile(path + "/" + f.Name())
			if err != nil {
				return nil, err
			}
			c, err := waf.NewCase(f)
			if err != nil {
				return nil, err
			}
			cases = append(cases, c)
		}
	}
	return cases, nil
}

func loadCrs(config map[string]string, w waf.WafRunner) error {
	rule := `SecAction "id:1,phase:1`
	for k, v := range config {
		rule += `,setvar:tx.` + k + `=` + v
	}
	rule += `"`
	if err := w.EvaluateString(rule); err != nil {
		return err
	}
	files := []string{
		"./rules/coraza.conf",
		"./rules/crs/crs-setup.conf.example",
		"./rules/crs/rules/*.conf",
	}
	for _, f := range files {
		if err := w.EvaluateFile(f); err != nil {
			return err
		}
	}
	return nil
}
