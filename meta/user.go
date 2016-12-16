package meta

import (
	"fmt"
	"io"

	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
	"github.com/influxdata/influxdb-client"
)

type user struct {
	name     string
	password string
	admin    bool
	grants   []grant
}

type grant struct {
	database  string
	privilege string
}

func (u user) CreateQuery() string {
	createUserQuery := "CREATE USER \"%v\" WITH PASSWORD '%v'"
	if u.admin {
		createUserQuery += " WITH ALL PRIVILEGES"
	}

	return fmt.Sprintf(createUserQuery, u.name, u.password)
}

func (u user) DropQuery() string {
	dropUserQuery := "DROP USER \"%v\""

	return fmt.Sprintf(dropUserQuery, u.name)
}

func (u user) GrantQueries() []string {
	qs := []string{}
	for _, g := range u.grants {
		qs = append(qs, fmt.Sprintf("GRANT %v ON \"%v\" TO \"%v\"", g.database, u.name))
	}

	return qs
}

func newAdmin() user {
	u := user{
		name:     "desa",
		password: "password",
		admin:    true,
	}

	return u
}

type User struct {
	user user
}

func NewUser() User {
	return User{user: newAdmin()}
}

func newQueryOptions(cfg write.ClientConfig) influxdb.QueryOptions {
	opt := influxdb.QueryOptions{
		Database: cfg.Database,
	}

	opt.Params = make(map[string]interface{})
	opt.Params["precision"] = cfg.Precision
	opt.Params["rp"] = cfg.RetentionPolicy
	opt.Params["consistency"] = cfg.Consistency

	return opt
}

func (u User) Seed(cfg write.ClientConfig) (int, error) {
	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return 0, err
	}
	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	err = querier.Execute(u.user.CreateQuery())
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func (u User) Test(cfg write.ClientConfig) ([]spec.Result, error) {
	results := []spec.Result{}

	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return nil, err
	}

	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	cur, err := querier.Select("SHOW USERS")
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

		name := r.ValueByName("user")
		admin := r.ValueByName("admin").(bool)

		if name == u.user.name {
			result := spec.Result{
				Pass:        admin == u.user.admin,
				Description: u.user.CreateQuery(),
				Expected:    fmt.Sprintf("%v", u.user.admin),
				Got:         fmt.Sprintf("%v", admin),
			}
			results = append(results, result)
		}

	}

	if len(results) == 0 {
		return nil, fmt.Errorf("User %v does not exist.", u.user.name)
	}

	return results, nil
}

func (u User) Teardown(cfg write.ClientConfig) error {
	client, err := influxdb.NewClient(cfg.BaseURL)
	if err != nil {
		return err
	}
	querier := client.Querier()

	opt := newQueryOptions(cfg)
	querier.QueryOptions = opt

	err = querier.Execute(u.user.DropQuery())
	if err != nil {
		return err
	}

	return nil
}

func (u User) Name() string {
	return "user"
}
