package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	fake "github.com/brianvoe/gofakeit/v6"
	"github.com/lnquy/quest-ei/pkg/model"
	qdb "github.com/questdb/go-questdb-client"
)

var (
	fStart                  string
	fEnd                    string
	fInterval               string
	fNoOfSites              int
	fNoOfChannelsPerSite    int
	fNoOfFleetsPerSite      int
	fNoOfTalkGroupsPerSites int
	fNoOfUnitsPerTalkGroup  int
	fFlushBatchSize         int
	fFlushBatchBufferMB     int
	fMinLoadFactor          float64
	fMaxLoadFactor          float64
	fOutMetricsFile         string
	fOutStaticFile          string
	fInStaticFile           string

	start    time.Time
	end      time.Time
	interval time.Duration
)

func init() {
	flag.StringVar(&fStart, "start", "2022-01-01T00:00:00Z", "Starting time to generate metrics data (RFC3339)")
	flag.StringVar(&fEnd, "end", "2022-01-01T01:00:01Z", "Ending time to generate metrics data (RFC3339)")
	flag.StringVar(&fInterval, "interval", "10s", "Interval duration for each loop when generating new metrics")
	flag.IntVar(&fNoOfSites, "sites", 1, "Number of sites")
	flag.IntVar(&fNoOfChannelsPerSite, "channels-per-site", 10, "Number of channels per site")
	flag.IntVar(&fNoOfFleetsPerSite, "fleets-per-site", 5, "Number of fleets per site")
	flag.IntVar(&fNoOfTalkGroupsPerSites, "talk-groups-per-site", 20, "Number of talk groups per site")
	flag.IntVar(&fNoOfUnitsPerTalkGroup, "units-per-talk-group", 5, "Number of unit per talk group")
	flag.IntVar(&fFlushBatchSize, "flush-batch-size", 10000, "Number of messages to flush to QuestDB in each batch. May need to increase flush-batch-buffer-mb if this value is too big.")
	flag.IntVar(&fFlushBatchBufferMB, "flush-batch-buffer-mb", 100, "Number of MB memory will be used for buffering. Increase this value if flush-batch-size is too big")
	flag.Float64Var(&fMinLoadFactor, "min-load", 0.0, `Minimum load factor of a site. At each "interval", at least "minLoadFactor" units will make a call`)
	flag.Float64Var(&fMaxLoadFactor, "max-load", 1.0, `Maximum load factor of a site. At each "interval", at most "maxLoadFactor" units will make a call`)
	flag.StringVar(&fOutMetricsFile, "out-metrics-file", "", "Optional path to write ILP messages to the file instead of flushing to QuestDB directly")
	flag.StringVar(&fOutStaticFile, "out-static-file", "", "Optional path to write static data (sites, channels, fleets, talk groups, units) to JSON the file")
	flag.StringVar(&fInStaticFile, "in-static-file", "", "Optional path to provide static JSON file. If this is set, no static records will be generated and only call metrics will be generated")

	flag.Parse()

	var err error
	start, err = time.Parse(time.RFC3339, fStart)
	panicIfError(err, "failed to parse start time from argument")
	end, err = time.Parse(time.RFC3339, fEnd)
	panicIfError(err, "failed to parse end time from argument")
	interval, err = time.ParseDuration(fInterval)
	panicIfError(err, "failed to parse interval from argument")
}

