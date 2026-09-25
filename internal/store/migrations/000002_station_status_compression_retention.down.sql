SELECT remove_retention_policy('station_status', if_exists => true);

SELECT remove_compression_policy('station_status', if_exists => true);

SELECT decompress_chunk(c, if_compressed => true) FROM show_chunks('station_status') c;

ALTER TABLE station_status SET (timescaledb.compress = false);
