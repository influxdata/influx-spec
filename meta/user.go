package meta

import (
	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
)

type User struct{}

func (u User) Seed(cfg write.ClientConfig) (int, error) {
	return 0, nil
}

func (u User) Test(cfg write.ClientConfig) ([]spec.Result, error) {
	return nil, nil
}

func (u User) Teardown(cfg write.ClientConfig) error {
	return nil
}

func (u User) Name() string {
	return ""
}
