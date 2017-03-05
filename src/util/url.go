package util

import "net/url"

type Url struct {
	Url string
}


func GetRegularUrl(baseUrl, targetUrl string) string {
	bu, _ := url.Parse(baseUrl)
	tu, _ := url.Parse(targetUrl)

	res := bu.ResolveReference(tu)

	return res.String()
}

func RemoveParameters(targetUrl string) string {
	u, _ := url.Parse(targetUrl)

	return u.Scheme + "://" + u.Host + u.Path
}

func (u *Url) getScheme() string {
	parsedUrl, _ := url.Parse(u.Url)

	return parsedUrl.Scheme
}

func (u *Url) getHost() string {
	parsedUrl, _ := url.Parse(u.Url)

	return parsedUrl.Host
}

func (u *Url) GetUpperPath() string {
	parsedUrl, _ := url.Parse(u.Url)

	tmp := parsedUrl.Path
	rs := []RegExp{}
	rs = append(rs, RegExp{`^(.*)\/[^\/]*$`, "$1"})

	return Replace(tmp, rs)
}