func main() {
	startedAt := time.Now()
	ctx := context.TODO()
	sender := newQuestDbILPSender(ctx)
	defer sender.Close()

	sites := make([]*model.Site, 0, fNoOfSites)

	// Init static data (sites, channels, fleets, talk groups, units)
	if fInStaticFile != "" { // Load from provided file
		log.Printf("Loading static records from JSON file: %s", fInStaticFile)
		b, err := ioutil.ReadFile(fInStaticFile)
		panicIfError(err, "failed to open static records JSON file")
		panicIfError(json.Unmarshal(b, &sites), "failed to decode static records JSON file")
		log.Printf("   + Sites: %d", len(sites))
		log.Printf("   + Channels (%d*%dsites): %d", len(sites[0].Channels), len(sites), len(sites[0].Channels)*len(sites))
		log.Printf("   + Fleets (%d*%dsites): %d", len(sites[0].Fleets), len(sites), len(sites[0].Fleets)*len(sites))
		log.Printf("   + TalkGroups (%d*%dsites): %d", len(sites[0].TalkGroups), len(sites), len(sites[0].TalkGroups)*len(sites))
		log.Printf("   + Units (%d*%dsites): %d", len(sites[0].Units), len(sites), len(sites[0].Units)*len(sites))
	} else { // or generate newly
		log.Printf("Generating static records from provided arguments")
		sites = generateStaticRecords(ctx, sender)
	}
	// Save static records to JSON file for later reuse, so we won't have to re-generate it again
	if fOutStaticFile != "" {
		b, err := json.Marshal(sites)
		panicIfError(err, "failed to encode sites to JSON")
		panicIfError(ioutil.WriteFile(fOutStaticFile, b, 0666), "failed to write static records JSON file")
		log.Printf("Static records JSON file written to: %s", fOutStaticFile)
	}

	// Init dynamic data (call metrics)
	log.Printf("Generating call metrics")
	sender = newQuestDbILPSender(ctx)
	generateCallMetrics(ctx, sender, sites)

	log.Printf("Finished in %s", time.Since(startedAt))
}

func newQuestDbILPSender(ctx context.Context) *qdb.LineSender {
	s, err := qdb.NewLineSender(ctx,
		qdb.WithBufferCapacity(fFlushBatchBufferMB*1024*1024),
	)
	panicIfError(err, "failed to init QuestDB line sender")
	return s
}

