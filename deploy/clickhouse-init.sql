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

CREATE TABLE IF NOT EXISTS anticheat.telemetry_events
(
  received_at       DateTime,
  batch_id          String,
  session_id        String,
  player_id_hash    String,

  fps_avg           Float32,
  ping_ms_avg       Float32,

  speed_avg_mps     Float32,
  speed_max_mps     Float32,
  teleport_suspect  UInt8,
  flight_suspect    UInt8,
  noclip_suspect    UInt8,

  anomaly_score     Float32
)
ENGINE = MergeTree
ORDER BY (received_at, session_id);


CREATE TABLE IF NOT EXISTS anticheat.incidents
(
  ts               DateTime,
  session_id       String,
  player_id_hash   String,
  rule             String,
  value            Float32,
  details          String
)
ENGINE = MergeTree
ORDER BY (ts, session_id);

