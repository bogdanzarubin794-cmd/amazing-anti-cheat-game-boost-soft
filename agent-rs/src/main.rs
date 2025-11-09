mod pb {
    include!("pb/anticheat.v1.rs");
}

use anyhow::Result;
use chrono::Utc;
use pb::*;
use reqwest::Client;
use std::time::Duration;
use uuid::Uuid;

#[derive(serde::Deserialize)]
struct Config { agent: AgentCfg }
#[derive(serde::Deserialize)]
struct AgentCfg { server_url: String, send_interval_secs: u64 }

#[tokio::main]
async fn main() -> Result<()> {
    let cfg: Config = {
        let txt = std::fs::read_to_string("config.toml")
            .expect("put config.toml near executable");
        toml::from_str(&txt)?
    };
    let client = Client::new();
    let session_id = Uuid::new_v4().to_string();
    let player_id_hash = format!("{:x}", md5::compute("demo_player"));

    loop {
        let batch = make_mock_batch(&session_id, &player_id_hash);
        let mut buf = Vec::with_capacity(1024);
        prost::Message::encode(&batch, &mut buf).unwrap();

        let resp = client
            .post(&cfg.agent.server_url)
            .header("Content-Type", "application/x-protobuf")
            .header("X-Batch-Id", &batch.batch_id)
            .header("X-Session-Id", &batch.session_id)
            .header("X-Player-Id", &batch.player_id_hash)
            .body(buf)
            .send()
            .await?;
        if !resp.status().is_success() {
            eprintln!("ingest error: {}", resp.status());
        }
        tokio::time::sleep(Duration::from_secs(cfg.agent.send_interval_secs)).await;
    }
}

fn make_mock_batch(session_id: &str, player_hash: &str) -> TelemetryBatch {
    let now_ms = Utc::now().timestamp_millis();

    let perf = PerfStats {
        fps_avg: 120.0, fps_p1: 90.0, fps_p99: 144.0,
        ping_ms_avg: 32.0, cpu_proc_percent: 12.0, mem_proc_mb: 512.0,
        gpu_usage_percent: 40.0, gpu_temp_c: 65.0,
    };
    let input = InputSummary {
        clicks_total: 23,
        buckets: vec![InputBucket { bucket_ms: 100, press_count: 12 }],
        click_variance: 0.18, triggerbot_suspect: false, recoil_perfect: false,
    };
    let move_sum = MovementSummary {
        distance_m: 30.0, speed_avg_mps: 4.2, speed_max_mps: 7.8,
        teleport_suspect: false, flight_suspect: false, noclip_suspect: false,
        vertical_speed_mps: 0.2,
    };
    let client_info = ClientInfo {
        agent_version: "0.1.0".into(), game_build: "dev".into(),
        os: std::env::consts::OS.into(), arch: std::env::consts::ARCH.into(),
    };
    let trig = RuleTriggers {
        speedhack:false, teleport:false, flight:false, noclip:false,
        aimbot:false, silent_aim:false, recoil_control:false, triggerbot:false, perfect_accuracy:false,
        impossible_tx:false, money_injection:false, item_dupe:false,
        report_count:0, honeypot_hit:false, unknown_module:false, packet_spoof:false,
        macro_pattern:false, auto_farm:false, no_spread:false, anomaly_score:0.0,
    };
    let event = TelemetryEvent {
        ts_unix_ms: now_ms,
        perf: Some(perf), input: Some(input), move_: Some(move_sum),
        modules: vec![],
        integrity: Some(IntegrityFlags { hash_mismatch:false, unknown_modules:false, hook_suspect:false, memory_edit:false }),
        triggers: Some(trig),
        forensic: Some(ForensicRef { has_snapshot:false, snapshot_id:"".into(), duration_sec:0 }),
        tags: [("map".into(),"dev".into())].into(),
    };
    TelemetryBatch {
        batch_id: Uuid::new_v4().to_string(),
        session_id: session_id.into(),
        player_id_hash: player_hash.into(),
        client: Some(client_info),
        events: vec![event],
        sent_at_unix_ms: now_ms,
    }
}