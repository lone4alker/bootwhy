package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/lone4alker/bootwhy/internal/analyze"
	"github.com/lone4alker/bootwhy/internal/diagnose"
	"github.com/lone4alker/bootwhy/internal/parse"
	"github.com/lone4alker/bootwhy/internal/rules"
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

	output.Reset()
	findings := []diagnose.Finding{
		{
			Contributor: analyze.Contributor{Unit: "NetworkManager-wait-online.service", Took: 8 * time.Second, Added: 6 * time.Second},
			Ranked: []diagnose.Scored{
				{
					Hypothesis: rules.Default()[6],
					Score:      11,
					Evidence:   []string{"journal: NetworkManager wait online"},
				},
			},
			Confidence: "high",
		},
	}
	if err := JSON(&output, parse.Times{Total: 10 * time.Second}, analyze.Result{Shape: analyze.Single}, findings); err != nil {
		t.Fatalf("JSON serialization with findings failed: %v", err)
	}
	if !json.Valid(output.Bytes()) {
		t.Fatal(output.String())
	}
}
