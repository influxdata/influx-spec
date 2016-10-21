# InfluxDB Specification Tool
`influx-spec` is a tool for confirming that an InfluxDB instance satisties the InfluxDB specification.

```
InfluxDB Specification tool.

Usage:
  influx-spec [command]

Available Commands:
  data        Run suite of tests to verify data ADD BETTER DESCRIPTION.

Flags:
  -h, --help   help for influx-spec

Use "influx-spec [command] --help" for more information about a command.
```

### Data Subcommand
```
influx-spec data -h
Run suite of tests to verify data ADD BETTER DESCRIPTION.

Usage:
  influx-spec data [flags]

Flags:
  -f, --filter string   Run test that match this filter. Better description
```

## Use
```
$ influx-spec data mock_data_dir/
```
