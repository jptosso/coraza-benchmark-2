package main

import (
	"fmt"
	"testing"
)

func BenchmarkWaf(b *testing.B) {
	cases, err := getCases("cases/")
	if err != nil {
		b.Error(err)
	}
	runners, err := getWafRunners()
	if err != nil {
		b.Error(err)
	}
	for _, c := range cases {
		req, err := c.HTTPRequest()
		if err != nil {
			b.Error(err)
		}
		resp, err := c.HTTPResponse()
		if err != nil {
			b.Error(err)
		}
		for _, w := range runners {
			waf := w()
			b.Run(fmt.Sprintf("%s/%s", c.Name, waf.Name()), func(b *testing.B) {
				waf := w()
				if err := waf.Init(); err != nil {
					b.Error(err)
				}
				if err := loadCrs(c.Variables, waf); err != nil {
					b.Error(err)
				}
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					if err := waf.Evaluate(req, resp); err != nil {
						b.Error(err)
					}
					if err := waf.Cleanup(); err != nil {
						b.Error(err)
					}
				}
			})
		}
	}
}
