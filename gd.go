package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
)

var (
	isWindows = strings.Contains(runtime.GOOS, "windows")
	gp        = os.Getenv("GOPATH")
	wd, _     = os.Getwd()
)

type dependency struct {
	Import  string `json:"package,omitempty"`
	Version string `json:"version,omitempty"`
	Repo    string `json:"repo,omitempty"`
	Conf    config `json:"conf,omitempty"`
}

type config struct {
	Proxy string `json:"proxy,omitempty"`
}

func main() {
	if gp == "" {
		if isWindows {
			gp = path.Join(os.Getenv("USERPROFILE"), "go")
		} else {
			gp = path.Join(os.Getenv("HOME"), "go")
		}
	}

	_, e := exec.LookPath("git")
	if e != nil {
		log.Fatalln("git command not found")
	}

	log.Println("reading vendor.json")
	b, e := ioutil.ReadFile("vendor.json")
	if e != nil {
		log.Fatalln(e.Error())
	}
	b = bytes.TrimSpace(b)
	if len(b) <= 0 {
		return
	}

	deps, e := conf(b)
	if e != nil {
		log.Fatalf("parse vendor.json failed\n%s\n", e.Error())
	}
	if len(deps) <= 0 {
		return
	}
	imports := make([]string, 0, len(deps))
	for _, ele := range deps {
		imports = append(imports, ele.Import)
	}

	e = fc(deps)
	if e != nil {
		log.Fatalf("git clone/pull failed\n%s\n", e.Error())
	}

	e = ck(deps)
	if e != nil {
		log.Fatalf("git checkout failed\n%s\n", e.Error())
	}

	e = cpy(imports)
	if e != nil {
		log.Fatalf("copy to vendor directory failed\n%s\n", e.Error())
	}

	e = rc(deps)
	if e != nil {
		log.Fatalf("revert checkout failed\n%s\n", e.Error())
	}

	e = rm(imports)
	if e != nil {
		log.Fatalf("remove .git directory failed\n%s\n", e.Error())
	}

	log.Println("success")
}

func conf(data []byte) ([]dependency, error) {
	v := struct {
		Dependencies []dependency `json:"dependencies,omitempty"`
		Conf         config       `json:"conf,omitempty"`
	}{}
	e := json.Unmarshal(data, &v)
	if e != nil {
		return nil, e
	}
	deps := make([]dependency, 0, len(v.Dependencies))
	for _, d := range v.Dependencies {
		if d.Import == "" {
			continue
		}
		deps = append(deps, d)
		if d.Conf.Proxy == "" {
			d.Conf.Proxy = v.Conf.Proxy
		}
	}
	return deps, nil
}

// 1. git clone/pull
func fc(deps []dependency) error {
	base := path.Join(gp, "src")
	for _, d := range deps {
		log.Printf("fetching %s\n", d.Import)
		os.Chdir(base)
		repo := dr(d.Import)
		if d.Repo != "" {
			repo = d.Repo
		}
		if d.Conf.Proxy != "" {
			exec.Command("git", "config", "http.proxy", d.Conf.Proxy).Run()
		}

		var r []byte
		var e error
		p, b, n1, n2 := validate(d.Import, repo)
		if b {
			os.Chdir(p)
			r, e = exec.Command("git", "pull").Output()
		} else {
			if e = os.MkdirAll(p, os.ModePerm); e != nil {
				return e
			}
			os.Chdir(p)
			r, e = exec.Command("git", "clone", repo).Output()
			if n1 != "" && n2 != "" {
				err := os.Rename(n1, n2)
				if err != nil {
					return err
				}
			}
		}

		if d.Conf.Proxy != "" {
			exec.Command("git", "config", "--unset", "http.proxy").Run()
		}
		if e != nil {
			return errors.New(string(r))
		}
	}
	return nil
}

// 2. checkout branch/tag
func ck(deps []dependency) error {
	for _, d := range deps {
		if d.Version == "" {
			continue
		}
		log.Printf("checking out %s\n", d.Import)
		os.Chdir(path.Join(gp, "src", d.Import))
		v := d.Version
		if len(v) == 40 {
			v = d.Version[:10]
		}
		r, e := exec.Command("git", "checkout", "-q", v).Output()
		if e != nil {
			return errors.New(string(r))
		}
	}
	return nil
}

// 3. copy to vendor
func cpy(imports []string) error {
	for _, imp := range imports {
		src := path.Join(gp, "src", imp)
		if !exists(src) {
			continue
		}
		log.Printf("copying %s\n", imp)
		dst := path.Join(wd, "vendor", imp)
		var r []byte
		var e error
		if isWindows {
			dst += "\\"
			src = strings.Replace(src, "/", "\\", -1)
			dst = strings.Replace(dst, "/", "\\", -1)
			r, e = exec.Command("xcopy", src, dst, "/E", "/H", "/Q", "/Y").Output()
		} else {
			dst += "/"
			r, e = exec.Command("cp", "-a", src, dst).Output()
		}
		if e != nil {
			return errors.New(string(r))
		}
	}
	return nil
}

// 4. revert checkout in gopath
func rc(deps []dependency) error {
	for _, d := range deps {
		if d.Version == "" {
			continue
		}
		log.Printf("reverting check out %s\n", d.Import)
		os.Chdir(path.Join(gp, "src", d.Import))
		r, e := exec.Command("git", "checkout", "-q", "master").Output()
		if e != nil {
			return errors.New(string(r))
		}
	}
	return nil
}

// 5. remove .git directory in vendor
func rm(imports []string) error {
	for _, imp := range imports {
		p := path.Join(wd, "vendor", imp, ".git")
		log.Printf("removing %s\n", p)
		e := os.RemoveAll(p)
		if e != nil {
			return e
		}
	}
	return nil
}

// dr returns the repo according to the given package
func dr(imp string) string {
	return "https://" + imp + ".git"
}

// exists returns the direcoty exists or not
func exists(path string) bool {
	f, e := os.Stat(path)
	return e == nil && f.IsDir()
}

// validate returns:
// 1. path that should be chdir
// 2. .git directory exists or not
// 3. shoule rename the cloned repo's name(old name)
// 4. shoule rename the cloned repo to this one(new name)
func validate(imp, repo string) (string, bool, string, string) {
	if repo == "" {
		repo = dr(imp)
	}
	n := strings.Split(path.Base(repo), ".")[0]
	arr := strings.Split(imp, "/")
	var p string
	var m bool
	for _, ele := range arr {
		p = path.Join(p, ele)
		if ele == n {
			m = true
			break
		}
	}
	base := path.Join(gp, "src")
	if exists(path.Join(base, p, ".git")) {
		return path.Join(base, p), true, "", ""
	}
	if m {
		return path.Join(base, path.Dir(p)), false, "", ""
	}
	return path.Join(base, path.Dir(p)), false, n, arr[len(arr)-1]
}
