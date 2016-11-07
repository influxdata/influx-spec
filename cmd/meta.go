package cmd

import (
	"fmt"
	"os"

	"github.com/influxdata/influx-spec/meta"
	"github.com/influxdata/influx-spec/spec"
	"github.com/influxdata/influx-stress/write"
	"github.com/spf13/cobra"
)

var mf metaFlags

func init() {
	metaCmd := &cobra.Command{
		Use:   "meta",
		Short: "Run suite of tests to verify that meta queries return expected results.",
		Run:   runMeta,
	}

	RootCmd.AddCommand(metaCmd)

	metaCmd.Flags().StringVarP(&host, "host", "", "http://localhost:8086", "HTTP address for the InfluxDB instance.")
	metaCmd.Flags().BoolVarP(&mf.rp, "rp", "", false, "Run retention policy tests")
	metaCmd.Flags().BoolVarP(&mf.user, "user", "", false, "Run user tests")
}

func runMeta(cmd *cobra.Command, args []string) {
	if len(args) != 0 {
		cmd.Usage()
		return
	}

	results := []spec.Result{}
	for _, s := range mf.specs() {
		cfg := write.ClientConfig{
			BaseURL:   host,
			Database:  fmt.Sprintf("INFLUXDB_SPECIFICATION_TEST_%v", s.Name()),
			Precision: "s",
		}

		fmt.Printf("Seeding Data for %v\n", s.Name())
		_, err := s.Seed(cfg)
		if err != nil {
			fmt.Printf("Encountered Error: %v\n", err)
			os.Exit(1)
			return
		}

		fmt.Printf("Running Spec for %v\n", s.Name())
		rs, err := s.Test(cfg)
		if err != nil {
			fmt.Printf("Encountered Error: %v\n", err)
			os.Exit(1)
			return
		}
		results = append(results, rs...)

		fmt.Printf("Teardown for %v\n", s.Name())
		err = s.Teardown(cfg)
		if err != nil {
			fmt.Printf("Encountered Error: %v\n", err)
			os.Exit(1)
			return
		}

	}
	// TODO: add support for other formats
	success := true
	for _, r := range results {
		if !r.Pass {
			success = false
			fmt.Fprintf(os.Stderr, "Query: %v\nExpected: \n%s\vGot: \n%s\n", r.Description, r.Expected, r.Got)
		}
	}

	if !success {
		os.Exit(1)
	}
}

type metaFlags struct {
	rp   bool
	user bool
}

func (m metaFlags) specs() []spec.Spec {
	s := []spec.Spec{}

	if m.rp {
		s = append(s, meta.RP{})
	}

	if m.user {
		s = append(s, meta.User{})
	}

	return s
}
