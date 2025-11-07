## README Структура репозитория !!!! ЭТО ВСЁ ДЛЯ УСЛОВНОЙ "ПРЕЗЕНТАЦИИ", НЕ ДЛЯ ПРОДА !!!!!
```
/agent-rs/          
/ingest-api/         
/rule-engine/        
/admin-ui/           
/proto/              
/deploy/            
/README.md          
```

### Сборка и запуск (локально) !!!! КОНТЕЙНЕР ПРИ НЕОБХОДИМОСТИ СДЕЛАЮ САМ !!!!!
**Требования:** Docker, Docker Compose, Rust toolchain, Go .

```bash
cd agent-rs
cargo build --release

cd ../deploy
docker compose up -d 

./target/release/agent-rs --server https://localhost:8443 --interval 10s
```

### (curl)
```bash
curl -X POST https://localhost:8443/api/v1/ingest \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @sample_batch.bin
```

### Конфигурация агента !!!! ПРИМЕР !!!!!
```toml
[agent]
server_url = "https://localhost:8443/api/v1/ingest"
send_interval_secs = 10
snapshot_duration_secs = 15
[game_boost]
enabled = true   # только при согласии игрока
```

### Набор базовых правил
- **speedhack**: скорость > порога по классу/транспорту;
- **impossible_tx**: выдача валюты/предмета без события‑источника;
- **honeypot_hit**: доступ к скрытому объекту;
- **report_density**: много репортов за короткое время;
- **unknown_module**: модуль вне белого списка.
- **speedhack**: скорость выше допустимой.
- **teleport**: мгновенное перемещение.
- **noclip / movement_no_clip**: проход через стены, объекты.
- **air_flight / антигравитация**: полёт без разрешённой механики.
- **vertical_speed_hack**: неестественно быстрый подъём вверх.
- **rapid_position_shift**: лаг-чит / fake lag switch (телепорты).
- **aimbot**: автоматическое наведение на игроков.
- **silent_aim**: попадания без видимого наведения.
- **recoil_control (anti-recoil)**: отсутствие отдачи оружия.
- **triggerbot**: авто-выстрел при наведении.
- **perfect_accuracy / no_spread**: нулевая разбросанность пуль.
- **fast_shooting / rapid_fire**: скорость стрельбы выше нормы.
- **impossible_kill_time**: реакция < человеческой (kill мгновенно).
- **impossible_hit_distance**: попадание с расстояния вне логики.
- **aim_smooth_bot_detection**: слишком ровное/идеальное наведение.
- **high_hitrate_on_moving_target**: высокий % попаданий по движущимся.
- **unreal_kill_combo**: несколько headshot'ов за <1–2 секунды.
- **impossible_tx**: деньги/лут без события (работы, миссии).
- **money_injection / economy_cheat**: резкое появление денег.
- **dupe_item / item_duplication**: дюп вещей.
- **teleport_item_transfer / mule_accounts**: передача предметов через фейковые аккаунты.
- **auto_farm / bot_behavior**: цикличные автоматические действия.
- **regular_macro_pattern**: повторяющиеся клики с одинаковым интервалом.
- **time_played_24_7**: игрок не выходит из игры 24/7.
- **bot_lua_script_detected**: присутствие внешнего скрипта (Lua и др.).
- **unknown_module**: загружена подозрительная DLL.
- **hash_mismatch**: изменён исполняемый файл/ресурсы игры.
- **hook_suspect (IAT/inline hook)**: вмешательство в функции игры.
- **packet_spoof / fake_client_packets**: подделка сетевых пакетов.
- **honeypot_hit**: игрок взаимодействует с объектом-ловушкой (видимым только для чита).
- **report_density**: слишком много жалоб от игроков.
- **multi_account_suspicion**: несколько аккаунтов с одного HWID/IP.
- **teleport_item_transfer / mule_accounts**: передача ресурсов между своими акками.
- **aim_smooth_bot_detection**: идеально плавное наведение (не человеческое).
- **movement_no_clip**: пересечение стен.
- **rapid_position_shift**: фейковый пинг для багов.
- **high_hitrate_on_moving_target**: аномально точные выстрелы по движущимся целям.

### Roadmap после пилота
- Разметка кейсов в админ‑панели → экспорт в датасет;
- ML‑классификатор поведения в observe‑режиме;
- Auto‑action для высоко‑точных кейсов, конфигурируемый.

