package model

import "time"

const (
	StatusActive Status = 1
)

type Status = int64

type Site struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Status Status `json:"status"`

	// Internal uses
	Channels   []*Channel   `json:"channels,omitempty"`
	Fleets     []*Fleet     `json:"fleets,omitempty"`
	TalkGroups []*TalkGroup `json:"talkGroups,omitempty"`
	Units      []*Unit      `json:"units,omitempty"`
}

// type SiteReading struct {
// 	SiteId    string `json:"site_id"`
// 	Status    Status `json:"status"`
// 	Timestamp time.Time `json:"timestamp"`
// }

type Channel struct {
	Id          string  `json:"id"`
	SiteId      string  `json:"siteId"`
	Name        string  `json:"name"`
	TxFrequency float64 `json:"txFrequency"`
	RxFrequency float64 `json:"rxFrequency"`
	Status      Status  `json:"status"`
}

type Fleet struct {
	Id     string `json:"id"`
	SiteId string `json:"siteId"`
	Name   string `json:"name"`
	Status Status `json:"status"`
}

type TalkGroup struct {
	Id      string `json:"id"`
	SiteId  string `json:"siteId"`
	FleetId string `json:"fleetId"`
	Name    string `json:"name"`
	Status  Status `json:"status"`
}

type Unit struct {
	Id          string `json:"id"`
	SiteId      string `json:"siteId"`
	TalkGroupId string `json:"talkGroupId"` // 1 unit can be in multiple talkgroup?
	Name        string `json:"name"`
	Status      Status `json:"status"`
}

type Call struct {
	Id                     string    `json:"id"`
	SiteId                 string    `json:"siteId"`
	ChannelId              string    `json:"channelId"`
	FleetId                string    `json:"fleetId"` // FleetId of the sourceTalkGroup/Unit?
	SourceUnitId           string    `json:"sourceUnitId"`
	DestinationTalkGroupId string    `json:"destinationTalkGroupId"`
	StartedAt              time.Time `json:"startedAt"`
	EndedAt                time.Time `json:"endedAt"` // Should track this or each call into 2 separated events
	DurationSecond         int64     `json:"durationSecond"`
}
