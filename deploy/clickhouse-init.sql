CREATE DATABASE IF NOT EXISTS anticheat;

CREATE TABLE IF NOT EXISTS anticheat.telemetry_raw
(
    batch_id String,
    session_id String,
    player_id_hash String,
    received_at DateTime DEFAULT now(),
    payload_bytes String
)
ENGINE = MergeTree
ORDER BY (received_at, batch_id);
