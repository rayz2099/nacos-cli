package output

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRender_TextEmpty_NoOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	if err := Render(buf, "text", "", nil); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if buf.String() != "" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestRenderTable(t *testing.T) {
	got := RenderTable("Total: 2  Page: 1/1", []string{"DATA_ID", "GROUP"}, [][]string{{"d1", "g1"}, {"d2", "g2"}})
	if !strings.Contains(got, "Total: 2  Page: 1/1") {
		t.Fatalf("missing summary: %q", got)
	}
	if !strings.Contains(got, "DATA_ID") || !strings.Contains(got, "GROUP") {
		t.Fatalf("missing headers: %q", got)
	}
	if !strings.Contains(got, "d1") || !strings.Contains(got, "g1") {
		t.Fatalf("missing rows: %q", got)
	}
}

func TestNormalizeConfigGetError_KnownPattern(t *testing.T) {
	err := NormalizeConfigGetError(errors.New("read config from both server and cache fail, err=timeout"), "d1", "g1", "ns1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "get config failed, dataId=d1, group=g1, namespace=ns1, reason=server unavailable and local cache missing" {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}

func TestNormalizeConfigGetError_UnknownPattern(t *testing.T) {
	orig := errors.New("some other error")
	err := NormalizeConfigGetError(orig, "d1", "g1", "ns1")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != orig.Error() {
		t.Fatalf("unexpected error: %q", err.Error())
	}
}
