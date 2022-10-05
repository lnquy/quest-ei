package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	fake "github.com/brianvoe/gofakeit/v6"
	"github.com/lnquy/quest-ei/pkg/model"
	qdb "github.com/questdb/go-questdb-client"
)

const (
	senderBufferSizeBytes = 100 * 1024 * 1024 // 100MB
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
	fLoadCapacity           float64
	fToFile                 string

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
	flag.IntVar(&fFlushBatchSize, "flush-batch-size", 10000, "Number of messages to flush to QuestDB in each batch (max=100000)")
	flag.Float64Var(&fLoadCapacity, "load", 0.5, `Load capacity of a site. At each "interval", how many "load" units will make a call [0.0-1.0]`)
	flag.StringVar(&fToFile, "to-file", "", "Optional path to write ILP messages to the file instead of flushing to QuestDB directly")

	flag.Parse()

	var err error
	start, err = time.Parse(time.RFC3339, fStart)
	panicIfError(err, "failed to parse start time from argument")
	end, err = time.Parse(time.RFC3339, fEnd)
	panicIfError(err, "failed to parse end time from argument")
	interval, err = time.ParseDuration(fInterval)
	panicIfError(err, "failed to parse interval from argument")
	if fFlushBatchSize <= 0 || fFlushBatchSize > 100000 {
		panic("invalid flush-batch-size argument")
	}
}

func main() {
	startedAt := time.Now()
	ctx := context.TODO()
	sender := newQuestDbILPSender(ctx, senderBufferSizeBytes)
	defer sender.Close()

	sites := make([]*model.Site, 0, fNoOfSites)
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
			talkGroup := model.TalkGroup{
				Id:      fake.UUID(),
				SiteId:  siteId,
				FleetId: fleets[fake.IntRange(0, len(fleets)-1)].Id, // Randomly assign talk group to a fleet
				Name:    "TalkGroup#" + fake.Word(),
				Status:  model.StatusActive,
			}

			// Units per talk group
			for k := 0; k < fNoOfUnitsPerTalkGroup; k++ {
				units = append(units, &model.Unit{
					Id:          fake.UUID(),
					SiteId:      siteId,
					TalkGroupId: talkGroup.Id,
					Name:        "Unit#" + fake.SafeColor(),
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

	initStaticRecords(ctx, sender, sites)

	// Init call metrics
	log.Printf("Start generating call metrics")
	generateCallMetrics(ctx, sender, sites)

	log.Printf(" > Finished in %s", time.Since(startedAt))
}

func newQuestDbILPSender(ctx context.Context, bufferSizeBytes int) *qdb.LineSender {
	s, err := qdb.NewLineSender(ctx,
		qdb.WithBufferCapacity(bufferSizeBytes),
	)
	panicIfError(err, "failed to init QuestDB line sender")
	return s
}

func initStaticRecords(ctx context.Context, s *qdb.LineSender, sites []*model.Site) {
	nowNs := time.Now().UnixNano()
	for _, site := range sites {
		log.Printf(" > Saving %q (%s) site", site.Name, site.Id)
		err := s.Table("sites").
			Symbol("id", site.Id).
			Symbol("name", site.Name).
			Int64Column("status", site.Status). // Active
			At(ctx, nowNs)
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
				At(ctx, nowNs)
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
				At(ctx, nowNs)
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
				At(ctx, nowNs)
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
				At(ctx, nowNs)
			panicIfError(err, "failed to save units record")
		}

		s = flushILPMessages(ctx, s)
		log.Printf(" Saved %q site", site.Name)
	}
}

func generateCallMetrics(ctx context.Context, s *qdb.LineSender, sites []*model.Site) {
	calls := make([]*model.Call, 0, fFlushBatchSize)
	totalCalls := 0
	for start.Before(end) {
		for _, site := range sites {
			var unit *model.Unit

			// For each "interval", only "loadCapacity" units will make a call
			// riggedLoad simulates relative loads around the fLoadCapacity (+-10%)
			riggedLoad := fLoadCapacity + fake.Float64Range(-0.1, 0.1)
			if riggedLoad < 0.0 {
				riggedLoad = 0.0
			}
			if riggedLoad > 1 {
				riggedLoad = 1.0
			}
			unitCalls := int(float64(len(site.Units)) * riggedLoad)
			for j := 0; j < unitCalls; j++ {
				unit = site.Units[fake.IntRange(0, len(site.Units)-1)] // Randomly pick a unit
				calls = append(calls, &model.Call{
					Id:                     fake.UUID(),
					SiteId:                 site.Id,
					SourceUnitId:           unit.Id,
					DestinationTalkGroupId: site.TalkGroups[fake.IntRange(0, len(site.TalkGroups)-1)].Id, // Randomly call to a talkGroup
					StartedAt:              start,
				})
			}

			if len(calls) <= fFlushBatchSize {
				continue
			}

			log.Printf(" > Flushing %d call metrics: start=%s, end=%s", len(calls), start.Format(time.RFC3339), end.Format(time.RFC3339))
			for _, c := range calls {
				err := s.Table("calls").
					Symbol("id", c.Id).
					Symbol("site_id", c.SiteId).
					Symbol("source_unit_id", c.SourceUnitId).
					Symbol("destination_talk_group_id", c.DestinationTalkGroupId).
					TimestampColumn("started_at", start.UnixNano()).
					At(ctx, start.UnixNano())
				panicIfError(err, "failed to save calls record")
			}
			s = flushILPMessages(ctx, s)
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
			Symbol("id", c.Id).
			Symbol("site_id", c.SiteId).
			Symbol("source_unit_id", c.SourceUnitId).
			Symbol("destination_talk_group_id", c.DestinationTalkGroupId).
			TimestampColumn("started_at", c.StartedAt.UnixNano()).
			At(ctx, c.StartedAt.UnixNano())
		panicIfError(err, "failed to save calls record")
	}
	s = flushILPMessages(ctx, s)
	totalCalls += len(calls)
	log.Printf("   + %d final call metrics saved, totalSaved=%d", len(calls), totalCalls)
}

func flushILPMessages(ctx context.Context, s *qdb.LineSender) *qdb.LineSender {
	if fToFile == "" {
		panicIfError(s.Flush(ctx), "failed to flush ILP messages to QuestDB")
		return s
	}

	// Write ILP to file instead of flushing to QuestDB directly.
	// This file then can be used on `tsbs_load_questdb --file qdb-data.ilp --workers 4`
	f, err := os.OpenFile(fToFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	panicIfError(err, "failed to open file to write")
	defer f.Close()
	_, err = f.WriteString(s.Messages())
	panicIfError(err, "failed to write ILP messages to file")
	_ = s.Close() // Close to remove all buffered messages first
	return newQuestDbILPSender(ctx, senderBufferSizeBytes)
}

func panicIfError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
