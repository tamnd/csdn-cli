package csdn

import "testing"

func TestParseArticleRef(t *testing.T) {
	cases := []struct {
		in       string
		wantUser string
		wantID   string
	}{
		{"https://blog.csdn.net/alice/article/details/123456", "alice", "123456"},
		{"blog.csdn.net/bob_dev/article/details/789", "bob_dev", "789"},
		{"alice/123456", "alice", "123456"},
	}
	for _, c := range cases {
		user, id, err := ParseArticleRef(c.in)
		if err != nil {
			t.Errorf("ParseArticleRef(%q) error: %v", c.in, err)
			continue
		}
		if user != c.wantUser || id != c.wantID {
			t.Errorf("ParseArticleRef(%q) = (%q, %q), want (%q, %q)", c.in, user, id, c.wantUser, c.wantID)
		}
	}
	if _, _, err := ParseArticleRef("123456"); err == nil {
		t.Error("a bare article id should be rejected as ambiguous")
	}
	if _, _, err := ParseArticleRef(""); err == nil {
		t.Error("empty input should error")
	}
}

func TestParseUsername(t *testing.T) {
	cases := map[string]string{
		"https://blog.csdn.net/alice": "alice",
		"blog.csdn.net/bob_dev":       "bob_dev",
		"carol":                       "carol",
	}
	for in, want := range cases {
		got, err := ParseUsername(in)
		if err != nil {
			t.Errorf("ParseUsername(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("ParseUsername(%q) = %q, want %q", in, got, want)
		}
	}
	if _, err := ParseUsername("phoenix"); err == nil {
		t.Error("reserved segment phoenix should be rejected")
	}
	if _, err := ParseUsername("https://blog.csdn.net/article"); err == nil {
		t.Error("reserved segment article in url should be rejected")
	}
	if _, err := ParseUsername(""); err == nil {
		t.Error("empty input should error")
	}
}

func TestStripEm(t *testing.T) {
	if got := stripEm("a <em>match</em> here"); got != "a match here" {
		t.Errorf("stripEm = %q", got)
	}
	if got := stripEm("<EM>X</EM>y"); got != "Xy" {
		t.Errorf("stripEm case-insensitive = %q", got)
	}
}

func TestStripTags(t *testing.T) {
	in := `<div>hello <b>world</b><script>var x=1;</script><style>.a{}</style> &amp; more</div>`
	got := stripTags(in)
	want := "hello world & more"
	if got != want {
		t.Errorf("stripTags = %q, want %q", got, want)
	}
}

func TestExtractBalancedJSON(t *testing.T) {
	s := `prefix={"a":1,"b":{"c":"}"}}suffix`
	obj, ok := extractBalancedJSON(s, 7)
	if !ok {
		t.Fatal("expected a balanced object")
	}
	want := `{"a":1,"b":{"c":"}"}}`
	if obj != want {
		t.Errorf("extractBalancedJSON = %q, want %q", obj, want)
	}
	if _, ok := extractBalancedJSON("no braces here", 0); ok {
		t.Error("expected false when no object present")
	}
}

func TestUnescapeHTML(t *testing.T) {
	in := `&lt;a&gt; &amp; &quot;b&quot; &#39;c&#39;&nbsp;d`
	got := unescapeHTML(in)
	want := `<a> & "b" 'c' d`
	if got != want {
		t.Errorf("unescapeHTML = %q, want %q", got, want)
	}
}
