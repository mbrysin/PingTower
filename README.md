<div align="center">

# pingtower

**website uptime monitoring, built with go**

track website health, report incidents, and explore uptime history.

[![cov](https://mbrysin.github.io/PingTower/badges/test.svg)](https://github.com/mbrysin/PingTower/actions)

<img src="https://skillicons.dev/icons?i=go,postgres,redis,docker,githubactions,html,css,typescript&theme=dark" alt="technology stack" />

</div>

> [!NOTE]
> this project is under active development. some components and features are planned and may not be implemented yet.

## about

pingtower is a website monitoring service. visitors can view website statuses and incident history. registered users can report incidents.

## components

- [x] **auth** — user registration and login.
- [ ] **checker** — checks website availability and records results.
- [ ] **frontend** — user interface for statuses, charts, and incident reports.
- [ ] **gateway** — single entry point that routes requests to backend services.
- [ ] **monitor** — manages monitored websites and schedules checks.
- [ ] **stats** — stores and aggregates check history for statistics and charts.

## architecture

```text
                         ┌────────────┐
                         │  frontend  │
                         └─────┬──────┘
                               │
                         ┌─────▼──────┐
                         │  gateway   │
                         └─┬───┬───┬──┘
                           │   │   │
             ┌─────────────┘   │   └─────────────┐
             ▼                 ▼                 ▼
        ┌─────────┐      ┌──────────┐       ┌─────────┐
        │  auth   │      │ monitor  │       │  stats  │
        └─────────┘      └────┬─────┘       └────▲────┘
                              ▼                  │
                         ┌─────────┐             │
                         │ checker ├─────────────┘
                         └────┬────┘
                              │
                    ┌─────────┴─────────┐
                    ▼                   ▼
              ┌──────────┐        ┌─────────┐
              │postgresql│        │  redis  │
              └──────────┘        └─────────┘
```

## tech stack

- **languages:** go + sql + html + css + typescript
- **backend:** gin + rest api
- **database:** postgresql + goose migrations
- **cache and rate limiting:** redis
- **messaging:** kafka
- **security:** bcrypt
- **infrastructure:** docker + docker compose
- **testing and ci:** go test + github actions

## project structure

```text
pingtower/
├── services/
│   ├── auth/
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   └── ...
│   ├── checker/
│   │   ├── Dockerfile
│   │   └── ...
│   ├── gateway/
│   │   ├── Dockerfile
│   │   └── ...
│   ├── monitor/
│   │   ├── Dockerfile
│   │   └── ...
│   └── stats/
│       ├── Dockerfile
│       └── ...
├── frontend/
│   ├── Dockerfile
│   └── ...
├── .github/
│   └── workflows/
│       └── tests.yml
├── docker-compose.yaml
├── .env.example
├── go.mod
├── go.sum
└── README.md
```
