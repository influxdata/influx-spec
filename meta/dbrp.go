package meta

import (
	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
)

type RP struct{}

func (r RP) Seed(cfg write.ClientConfig) (int, error) {
	return 0, nil
}

func (r RP) Test(cfg write.ClientConfig) ([]spec.Result, error) {
	return nil, nil
}

func (r RP) Teardown(cfg write.ClientConfig) error {
	return nil
}

func (r RP) Name() string {
	return ""
}
