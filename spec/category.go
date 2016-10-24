package spec

import "github.com/influxdata/influx-stress/write"

type Spec interface {
	Seed(write.ClientConfig) (int, error)
	Test(write.ClientConfig) error
	Teardown(write.ClientConfig) error
	Name() string
}
