# Build process
```
glide install
go install $(glide nv)
```

# InfluxDB Specification Tool
`influx-spec` is a tool for confirming that an InfluxDB instance satisties the InfluxDB specification.

```
InfluxDB Specification tool.

Usage:
  influx-spec [command]

Available Commands:
  dataset     Run suite of tests to verify that queries return expected results.
  meta        Run suite of tests to verify that meta queries return expected results.

Flags:
  -h, --help   help for influx-spec

Use "influx-spec [command] --help" for more information about a command.
```

## Data Subcommand
```
influx-spec data -h
Run suite of tests to verify data ADD BETTER DESCRIPTION.

Usage:
  influx-spec data [flags]

Flags:
  -f, --filter string   Run test that match this filter. Better description
```

### Use
```
$ influx-spec data mock_data_dir/
```

## Meta Subcommand

```
Run suite of tests to verify that meta queries return expected results.

Usage:
  influx-spec meta [flags]

Flags:
      --dbrp          Verify that you can create/destroy databases and retention policies.
      --host string   HTTP address for the InfluxDB instance. (default "http://localhost:8086")
      --user          Verify that you can create/modify/destroy users.
```

### Use
```
$ influx-spec meta --dbrp --user
```