func generateStaticRecords(ctx context.Context, s *qdb.LineSender) []*model.Site {
	sites := make([]*model.Site, 0, fNoOfSites)

	// Generate static records (sites, channels, fleets, talk groups, units)
	for i := 0; i < fNoOfSites; i++ {
		siteId := fake.UUID()
		// Units of a site
		units := make([]*model.Unit, 0, fNoOfUnitsPerTalkGroup*fNoOfTalkGroupsPerSites)

		// Fleets of a site
		fleets := make([]*model.Fleet, 0, fNoOfFleetsPerSite)
		for j := 0; j < fNoOfFleetsPerSite; j++ {
			fleets = append(fleets, &model.Fleet{
				Id:     fake.UUID(),
				SiteId: siteId,
				Name:   "Fleet#" + fake.CountryAbr(),
				Status: model.StatusActive,
			})
		}

		// Channels of a site
		channels := make([]*model.Channel, 0, fNoOfChannelsPerSite)
		for j := 0; j < fNoOfChannelsPerSite; j++ {
			channels = append(channels, &model.Channel{
				Id:          fake.UUID(),
				SiteId:      siteId,
				Name:        fmt.Sprintf("Channel#%d", j),
				TxFrequency: fake.Float64(),
				RxFrequency: fake.Float64(),
				Status:      model.StatusActive,
			})
		}

		// TalkGroups of a site
		talkGroups := make([]*model.TalkGroup, 0, fNoOfTalkGroupsPerSites)
		for j := 0; j < fNoOfTalkGroupsPerSites; j++ {
			tgName := fake.Word()
			tgName = strings.ToUpper(string(tgName[0])) + tgName[1:]
			talkGroup := model.TalkGroup{
				Id:      fake.UUID(),
				SiteId:  siteId,
				FleetId: fleets[fake.IntRange(0, len(fleets)-1)].Id, // Randomly assign talk group to a fleet
				Name:    "TalkGroup#" + tgName,
				Status:  model.StatusActive,
			}

			// Units per talk group
			for k := 0; k < fNoOfUnitsPerTalkGroup; k++ {
				unitName := fake.LoremIpsumWord()
				unitName = strings.ToUpper(string(unitName[0])) + unitName[1:]
				units = append(units, &model.Unit{
					Id:          fake.UUID(),
					SiteId:      siteId,
					TalkGroupId: talkGroup.Id,
					Name:        "Unit#" + unitName,
					Status:      model.StatusActive,
				})
			}

			talkGroups = append(talkGroups, &talkGroup)
		}

		// Site
		sites = append(sites, &model.Site{
			Id:         siteId,
			Name:       "Site#" + fake.Fruit(),
			Status:     model.StatusActive,
			Channels:   channels,
			Fleets:     fleets,
			TalkGroups: talkGroups,
			Units:      units,
		})
	}

	// Flush static records to QuestDB or file
	ts := start.UnixNano()
	for _, site := range sites {
		log.Printf(" > Saving %q (%s) site", site.Name, site.Id)
		err := s.Table("sites").
			Symbol("id", site.Id).
			Symbol("name", site.Name).
			Int64Column("status", site.Status). // Active
			At(ctx, ts)
		panicIfError(err, "failed to save sites record")

		// err = s.Table("site_readings").
		// 	Symbol("site_id", siteId).
		// 	Int64Column("status", 1).
		// 	TimestampColumn("timestamp", now.UnixMicro()).
		// 	At(ctx, now.UnixNano())
		// panicIfError(err, "failed to save site_readings record")

		// Channel
		log.Printf("   + Saving %d channels", len(site.Channels))
		for _, channel := range site.Channels {
			err := s.Table("channels").
				Symbol("id", channel.Id).
				Symbol("site_id", channel.SiteId).
				Symbol("name", channel.Name).
				Float64Column("tx_freq", channel.TxFrequency).
				Float64Column("rx_freq", channel.RxFrequency).
				Int64Column("status", channel.Status).
				At(ctx, ts)
			panicIfError(err, "failed to save channels record")
		}

		// Fleet
		log.Printf("   + Saving %d fleets", len(site.Fleets))
		for _, fleet := range site.Fleets {
			err := s.Table("fleets").
				Symbol("id", fleet.Id).
				Symbol("site_id", fleet.SiteId).
				Symbol("name", fleet.Name).
				Int64Column("status", fleet.Status).
				At(ctx, ts)
			panicIfError(err, "failed to save fleets record")
		}

		// TalkGroup
		log.Printf("   + Saving %d talk groups", len(site.TalkGroups))
		for _, talkGroup := range site.TalkGroups {
			err := s.Table("talk_groups").
				Symbol("id", talkGroup.Id).
				Symbol("site_id", talkGroup.SiteId).
				Symbol("fleet_id", talkGroup.FleetId).
				Symbol("name", talkGroup.Name).
				Int64Column("status", talkGroup.Status).
				At(ctx, ts)
			panicIfError(err, "failed to save talk_groups record")
		}

		// Units
		log.Printf("   + Saving %d units", len(site.Units))
		for _, units := range site.Units {
			err := s.Table("units").
				Symbol("id", units.Id).
				Symbol("site_id", units.SiteId).
				Symbol("talk_group_id", units.TalkGroupId).
				Symbol("name", units.Name).
				Int64Column("status", units.Status).
				At(ctx, ts)
			panicIfError(err, "failed to save units record")
		}

		s = flushILPMessages(ctx, *s)
		log.Printf("   Saved %q site", site.Name)
	}

	return sites
}

