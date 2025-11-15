```
AKTIV-HIKE-BOT/
├─ .env
├─ cmd/bot/
│  └─ main.go
├─ internal/
│  ├─ app/
│  │  └─ config/
│  │     └─ config.go
│  ├─ bot/
│  │  ├─ commands/
│  │  │  └─ commands.go        # BotCommandScope: клиент/админ
│  │  ├─ router/
│  │  │  └─ router.go          # маршрутизация апдейтов
│  │  ├─ notify/
│  │  │  └─ notify.go          # отправка в ADMIN_CHAT_ID
│  │  ├─ admin/
│  │  │  ├─ handler.go         # меню/списки/карточка/тоггл publish (минимум)
│  │  │  └─ fsm.go             # каркас FSM "создать хайк"
│  │  └─ client/
│  │     └─ handler.go         # заглушка клиентских сценариев
│  ├─ db/
│  │  ├─ queries/
│  │  │  └─ admin/
│  │  │     └─ hikes.sql
│  │  └─ sqlc/                 # сгенерирует sqlc
└─ sqlc.yaml
```