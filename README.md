# quest-ei
Generate simulated EI metric data for QuestDB.

## Installation
```shell
$ go get github.com/lnquy/quest-ei
```

### Build from source
```shell
$ git clone https://github.com/lnquy/quest-ei
$ cd quest-ei
$ go build -o quest-ei main.go
```

## Usage
#### Generate historical metrics
You can generate historical data by providing `--start` and `--end` time as below.  
Optionally, you can either save the generated metrics to file (via `--out-static-file` and `--out-metrics-file`) or ship the generated ILP messages directly to QuestDB at `127.0.0.1:9009` by default.  
For the `--out-metrics-file` option, metrics will be written to the file in Influx Line Protocol (ILP) format. You can then use `tsbs_load_questdb --file generated_file.ilp` to ingest the metrics to QuestDB. More information from [here](https://github.com/timescale/tsbs).
```shell
# Flush metric data to QuestDB running locally on the same host, 
# using default parameters.
$ quest-ei

# Write metric data to the file with customized parameters
$ quest-ei \
  --start=2022-01-01T00:00:00Z \
  --end=2022-01-10T00:00:00Z \
  --interval=30s \
  --min-load=0.0 \
  --max-load=1.0 \
  --sites=30 \
  --fleets-per-site=10 \
  --channels-per-site=10 \
  --talk-groups-per-site=20 \
  --units-per-talk-group=5 \
  --flush-batch-size=1000000 \
  --flush-batch-buffer-mb=500 \
  --out-static-file=qdb-static.json \
  --out-metrics-file=qdb-data.ilp

# Continue to generate historical metrics using the input from the previous generation via --in-static-file
$ quest-ei \
  --start=2022-01-10T00:00:00Z \
  --end=2022-02-01T00:00:00Z \
  --interval=1m \
  --min-load=0.1 \
  --max-load=0.7 \
  --flush-batch-size=1000000 \
  --flush-batch-buffer-mb=500 \
  --out-metrics-file=qdb-data-1.ilp \
  --in-static-file=qdb-static.json
```

#### Generate live metrics
You can run the app in live mode (`--live`) to let it continuously generating realtime metrics.
```shell
$ quest-ei \
  --interval=10s \
  --min-load=0.3 \
  --max-load=0.9 \
  --flush-batch-size=1000000 \
  --flush-batch-buffer-mb=500 \
  --in-static-file=qdb-static.json \
  --live
```

#### Help
```shell
$ quest-ei -h
Usage of ./quest-ei:
  -channels-per-site int
        Number of channels per site (default 10)
  -end string
        Ending time to generate metrics data (RFC3339) (default "2022-01-01T01:00:01Z")
  -fleets-per-site int
        Number of fleets per site (default 5)
  -flush-batch-buffer-mb int
        Number of MB memory will be used for buffering. Increase this value if flush-batch-size is too big (default 100)
  -flush-batch-size int
        Number of messages to flush to QuestDB in each batch. May need to increase flush-batch-buffer-mb if this value is too big. (default 10000)
  -in-static-file string
        Optional path to provide static JSON file. If this is set, no static records will be generated and only call metrics will be generated
  -interval string
        Interval duration for each loop when generating new metrics (default "10s")
  -live
        Generate the data in real time
  -max-load float
        Maximum load factor of a site. At each "interval", at most "maxLoadFactor" units will make a call (default 1)
  -min-load float
        Minimum load factor of a site. At each "interval", at least "minLoadFactor" units will make a call
  -out-metrics-file string
        Optional path to write ILP messages to the file instead of flushing to QuestDB directly
  -out-static-file string
        Optional path to write static data (sites, channels, fleets, talk groups, units) to JSON the file
  -sites int
        Number of sites (default 1)
  -start string
        Starting time to generate metrics data (RFC3339) (default "2022-01-01T00:00:00Z")
  -talk-groups-per-site int
        Number of talk groups per site (default 20)
  -units-per-talk-group int
        Number of unit per talk group (default 5)
```
