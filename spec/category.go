package spec

import "github.com/influxdata/influx-stress/write"

type Spec interface {
	Seed(write.ClientConfig) (int, error)
	Test(write.ClientConfig) (Results, error)
	Teardown(write.ClientConfig) error
	Name() string
}

type Result struct {
	Pass        bool
	Description string
	Expected    string
	Got         string
}

type Results []Result
