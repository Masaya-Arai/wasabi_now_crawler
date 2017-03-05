package util

import (
	"regexp"
)

type Html struct {
	Url  Url
	Body string
}

func (html *Html) replace(rs []RegExp) *Html {
	html.Body = Replace(html.Body, rs)

	return html
}

func (html *Html) RemoveAllTags() *Html {
	rs := []RegExp{}
	rs = append(rs, RegExp{`<[^>]+?>`, ``})

	return html.replace(rs)
}

func (html *Html) RemoveNonTextElements() *Html {
	rs := []RegExp{}

	// CSS・JavaScript・NoScript・iFrameを削除
	rs = append(rs, RegExp{`(?i)<style.*?>.*?</style>`, ``})
	rs = append(rs, RegExp{`(?i)<script.*?>.*?</script>`, ``})
	rs = append(rs, RegExp{`(?i)<noscript>.*?</noscript>`, ``})
	rs = append(rs, RegExp{`(?i)<iframe.*?>.*?</iframe>`, ``})

	// Form要素を削除
	rs = append(rs, RegExp{`(?i)<form[^>]*?>.*?</form>`, ``})
	rs = append(rs, RegExp{`(?i)<input[^>]*?>`, ``})
	rs = append(rs, RegExp{`(?i)<label[^>]*?>.*?</label>`, ``})
	rs = append(rs, RegExp{`(?i)<select[^>]*?>.*?</select>`, ``})
	rs = append(rs, RegExp{`(?i)<option[^>]*?>.*?</option>`, ``})

	// brとhrを削除
	rs = append(rs, RegExp{`(?i)<[bh]r.*?>`, ``})

	return html.replace(rs)
}

func (html *Html) Compress() *Html {
	rs := []RegExp{}

	// スペース要素を圧縮
	rs = append(rs, RegExp{`　`, ` `})
	rs = append(rs, RegExp{`[\f\n\r\t]+`, ` `})
	rs = append(rs, RegExp{`\s+`, ` `})

	// コメントを削除
	rs = append(rs, RegExp{`<!--.*?-->`, ``})

	return html.replace(rs)
}

func (html *Html) Simplify() *Html {
	rs := []RegExp{}

	// anchor・imageを整形
	rs = append(rs, RegExp{`(?i)<a[^>]*?(href="[^"]*?").*?>`, `<a $1 target="_blank">`})
	rs = append(rs, RegExp{`(?i)<img[^>]*?(src="[^"]*?").*?>`, `<img $1>`})
	rs = append(rs, RegExp{`(?i)<(a[^>]+)>|<(img[^>]+)>|<([a-z0-9]+)[^>]*>`, `<$1$2$3>`})

	html.replace(rs)

	rs2 := []RegExp{}

	// divisionを圧縮・整形（Golangはlookahead/behindに対応していない）
	// TODO : <div>の軽量化。<div>.*?</div>を内包する部分の処理が不完全
	rs2 = append(rs2, RegExp{`(?i)<div>(\s*?)<div>(([^<]|<[^d]|<d[^i]|<di[^v]|<div[^>])*?)</div>(\s*?)</div>`, `<div>$2</div>`})
	rs2 = append(rs2, RegExp{`(?i)<[^/i>]+[^>]*?>\s*?</[^>]+?>`, ``})

	// TODO : 再帰的マッチングの効率化
	var oldStr string
	for oldStr != html.Body {
		oldStr = html.Body
		for _, r := range rs2 {
			patt := regexp.MustCompile(r.Pattern)
			repl := r.Replace
			html.Body = patt.ReplaceAllString(html.Body, repl)
		}
	}

	// TODO : テキストを含む<div>を<p>へ変換
	// TODO : <p>.*?</p>を内包する<div.*?</div>の削除
	/*
	pattf := regexp.MustCompile(`<div>(([^<]|<[^d]|<d[^i]|<di[^v]|<div[^>])*?)</div>`)
	replf := `<p>$1</p>`
	html.Body = pattf.ReplaceAllString(html.Body, replf)
	*/

	return html
}

func (html *Html) ReplaceReference() *Html {
	rs := []RegExp{}
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="([^"]*?)".*?>`,
			       `<a href="` + GetRegularUrl(html.Url.Url, `$1`) + `" target="_blank">`})
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="([^"]*?)".*?>`,
			       `<img src="` + GetRegularUrl(html.Url.Url, `$1`) + `">`})

	return html.replace(rs)
}

func (html *Html) ReplaceUrl() *Html {
	aPrefix := `<a href="`
	aSuffix := `" target="_blank">`
	imgPrefix := `<img src="`
	imgSuffix := `">`

	rs := []RegExp{}

	// 絶対URL
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="(https?:\/\/[^"]*?)".*?>`,
			       aPrefix + "$1" + aSuffix })
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="(https?:\/\/[^"]*?)".*?>`,
			       imgPrefix + "$1" + imgSuffix})

	// スキーマ指定
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="([\/]{2}[^\/][^"]*?)".*?>`,
			       aPrefix + html.Url.getScheme() + `:$1` + aSuffix})
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="([\/]{2}[^\/][^"]*?)".*?>`,
			       imgPrefix + html.Url.getScheme() + `:$1` + imgSuffix})

	// 絶対パス(1)
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="(\/{1}([^\/][^"]*)?)".*?>`,
			       aPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + `$1` + aSuffix})
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="(\/{1}([^\/][^"]*)?)".*?>`,
			       imgPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + `$1` + imgSuffix})

	// 絶対パス(3-)
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="\/{3,}([^\/][^"]*?)".*?>`,
			       aPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + `/$1` + aSuffix})
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="\/{3,}([^\/][^"]*?)".*?>`,
			       imgPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + `/$1` + imgSuffix})

	// 相対パス
	rs = append(rs, RegExp{`(?i)<a[^>]*?href="([^(https?:\/\/|\/)][^"]*?)".*?>`,
			       aPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + html.Url.GetUpperPath() + `/$1` + aSuffix})
	rs = append(rs, RegExp{`(?i)<img[^>]*?src="([^(https?:\/\/|\/)][^"]*?)".*?>`,
			       imgPrefix + html.Url.getScheme() + `://` + html.Url.getHost() + html.Url.GetUpperPath() + `/$1` + imgSuffix})

	return html.replace(rs)
}
























