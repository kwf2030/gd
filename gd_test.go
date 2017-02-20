package main

import (
	"fmt"
	"testing"
)

func TestVars(t *testing.T) {
	t.Logf("isWindows = %t\n", isWindows)
	t.Logf("gopath = %s\n", gp)
	t.Logf("working dir = %s\n", wd)
}

func TestExists(t *testing.T) {
	var data = []struct {
		path string
		want bool
	}{
		{"D:\\Workspace\\gopath\\src", true},
		{"D:\\Workspace\\gopath\\bin", true},
		{"D:\\Workspace\\gopath\\pkg", true},
		{"D:\\Workspace\\gopath\\src\\golang.org/x/tools/cmd/godoc", true},
		{"D:/Workspace/gopath/src/golang.org/", true},
		{"D:/Workspace/gopath/src/golang.org/x", true},
		{"D:/Workspace/gopath/src\\golang.org/x/net/html/", true},
		{"D:\\Workspaces\\gopath\\src", false},
		{"D:/Workspaces/gopath/src/", false},
		{"D:/Workspace/gopath\\gen/", false},
		{"E:\\Workspace\\gopath/bin", false},
	}

	for _, ele := range data {
		if r := exists(ele.path); r != ele.want {
			t.Errorf("exists(%s) = %t, want %t", ele.path, r, ele.want)
		}
	}
}

func TestValidate(t *testing.T) {
	var data = []struct {
		imp   string
		repo  string
		want1 string
		want2 bool
		want3 string
		want4 string
	}{
		{"github.com/andlabs/ui", "", fmt.Sprintf("%s/src/github.com/andlabs/ui", gp), true, "", ""},
		{"github.com/andlabs/ui", "https://github.com/andlabs/ui.git", fmt.Sprintf("%s/src/github.com/andlabs/ui", gp), true, "", ""},
		{"github.com/andlabs/ui", "https://github.com/andlabs/ui-wrong.git", fmt.Sprintf("%s/src/github.com/andlabs/ui", gp), true, "", ""},
		{"github.com/whoami/code", "https://gitlab.com/none/foobar.git", fmt.Sprintf("%s/src/github.com/whoami", gp), false, "foobar", "code"},
		{"golang.org/x/net/html", "", fmt.Sprintf("%s/src/golang.org/x/net", gp), false, "", ""},
		{"golang.org/x/net/html", "https://github.com/golang/x/net.git", fmt.Sprintf("%s/src/golang.org/x/net", gp), true, "", ""},
		{"golang.org/x/net/html", "https://github.com/golang/x/net-invalid.git", fmt.Sprintf("%s/src/golang.org/x/net", gp), false, "net-invalid", "html"},
		{"golang.org/x/net/norepo", "", fmt.Sprintf("%s/src/golang.org/x/net", gp), false, "", ""},
		{"golang.org/x/net/norepo", "https://github.com/golang/x/net/invalid.git", fmt.Sprintf("%s/src/golang.org/x/net", gp), false, "invalid", "norepo"},
	}

	for _, ele := range data {
		r1, r2, r3, r4 := validate(ele.imp, ele.repo)
		if r1 != ele.want1 || r2 != ele.want2 || r3 != ele.want3 || r4 != ele.want4 {
			t.Errorf("validate(%s, %s) = (%s, %t, %s, %s), want (%s, %t, %s, %s)", ele.imp, ele.repo, r1, r2, r3, r4, ele.want1, ele.want2, ele.want3, ele.want4)
		}
	}
}

func TestCpy(t *testing.T) {
	var data = []string{
		"github.com/cweill/gotests",
		"github.com/nsf/gocode",
		"github.com/rancher/trash",
		"github.com/s-rah/onionscan",
		"golang.org/x/net/http2",
		"golang.org/x/tools/go",
		"sourcegraph.com/sqs/goreturns",
		"github.com/x/none",
	}

	e := cpy(data)
	if e != nil {
		t.Errorf(e.Error())
	}
}
