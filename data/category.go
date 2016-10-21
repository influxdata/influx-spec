package data

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
)

func GetDataDirs(root, pattern string) []string {
	var specs []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "data.lineprotocol" {
			specs = append(specs, filepath.Dir(path))
		}
		return nil
	})

	return specs
}

func NewCategories(dirs []string) []spec.Spec {
	var cats []spec.Spec
	for _, dir := range dirs {
		cats = append(cats, NewCategory(dir))
	}
	return cats
}

func NewCategory(dir string) *Category {
	c := &Category{
		Dir:   dir,
		name:  filepath.Base(dir),
		specs: map[string]specification{},
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".json") {
			testName := strings.TrimSuffix(path, ".json")
			s := c.specs[testName]
			s.ResultFile = path
			c.specs[testName] = s
		}
		if strings.HasSuffix(path, ".query") {
			testName := strings.TrimSuffix(path, ".query")
			s := c.specs[testName]
			s.QueryFile = path
			c.specs[testName] = s
		}
		return nil
	})

	return c
}

type Category struct {
	Dir   string
	name  string
	specs map[string]specification
}

func (c *Category) Name() string {
	return c.name
}

func (c *Category) Seed(cfg write.ClientConfig) (int, error) {
	f, err := os.Open(c.Dir + "/data.lineprotocol")
	if err != nil {
		return 0, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	client := write.NewClient(cfg)
	client.Create("")

	buf := bytes.NewBuffer(nil)

	ctr := 0
	for {
		ctr++

		line, err := r.ReadBytes('\n')
		if err == io.EOF {
			if _, _, err := client.Send(buf.Bytes()); err != nil {
				return ctr, err
			}
			return ctr, nil
		}

		if err != nil {
			return ctr, err
		}

		buf.Write(line)

		if ctr%5000 == 0 {
			if _, _, err := client.Send(buf.Bytes()); err != nil {
				return ctr, err
			}
			buf.Reset()
		}
	}
}

func (c *Category) Test(cfg write.ClientConfig) error {
	for _, s := range c.specs {
		err := s.Test(cfg)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

type specification struct {
	QueryFile  string
	ResultFile string
}

func (s *specification) Name() string {
	return s.QueryFile
}

func (s *specification) Seed(cfg write.ClientConfig) error {
	return nil
}

func (s *specification) Test(cfg write.ClientConfig) error {
	q, err := ioutil.ReadFile(s.QueryFile)
	if err != nil {
		return err
	}

	exp, err := ioutil.ReadFile(s.ResultFile)
	if err != nil {
		return err
	}

	vals := url.Values{}
	vals.Set("epoch", "ns") // need to fix
	vals.Set("q", string(q))
	vals.Set("db", cfg.Database)
	resp, err := http.PostForm(cfg.BaseURL+"/query", vals)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	got, err := ioutil.ReadAll(resp.Body)

	eq, err := JSONEqual(exp, got)
	if err != nil {
		return err
	}

	if !eq {
		return fmt.Errorf("expected:\n%s\vgot:\n%s\n", exp, got)
	}

	return nil
}

func JSONEqual(l, r []byte) (bool, error) {
	var li, ri interface{}
	if err := json.Unmarshal(l, &li); err != nil {
		return false, err
	}
	if err := json.Unmarshal(r, &ri); err != nil {
		return false, err
	}
	return reflect.DeepEqual(li, ri), nil
}
