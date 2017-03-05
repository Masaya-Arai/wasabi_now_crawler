package util

import "testing"

func TestGetRegularUrl(t *testing.T) {
	in1 := "http://www.example.com"
	in2 := "sample"
	out := "http://www.example.com/sample"

	if res := GetRegularUrl(in1, in2); res != out {
		t.Fatal("")
	}
}

func TestGetRegularUrl2(t *testing.T) {
	in1 := "http://www.example.com/sample"
	in2 := "trial"
	out := "http://www.example.com/trial"

	if res := GetRegularUrl(in1, in2); res != out {
		t.Fatal("")
	}
}

func TestGetRegularUrl3(t *testing.T) {
	in1 := "http://www.example.com/sample/"
	in2 := "trial"
	out := "http://www.example.com/sample/trial"

	if res := GetRegularUrl(in1, in2); res != out {
		t.Fatal("")
	}
}

func TestGetRegularUrl4(t *testing.T) {
	in1 := "http://www.example.com/sample"
	in2 := "/trial"
	out := "http://www.example.com/trial"

	if res := GetRegularUrl(in1, in2); res != out {
		t.Fatal("")
	}
}

func TestGetRegularUrl5(t *testing.T) {
	in1 := "http://www.example.com/sample/a/b?param=val"
	in2 := "/trial"
	out := "http://www.example.com/trial"

	if res := GetRegularUrl(in1, in2); res != out {
		t.Fatal(res)
	}
}

func TestRemoveParameters(t *testing.T) {
	in := "http://example.com/sample?param1=val1&param2=val2"
	out := "http://example.com/sample"

	if res := RemoveParameters(in); res != out {
		t.Fatal(res)
	}
}

func Test(t *testing.T) {
	in := "http://example.com/foo/bar/baz?param1=val1&param2=val2"
	out := "/foo/bar"

	a := Url{Url: in}

	if res := a.GetUpperPath(); res != out {
		t.Fatal(res)
	}
}
