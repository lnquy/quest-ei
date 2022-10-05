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
```shell
# Flush metric data to QuestDB running locally on the same host, 
# using default parameters.
$ quest-ei

# Write metric data to the file with customized parameters
$ quest-ei \
  --start=2022-01-01T00:00:00Z \
  --end=2022-01-10T00:00:01Z \
  --interval=10s \
  --load=0.6 \
  --sites=2 \
  --fleets-per-site=10 \
  --channels-per-site=10 \
  --talk-groups-per-site=50 \
  --units-per-talk-group=10 \
  --flush-batch-size=100000 \
  --to-file=qdb-data.ilp
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
  -flush-batch-size int
        Number of messages to flush to QuestDB in each batch (default 10000)
  -interval string
        Interval duration for each loop when generating new metrics (default "10s")
  -load float
        Load capacity of a site. At each "interval", how many "load" units will make a call [0.0-1.0] (default 0.5)
  -sites int
        Number of sites (default 1)
  -start string
        Starting time to generate metrics data (RFC3339) (default "2022-01-01T00:00:00Z")
  -talk-groups-per-site int
        Number of talk groups per site (default 20)
  -to-file string
        Optional path to write ILP messages to the file instead of flushing to QuestDB directly
  -units-per-talk-group int
        Number of unit per talk group (default 5)
```
