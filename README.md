<div align="center">

[![PingTowerLogo](/images/PingTowerDarkBg.svg)](https://github.com/mbrysin/PingTower)

**website uptime monitoring, built with go**

---

track website health, report incidents, and explore uptime history.

[![Go tests](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml/badge.svg?branch=main)](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml)

[![Go](https://skillicons.dev/icons?i=go)](https://go.dev/) [![PostgreSQL](https://skillicons.dev/icons?i=postgres)](https://www.postgresql.org/) [![Redis](https://skillicons.dev/icons?i=redis)](https://redis.io/) [![Docker](https://skillicons.dev/icons?i=docker)](https://www.docker.com/) [![HTML](https://skillicons.dev/icons?i=html)](https://html.spec.whatwg.org/) [![CSS](https://skillicons.dev/icons?i=css)](https://www.w3.org/Style/CSS/) 

</div>

> [!NOTE]
> this project is under active development. some components and features are planned and may not be implemented yet.

## about

pingtower is a website monitoring service. visitors can view website statuses and incident history. registered users can report incidents.

## components

- [x] **auth** — user registration and login.
- [x] **sites** — manages monitored websites.
- [x] **checker** — probes enabled websites and records results and daily statistics.
- [ ] **frontend** — displays statuses, charts, and incident reports.
- [ ] **gateway** — provides a single entry point and rate limiting.
- [ ] **stats** — serves historical statistics and chart data.

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
        │  auth   │      │  sites   │       │  stats  │
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
│   │   ├── handler/
│   │   ├── middleware/
│   │   ├── model/
│   │   ├── repo/
│   │   ├── service/
│   │   ├── token/
│   │   ├── .env
│   │   ├── .env.example
│   │   └── main.go
│   ├── checker/
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   ├── model/
│   │   ├── repo/
│   │   ├── checker/
│   │   ├── cache/
│   │   └── main.go
│   ├── gateway/
│   │   ├── Dockerfile
│   │   └── ...
│   ├── sites/
│   │   ├── migrations/
│   │   ├── Dockerfile
│   │   ├── handler/
│   │   ├── model/
│   │   ├── repo/
│   │   ├── service/
│   │   ├── .env
│   │   ├── .env.example
│   │   └── main.go
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
