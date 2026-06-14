package csdn

import (
	"context"
	"testing"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// TestDomainRegistersOps builds a kit App, registers the csdn domain, and
// asserts the six record operations are present. Those same ops back the CLI,
// the HTTP API, and the MCP server.
func TestDomainRegistersOps(t *testing.T) {
	app := kit.New(BaseIdentity())
	Domain{}.Register(app)
	have := map[string]bool{}
	for _, op := range app.Ops() {
		have[op.Meta().Name] = true
	}
	for _, w := range []string{"hot", "search", "article", "user", "posts", "comments"} {
		if !have[w] {
			t.Errorf("missing operation %q", w)
		}
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		in       string
		wantType string
		wantID   string
	}{
		{"https://blog.csdn.net/alice/article/details/123456", "article", "alice/123456"},
		{"alice/123456", "article", "alice/123456"},
		{"https://blog.csdn.net/bob_dev", "user", "bob_dev"},
		{"carol", "user", "carol"},
	}
	for _, c := range cases {
		gotType, gotID, err := Domain{}.Classify(c.in)
		if err != nil {
			t.Errorf("Classify(%q) error: %v", c.in, err)
			continue
		}
		if gotType != c.wantType || gotID != c.wantID {
			t.Errorf("Classify(%q) = (%q, %q), want (%q, %q)", c.in, gotType, gotID, c.wantType, c.wantID)
		}
	}
}

func TestClassifyUnrecognized(t *testing.T) {
	if _, _, err := (Domain{}).Classify(""); err == nil {
		t.Fatal("Classify(\"\") want error, got nil")
	}
}

func TestLocate(t *testing.T) {
	cases := []struct {
		uriType string
		id      string
		want    string
	}{
		{"article", "alice/123456", blogHost + "/alice/article/details/123456"},
		{"user", "bob_dev", blogHost + "/bob_dev"},
	}
	for _, c := range cases {
		got, err := Domain{}.Locate(c.uriType, c.id)
		if err != nil {
			t.Errorf("Locate(%q, %q) error: %v", c.uriType, c.id, err)
			continue
		}
		if got != c.want {
			t.Errorf("Locate(%q, %q) = %q, want %q", c.uriType, c.id, got, c.want)
		}
	}
	if _, err := (Domain{}).Locate("nope", "x"); err == nil {
		t.Fatal("Locate with unknown type want error, got nil")
	}
}

func TestClassifyLocateRoundTrip(t *testing.T) {
	for _, link := range []string{
		"https://blog.csdn.net/alice/article/details/123456",
		"https://blog.csdn.net/bob_dev",
	} {
		typ, id, err := Domain{}.Classify(link)
		if err != nil {
			t.Fatal(err)
		}
		url, err := Domain{}.Locate(typ, id)
		if err != nil {
			t.Fatal(err)
		}
		typ2, id2, err := Domain{}.Classify(url)
		if err != nil {
			t.Fatal(err)
		}
		if typ2 != typ || id2 != id {
			t.Fatalf("round trip drifted: (%q,%q) -> %q -> (%q,%q)", typ, id, url, typ2, id2)
		}
	}
}

func TestDefaults(t *testing.T) {
	var c kit.Config
	Defaults(&c)
	def := DefaultConfig()
	if c.Rate != def.Rate || c.Timeout != def.Timeout || c.UserAgent != def.UserAgent {
		t.Errorf("Defaults did not seed csdn baseline: %+v", c)
	}
}

func TestNewClientOverlays(t *testing.T) {
	sess, err := newClient(context.Background(), kit.Config{
		UserAgent: "custom-agent",
		Rate:      time.Second,
		Timeout:   10 * time.Second,
		Retries:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	s, ok := sess.(*Session)
	if !ok {
		t.Fatalf("newClient returned %T, want *Session", sess)
	}
	if s.Client.UserAgent() != "custom-agent" {
		t.Errorf("UserAgent = %q, want custom-agent", s.Client.UserAgent())
	}
	if s.Client.cfg.Rate != time.Second || s.Client.cfg.Timeout != 10*time.Second || s.Client.cfg.Retries != 2 {
		t.Errorf("overlay drift: %+v", s.Client.cfg)
	}
}

func TestMapErr(t *testing.T) {
	if MapErr(nil) != nil {
		t.Error("nil should map to nil")
	}
	if got := MapErr(ErrNotFound); errs.KindOf(got) != errs.KindNotFound {
		t.Errorf("ErrNotFound maps to %v, want not-found", errs.KindOf(got))
	}
	if got := MapErr(ErrWalled); errs.KindOf(got) != errs.KindNeedAuth {
		t.Errorf("ErrWalled maps to %v, want need-auth", errs.KindOf(got))
	}
}

func TestEffectiveLimit(t *testing.T) {
	if got := effectiveLimit(0, 40); got != 40 {
		t.Errorf("effectiveLimit(0, 40) = %d, want 40", got)
	}
	if got := effectiveLimit(10, 40); got != 10 {
		t.Errorf("effectiveLimit(10, 40) = %d, want 10", got)
	}
}
