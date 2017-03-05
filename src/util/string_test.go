package util

import "testing"

func TestXPathToSelector(t *testing.T) {
	in := `//div[@id="sample"]//div[@class="trial"]/p[@class="part"]`
	out := `div#sample div.trial p.part`

	if XPathToSelector(in) != out {
		t.Fatal("")
	}
}
