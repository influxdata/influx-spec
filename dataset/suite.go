package dataset

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

// GetDatasetDirs walks the filesystem starting at path `root` looking for
// directories that contain a data.lineprotocol file, returning a slice of all
// the directories it comes across.
func GetDatasetDirs(root string) []string {
	var specs []string

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && info.Name() == "data.lineprotocol" {
			specs = append(specs, filepath.Dir(path))
		}
		return nil
	})

	return specs
}

// NewSuites creates a slice of spec.Specs for each directory passed in as dirs
func NewSuites(dirs []string, filter string) []spec.Spec {
	var cats []spec.Spec
	for _, dir := range dirs {
		cat := NewSuite(dir)
		if strings.Contains(cat.name, filter) {
			cats = append(cats, cat)
		}
	}
	return cats
}

// NewSuite walks the direcory starting at `dir` adding pairs of files
// (*.query and *.json). If a pair is found, then it is added to the suite.
func NewSuite(dir string) *Suite {
	c := &Suite{
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

// Suite implements the spec.Spec interface. Logically, a suite represents
// a dataset directory, and map of subtests.
type Suite struct {
	Dir   string
	name  string
	specs map[string]specification
}

// Name returns the name of the Suite that is being ran.
func (c *Suite) Name() string {
	return c.name
}

// Seed will seed an InfluxDB instance with all of the data
// in <dir>/data.lineprotocol files and returns the number of lines
// that were written.
func (c *Suite) Seed(cfg write.ClientConfig) (int, error) {
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

// Test runs each of the specifications in the Suite. If an error is
// encountered, then the error is written to stderr.
func (c *Suite) Test(cfg write.ClientConfig) (rs []spec.Result, err error) {
	for _, s := range c.specs {
		r, err := s.Test(cfg)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r...)
	}
	return
}

// Teardown drops the database associated with a Suite.
func (c *Suite) Teardown(cfg write.ClientConfig) error {
	client := write.NewClient(cfg)
	client.Create(fmt.Sprintf("DROP DATABASE %v", cfg.Database))

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

func (s *specification) Teardown(cfg write.ClientConfig) error {
	return nil
}

func (s *specification) Test(cfg write.ClientConfig) ([]spec.Result, error) {
	q, err := ioutil.ReadFile(s.QueryFile)
	if err != nil {
		return nil, err
	}

	exp, err := ioutil.ReadFile(s.ResultFile)
	if err != nil {
		return nil, err
	}

	vals := url.Values{}
	vals.Set("epoch", "ns") // need to fix
	vals.Set("q", string(q))
	vals.Set("db", cfg.Database)
	resp, err := http.PostForm(cfg.BaseURL+"/query", vals)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	got, err := ioutil.ReadAll(resp.Body)

	eq, err := JSONEqual(exp, got)
	if err != nil {
		return nil, err
	}

	r := spec.Result{
		Pass:        eq,
		Description: string(q),
		Expected:    string(exp),
		Got:         string(got),
	}

	return []spec.Result{r}, nil
}

// JSONEqual checks to see if two byte slices encode the same
// underlying JSON object.
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
