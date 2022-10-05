CREATE TABLE 'sites' (
                         id SYMBOL CAPACITY 100 CACHE, -- At most 28 sites
                         name SYMBOL CAPACITY 100 CACHE,
                         status LONG,
                         timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;
ALTER TABLE sites ALTER COLUMN id ADD INDEX;
ALTER TABLE sites ALTER COLUMN name ADD INDEX;

CREATE TABLE 'channels' (
                            id SYMBOL CAPACITY 1000 CACHE, -- Assume normally 10 channels per site
                            site_id SYMBOL CAPACITY 100 CACHE,
                            name SYMBOL CAPACITY 1000 CACHE,
                            tx_freq DOUBLE,
                            rx_freq DOUBLE,
                            status LONG,
                            timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;
ALTER TABLE channels ALTER COLUMN id ADD INDEX;
ALTER TABLE channels ALTER COLUMN site_id ADD INDEX;
ALTER TABLE channels ALTER COLUMN name ADD INDEX;

CREATE TABLE 'fleets' (
                          id SYMBOL CAPACITY 5000 CACHE, -- Assume normally 50 fleets per site
                          site_id SYMBOL CAPACITY 100 CACHE,
                          name SYMBOL CAPACITY 5000 CACHE,
                          status LONG,
                          timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;
ALTER TABLE fleets ALTER COLUMN id ADD INDEX;
ALTER TABLE fleets ALTER COLUMN site_id ADD INDEX;
ALTER TABLE fleets ALTER COLUMN name ADD INDEX;

CREATE TABLE 'talk_groups' (
                               id SYMBOL CAPACITY 10000 CACHE, -- Assume 100 talk groups per site
                               site_id SYMBOL CAPACITY 100 CACHE,
                               fleet_id SYMBOL CAPACITY 5000 CACHE,
                               name SYMBOL CAPACITY 10000 CACHE,
                               status LONG,
                               timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;
ALTER TABLE talk_groups ALTER COLUMN id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN site_id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN fleet_id ADD INDEX;
ALTER TABLE talk_groups ALTER COLUMN name ADD INDEX;

CREATE TABLE 'units' (
                         id SYMBOL CAPACITY 50000 CACHE, -- Assume 500 units per site
                         site_id SYMBOL CAPACITY 100 CACHE,
                         talk_group_id SYMBOL CAPACITY 10000 CACHE,
                         name SYMBOL CAPACITY 50000 CACHE,
                         status LONG,
                         timestamp TIMESTAMP
) timestamp (timestamp) PARTITION BY DAY;
ALTER TABLE units ALTER COLUMN id ADD INDEX;
ALTER TABLE units ALTER COLUMN site_id ADD INDEX;
ALTER TABLE units ALTER COLUMN talk_group_id ADD INDEX;
ALTER TABLE units ALTER COLUMN name ADD INDEX;

CREATE TABLE 'calls' (
                         -- Purposely set this field as STRING, as SYMBOL causing ingestion overhead
                         -- and we dont want to search these individual records.
                         id STRING,
                         site_id SYMBOL CAPACITY 100 CACHE,
                         channel_id SYMBOL CAPACITY 10000 CACHE,
                         fleet_id SYMBOL CAPACITY 10000 CACHE,
                         source_unit_id SYMBOL CAPACITY 50000 CACHE,
                         destination_talk_group_id SYMBOL CAPACITY 10000 CACHE,
                         started_at TIMESTAMP,
                         ended_at TIMESTAMP,
                         duration_second LONG
) timestamp (started_at) PARTITION BY DAY;
-- ALTER TABLE calls ALTER COLUMN id ADD INDEX;
ALTER TABLE calls ALTER COLUMN site_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN channel_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN fleet_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN source_unit_id ADD INDEX;
ALTER TABLE calls ALTER COLUMN destination_talk_group_id ADD INDEX;
