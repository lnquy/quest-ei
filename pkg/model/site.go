package model

import "time"

const (
	StatusActive Status = 1
)

type Status = int64

// Site
/*
CREATE TABLE 'sites' (
                         id SYMBOL capacity 36 CACHE,
                         name SYMBOL capacity 256 CACHE,
                         status LONG,
                         timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE sites ALTER COLUMN id ADD INDEX;
ALTER TABLE sites ALTER COLUMN name ADD INDEX;
*/
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

// SiteReading
/*
CREATE TABLE 'site_readings' (
  site_id SYMBOL capacity 36 CACHE,
  status LONG,
  timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE site_readings ALTER COLUMN site_id ADD INDEX;
*/
type SiteReading struct {
	SiteId    string
	Status    Status
	Timestamp time.Time
}

// Channel
/*
CREATE TABLE 'channels' (
  id SYMBOL capacity 36 CACHE,
  site_id SYMBOL capacity 36 CACHE,
  name SYMBOL capacity 256 CACHE,
  tx_freq DOUBLE,
  rx_freq DOUBLE,
  status LONG,
  timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE channels ALTER COLUMN id ADD INDEX;
ALTER TABLE channels ALTER COLUMN site_id ADD INDEX;
ALTER TABLE channels ALTER COLUMN name ADD INDEX;
*/
type Channel struct {
	Id          string
	SiteId      string
	Name        string
	TxFrequency float64
	RxFrequency float64
	Status      Status
}

// Fleet
/*
CREATE TABLE 'fleets' (
  id SYMBOL capacity 36 CACHE,
  site_id SYMBOL capacity 36 CACHE,
  name SYMBOL capacity 256 CACHE,
  status LONG,
  timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE fleets ALTER COLUMN id ADD INDEX;
ALTER TABLE fleets ALTER COLUMN site_id ADD INDEX;
ALTER TABLE fleets ALTER COLUMN name ADD INDEX;
*/
type Fleet struct {
	Id     string
	SiteId string
	Name   string
	Status Status
}

// TalkGroup
/*
CREATE TABLE 'talk_groups' (
  id SYMBOL capacity 36 CACHE,
  site_id SYMBOL capacity 36 CACHE,
  fleet_id SYMBOL capacity 36 CACHE,
  name SYMBOL capacity 256 CACHE,
  status LONG,
  timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE talk_groups ALTER COLUMN id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN site_id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN fleet_id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN name ADD INDEX;
*/
type TalkGroup struct {
	Id      string
	SiteId  string
	FleetId string
	Name    string
	Status  Status
}

// Unit
/*
CREATE TABLE 'units' (
  id SYMBOL capacity 36 CACHE,
  site_id SYMBOL capacity 36 CACHE,
  talk_group_id SYMBOL capacity 36 CACHE,
  name SYMBOL capacity 256 CACHE,
  status LONG,
  timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;

ALTER TABLE units ALTER COLUMN id ADD INDEX;
ALTER TABLE units ALTER COLUMN site_id ADD INDEX;
ALTER TABLE units ALTER COLUMN talk_group_id ADD INDEX;
ALTER TABLE units ALTER COLUMN name ADD INDEX;
*/
type Unit struct {
	Id          string
	SiteId      string
	TalkGroupId string // 1 unit can be in multiple talkgroup?
	Name        string
	Status      Status
}

// Call
/*
CREATE TABLE 'calls' (
  id SYMBOL capacity 36 CACHE,
  site_id SYMBOL capacity 36 CACHE,
  source_unit_id SYMBOL capacity 36 CACHE,
  destination_talk_group_id SYMBOL capacity 256 CACHE,
  started_at TIMESTAMP
) timestamp (started_at) PARTITION BY DAY;

ALTER TABLE calls ALTER COLUMN id ADD INDEX;
ALTER TABLE calls ALTER COLUMN site_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN source_unit_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN destination_talk_group_id ADD INDEX;
*/
type Call struct {
	Id     string
	SiteId string
	// FleetId string // FleetId of the sourceTalkGroup/Unit?
	SourceUnitId           string
	DestinationTalkGroupId string
	StartedAt              time.Time
	// EndedAt time.Time // Should track this or each call into 2 separated events
}
