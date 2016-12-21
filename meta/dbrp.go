package meta

import (
	"fmt"
	"io"

	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
	"github.com/influxdata/influxdb-client"
)

type dbrp struct {
	name string
	with *retentionPolicy
	rps  []retentionPolicy
}

func (d *dbrp) CreateQuery() string {
	createDBQuery := "CREATE DATABASE \"%v\""
	if d.with != nil {
		createDBQuery += " WITH DURATION %v REPLICATION %v SHARD DURATION %v NAME \"%v\""

		return fmt.Sprintf(createDBQuery,
			d.name, d.with.duration, d.with.replication, d.with.shardDuration, d.with.name)
	}

	return fmt.Sprintf(createDBQuery, d.name)
}

func (d *dbrp) DropQuery() string {
	dropDBQuery := "DROP DATABASE \"%v\""

	return fmt.Sprintf(dropDBQuery, d.name)
}

type retentionPolicy struct {
	name          string
	duration      string
	shardDuration string
	replication   int
	isDefault     bool
}

func (r *retentionPolicy) equal(name, duration, shardDuration string, replication int, isDefault bool) bool {
	return name == r.name && duration == r.duration && shardDuration == r.shardDuration &&
		replication == r.replication && isDefault == r.isDefault
}

func (r *retentionPolicy) CreateQuery(db string) string {
	createRPQuery := "CREATE RETENTION POLICY \"%v\" ON \"%v\" DURATION %v REPLICATION %v SHARD DURATION %v"
	if r.isDefault {
		createRPQuery += " DEFAULT"
	}

	return fmt.Sprintf(createRPQuery, r.name, db, r.duration, r.replication, r.shardDuration)
}

func (r *retentionPolicy) DropQuery(db string) string {
	dropRPQuery := "DROP RETENTION POLICY \"%v\" ON \"%v\""

	return fmt.Sprintf(dropRPQuery, r.name, db)
}

func NewDBRP(name string) dbrp {
	return dbrp{
		name: name,
		rps: []retentionPolicy{
			{name: "myrp", duration: "24h0m0s", shardDuration: "1h0m0s", replication: 1, isDefault: true},
			{name: "myotherrp", duration: "672h0m0s", shardDuration: "24h0m0s", replication: 2, isDefault: false},
		},
	}
}

func (d dbrp) Seed(cfg write.ClientConfig) (int, error) {
	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return 0, err
	}
	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	err = querier.Execute(d.CreateQuery())
	if err != nil {
		return 0, err
	}

	for _, rp := range d.rps {
		err = querier.Execute(rp.CreateQuery(d.name))
		if err != nil {
			return 0, err
		}
	}

	return 0, nil
}

func (d dbrp) Test(cfg write.ClientConfig) ([]spec.Result, error) {
	results := []spec.Result{}

	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	cur, err := querier.Select("SHOW DATABASES")
	if err != nil {
		return nil, err
	}

	rs, err := cur.NextSet()
	if err != nil {
		return nil, err
	}

	s, err := rs.NextSeries()
	if err != nil {
		return nil, err
	}

	for {
		r, err := s.NextRow()

		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		name := r.ValueByName("name")

		if name == d.name {
			result := spec.Result{
				Pass:        true,
				Description: d.CreateQuery(),
				Expected:    fmt.Sprintf("%v", d.name),
				Got:         fmt.Sprintf("%v", name),
			}
			results = append(results, result)
			break
		}

	}

	// Fix should add failing test
	if len(results) == 0 {
		return nil, fmt.Errorf("Database %v does not exist.", d.name)
	}

	for _, rp := range d.rps {
		rpResults := []spec.Result{}

		cur, err := querier.Select(fmt.Sprintf("SHOW RETENTION POLICIES ON \"%v\"", d.name))
		if err != nil {
			return nil, err
		}

		rs, err := cur.NextSet()
		if err != nil {
			return nil, err
		}

		s, err := rs.NextSeries()
		if err != nil {
			return nil, err
		}

		for {
			r, err := s.NextRow()

			if err == io.EOF {
				break
			}

			if err != nil {
				return nil, err
			}

			name := r.ValueByName("name").(string)
			duration := r.ValueByName("duration").(string)
			shardDuration := r.ValueByName("shardGroupDuration").(string)
			replication := int(r.ValueByName("replicaN").(float64))
			isDefault := r.ValueByName("default").(bool)

			if name == rp.name {
				result := spec.Result{
					Pass:        true,
					Description: rp.CreateQuery(d.name),
					Expected: fmt.Sprintf("Name: %v Duration: %v Shard Duration: %v Replication: %v Default: %v",
						rp.name, rp.duration, rp.shardDuration, rp.replication, rp.isDefault),
					Got: fmt.Sprintf("Name: %v Duration: %v Shard Duration: %v Replication: %v Default: %v",
						name, duration, shardDuration, replication, isDefault),
				}

				if !rp.equal(name, duration, shardDuration, replication, isDefault) {
					result.Pass = false
				}

				results = append(results, rpResults...)
				break
			}

		}
	}

	return results, nil
}

func (r dbrp) Teardown(cfg write.ClientConfig) error {
	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return err
	}
	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	err = querier.Execute(r.DropQuery())
	if err != nil {
		return err
	}

	return nil
}

func (r dbrp) Name() string {
	return "dbrp"
}
