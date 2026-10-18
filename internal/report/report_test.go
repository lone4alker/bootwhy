package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/parse"
)

func TestTableNormal(t *testing.T) {
	var output bytes.Buffer
	Table(&output, parse.Times{Total: time.Second}, analyze.Result{Shape: analyze.Normal}, nil, "systemd-analyze")
	if !strings.Contains(output.String(), "normal") {
		t.Fatal(output.String())
	}
}

func TestJSON(t *testing.T) {
	var output bytes.Buffer
	if err := JSON(&output, parse.Times{}, analyze.Result{}, nil); err != nil {
		t.Fatal(err)
	}
	if !json.Valid(output.Bytes()) {
		t.Fatal(output.String())
	}
}
