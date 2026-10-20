package command

import (
	"testing"

	"modstart-cli/lib/api"
)

func TestParseParamsJSON(t *testing.T) {
	raw, err := parseParamsJSON(`{
		"page": 1,
		"pageSize": 20,
		"enable": true,
		"title": "标题",
		"ids": [4, 5],
		"meta": {"a": 1}
	}`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	def := &api.ApiDef{
		Key: "test.index",
		Params: []api.ParamDef{
			{Name: "page", Type: "int"},
			{Name: "pageSize", Type: "int"},
			{Name: "enable", Type: "bool"},
			{Name: "title", Type: "string"},
			{Name: "ids", Type: "string"},
			{Name: "meta", Type: "json"},
		},
	}
	params := normalizeParams(def, raw)
	cases := map[string]string{
		"page":     "1",
		"pageSize": "20",
		"enable":   "true",
		"title":    "标题",
		"ids":      "4,5",
		"meta":     `{"a":1}`,
	}
	for k, want := range cases {
		if got := params[k]; got != want {
			t.Errorf("param %s = %q, want %q", k, got, want)
		}
	}
}

func TestParseParamsJSONEmpty(t *testing.T) {
	params, err := parseParamsJSON("")
	if err != nil {
		t.Fatalf("empty parse failed: %v", err)
	}
	if len(params) != 0 {
		t.Errorf("expected empty params, got %v", params)
	}
}

func TestParseParamsJSONInvalid(t *testing.T) {
	invalid := []string{
		`[1,2]`,
		`"text"`,
		`page=1`,
		`{"page":1} extra`,
	}
	for _, raw := range invalid {
		if _, err := parseParamsJSON(raw); err == nil {
			t.Errorf("expected error for %q", raw)
		}
	}
}

func TestNormalizeParamsArrayTypes(t *testing.T) {
	raw, err := parseParamsJSON(`{"ids":[4,5],"meta":{"a":1},"list":[{"id":1}]}`)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	def := &api.ApiDef{
		Key: "test.delete",
		Params: []api.ParamDef{
			{Name: "ids", Type: "string"},
			{Name: "meta", Type: "object"},
		},
	}
	params := normalizeParams(def, raw)
	if params["ids"] != "4,5" {
		t.Errorf("ids = %q, want 4,5", params["ids"])
	}
	if params["meta"] != `{"a":1}` {
		t.Errorf("meta = %q", params["meta"])
	}
	if params["list"] != `[{"id":1}]` {
		t.Errorf("list = %q", params["list"])
	}
}