func generateCallMetrics(ctx context.Context, s *qdb.LineSender, sites []*model.Site) {
	calls := make([]*model.Call, 0, fFlushBatchSize)
	totalCalls := 0
	for start.Before(end) {
		for _, site := range sites {
			var unit *model.Unit

			// For each "interval", only "loadFactor" units will make a call
			// This randomization simulates different load on each system at a time
			loadFactor := fake.Float64Range(fMinLoadFactor, fMaxLoadFactor)
			unitCalls := int(float64(len(site.Units)) * loadFactor)
			for j := 0; j < unitCalls; j++ {
				unit = site.Units[fake.IntRange(0, len(site.Units)-1)]                 // Randomly pick a unit
				talkGroup := site.TalkGroups[fake.IntRange(0, len(site.TalkGroups)-1)] // Randomly pick a talkGroup
				endedAt := fake.DateRange(start, start.Add(15*time.Minute))
				calls = append(calls, &model.Call{
					Id:                     fake.UUID(),
					SiteId:                 site.Id,
					ChannelId:              site.Channels[fake.IntRange(0, len(site.Channels)-1)].Id, // Randomly pick a channel
					FleetId:                talkGroup.FleetId,
					SourceUnitId:           unit.Id,
					DestinationTalkGroupId: talkGroup.Id,
					StartedAt:              start,
					EndedAt:                endedAt, // Randomize call duration, at most 15m
					DurationSecond:         int64(endedAt.Sub(start).Seconds()),
				})
			}

			if len(calls) <= fFlushBatchSize {
				continue
			}

			log.Printf(" > Flushing %d call metrics: start=%s, end=%s", len(calls), start.Format(time.RFC3339), end.Format(time.RFC3339))
			for _, c := range calls {
				err := s.Table("calls").
					Symbol("site_id", c.SiteId).
					Symbol("channel_id", c.ChannelId).
					Symbol("fleet_id", c.FleetId).
					Symbol("source_unit_id", c.SourceUnitId).
					Symbol("destination_talk_group_id", c.DestinationTalkGroupId).
					StringColumn("id", c.Id).
					TimestampColumn("started_at", c.StartedAt.UnixNano()).
					TimestampColumn("ended_at", c.EndedAt.UnixNano()).
					Int64Column("duration_sec", c.DurationSecond).
					At(ctx, c.StartedAt.UnixNano())
				panicIfError(err, "failed to save calls record")
			}
			s = flushILPMessages(ctx, *s)
			totalCalls += len(calls)
			log.Printf("   + %d call metrics saved, totalSaved=%d", len(calls), totalCalls)
			calls = make([]*model.Call, 0, fFlushBatchSize) // Reset batch
		}

		start = start.Add(interval) // Jump to the next interval
	}

	if len(calls) == 0 {
		return
	}
	// Last flush
	log.Printf(" > Flushing %d final call metrics", len(calls))
	for _, c := range calls {
		err := s.Table("calls").
			Symbol("site_id", c.SiteId).
			Symbol("channel_id", c.ChannelId).
			Symbol("fleet_id", c.FleetId).
			Symbol("source_unit_id", c.SourceUnitId).
			Symbol("destination_talk_group_id", c.DestinationTalkGroupId).
			StringColumn("id", c.Id).
			TimestampColumn("started_at", c.StartedAt.UnixNano()).
			TimestampColumn("ended_at", c.EndedAt.UnixNano()).
			Int64Column("duration_sec", c.DurationSecond).
			At(ctx, c.StartedAt.UnixNano())
		panicIfError(err, "failed to save calls record")
	}
	s = flushILPMessages(ctx, *s)
	totalCalls += len(calls)
	log.Printf("   + %d final call metrics saved, totalSaved=%d", len(calls), totalCalls)
}

func flushILPMessages(ctx context.Context, s qdb.LineSender) *qdb.LineSender {
	if fOutMetricsFile == "" {
		panicIfError(s.Flush(ctx), "failed to flush ILP messages to QuestDB")
		return &s
	}

	// Write ILP to file instead of flushing to QuestDB directly.
	// This file then can be used on `tsbs_load_questdb --file qdb-data.ilp --workers 4`
	f, err := os.OpenFile(fOutMetricsFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	panicIfError(err, "failed to open file to write")
	defer f.Close()
	_, err = f.WriteString(s.Messages())
	panicIfError(err, "failed to write ILP messages to file")
	_ = s.Close() // Close to remove all buffered messages first
	return newQuestDbILPSender(ctx)
}

func panicIfError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
