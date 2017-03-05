package util

import "testing"

func TestHtml_RemoveNonTextElements(t *testing.T) {
	in := `<noscript>This is testing.</noscript><div><div><p>aaa<a href="/test">bbb</a>ccc</p><div>ddd</div><div><div>eee</div></div><img src="/test"/></div></div><form method="POST" action="/test"><input type="text" name="test"><label for="test">Test Input</label><select><option value="1">1</option><option value="2">2</option><option value="3">3</option><option value="4">4</option></select></form><iframe src="/test"></iframe><script type="text/javascript" src="test.js" async></script><script type="text/javascript"><!-- testing -->(function(){var test=1;if(test==0){alert("Not Testing.");}console.log(test);}());</script><style>#test{display:none;}</style>`
	out := `<div><div><p>aaa<a href="/test">bbb</a>ccc</p><div>ddd</div><div><div>eee</div></div><img src="/test"/></div></div>`

	h := Html{Body: in}
	h.RemoveNonTextElements()
	if h.Body != out {
		t.Fatal("")
	}
}

func TestHtml_RemoveAllTags(t *testing.T) {
	in := `<div><div><p>aaa<a href="/test">bbb</a>ccc</p><div>ddd</div><div><div>eee</div></div><img src="/test"/></div></div>`
	out := `aaabbbcccdddeee`

	h := Html{Body: in}
	h.RemoveAllTags()
	if h.Body != out {
		t.Fatal("")
	}
}

func TestHtml_Compress(t *testing.T) {
	in := `aaa　bbb  ccc ddd
	eee`
	out := `aaa bbb ccc ddd eee`

	h := Html{Body: in}
	h.Compress()
	if h.Body != out {
		t.Fatal("")
	}
}

func TestHtml_Simplify(t *testing.T) {
	in := `<div><div><p>aaa<a href="/test">bbb</a>ccc</p><div>ddd</div><div><div>eee</div></div><img src="/test"/></div></div>`
	out := `<div><div><p>aaa<a href="/test" target="_blank">bbb</a>ccc</p><div>ddd</div><div>eee</div><img src="/test"></div></div>`

	h := Html{Body: in}
	h.Simplify()
	if h.Body != out {
		t.Fatal("")
	}
}


// 絶対URL（同一ドメイン）
func TestHtml_ReplaceUrl(t *testing.T) {
	in1 := `http://www.example.com/foo/bar/baz.html`
	in2 := `<a href="http://www.example.com/qux/quux.html">`
	out := `<a href="http://www.example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 絶対URL（別ドメイン）
func TestHtml_ReplaceUrl1(t *testing.T) {
	in1 := `http://www.example.com/foo/bar/baz.html`
	in2 := `<a href="https://trial.jp/qux/quux.html">`
	out := `<a href="https://trial.jp/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// スキーマ省略（同一ドメイン）
func TestHtml_ReplaceUrl2(t *testing.T) {
	in1 := `http://example.com/foo/bar/baz.html`
	in2 := `<a href="//example.com/qux/quux.html">`
	out := `<a href="http://example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// スキーマ省略（別ドメイン）
func TestHtml_ReplaceUrl3(t *testing.T) {
	in1 := `https://example.com/foo/bar/baz.html`
	in2 := `<a href="//www.trial.jp/qux/quux.html">`
	out := `<a href="https://www.trial.jp/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 絶対パス(1)
func TestHtml_ReplaceUrl5(t *testing.T) {
	in1 := `http://www.example.com/foo/bar/baz.html`
	in2 := `<a href="/qux/quux.html">`
	out := `<a href="http://www.example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 絶対パス(3)
func TestHtml_ReplaceUrl6(t *testing.T) {
	in1 := `https://www.example.com/foo/bar/baz.html`
	in2 := `<a href="///qux/quux.html">`
	out := `<a href="https://www.example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 絶対パス(4)
func TestHtml_ReplaceUrl7(t *testing.T) {
	in1 := `http://www.example.com/foo/bar/baz.html`
	in2 := `<a href="////qux/quux.html">`
	out := `<a href="http://www.example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 絶対パス(5)
func TestHtml_ReplaceUrl8(t *testing.T) {
	in1 := `https://www.example.com/foo/bar/baz.html`
	in2 := `<a href="/////qux/quux.html">`
	out := `<a href="https://www.example.com/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}

// 相対パス
func TestHtml_ReplaceUrl9(t *testing.T) {
	in1 := `http://www.example.com/foo/bar/baz.html`
	in2 := `<a href="qux/quux.html">`
	out := `<a href="http://www.example.com/foo/bar/qux/quux.html" target="_blank">`

	h := Html{Url: Url{Url: in1}, Body: in2}
	h.ReplaceUrl()
	if h.Body != out {
		t.Fatal(h.Body)
	}
}
