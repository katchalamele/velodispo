-- La prod a été configurée à la main avant cette migration : elle doit pouvoir se rejouer sans erreur.
ALTER TABLE station_status SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'station_pk',
    timescaledb.compress_orderby = 'time DESC'
);

SELECT add_compression_policy('station_status', INTERVAL '7 days', if_not_exists => true);

SELECT add_retention_policy('station_status', INTERVAL '6 months', if_not_exists => true);
