CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TABLE cities (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE stations (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    city_id    BIGINT NOT NULL REFERENCES cities(id) ON DELETE CASCADE,
    station_id TEXT NOT NULL,
    name       TEXT NOT NULL,
    lat        DOUBLE PRECISION NOT NULL,
    lon        DOUBLE PRECISION NOT NULL,
    address    TEXT NOT NULL DEFAULT '',
    capacity   INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (city_id, station_id)
);

CREATE TABLE station_status (
    time            TIMESTAMPTZ NOT NULL,
    station_pk      BIGINT NOT NULL REFERENCES stations(id) ON DELETE CASCADE,
    bikes_available INTEGER NOT NULL,
    docks_available INTEGER NOT NULL,
    bikes_disabled  INTEGER NOT NULL DEFAULT 0,
    docks_disabled  INTEGER NOT NULL DEFAULT 0,
    is_installed    BOOLEAN NOT NULL,
    is_renting      BOOLEAN NOT NULL,
    is_returning    BOOLEAN NOT NULL,
    last_reported   TIMESTAMPTZ
);

SELECT create_hypertable('station_status', by_range('time'));

CREATE INDEX station_status_station_pk_time_idx ON station_status (station_pk, time DESC);
