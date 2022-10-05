package model

import "time"

const (
	StatusActive Status = 1
)

type Status = int64

type Site struct {
	Id     string
	Name   string
	Status Status

	// Internal uses
	Channels   []*Channel
	Fleets     []*Fleet
	TalkGroups []*TalkGroup
	Units      []*Unit
}

type SiteReading struct {
	SiteId    string
	Status    Status
	Timestamp time.Time
}

type Channel struct {
	Id          string
	SiteId      string
	Name        string
	TxFrequency float64
	RxFrequency float64
	Status      Status
}

type Fleet struct {
	Id     string
	SiteId string
	Name   string
	Status Status
}

type TalkGroup struct {
	Id      string
	SiteId  string
	FleetId string
	Name    string
	Status  Status
}

type Unit struct {
	Id          string
	SiteId      string
	TalkGroupId string // 1 unit can be in multiple talkgroup?
	Name        string
	Status      Status
}

type Call struct {
	// Id     string
	SiteId string
	// FleetId string // FleetId of the sourceTalkGroup/Unit?
	SourceUnitId           string
	DestinationTalkGroupId string
	StartedAt              time.Time
	// EndedAt time.Time // Should track this or each call into 2 separated events
}
