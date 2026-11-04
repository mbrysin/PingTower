<div align="center">

[![PingTowerLogo](/images/PingTowerDarkBg.svg)](https://github.com/mbrysin/PingTower)

**website uptime monitoring, built with go**

---

track website health, report incidents, and explore uptime history.

[![Go tests](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml/badge.svg?branch=main)](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml)

[![Go](https://skillicons.dev/icons?i=go)](https://go.dev/) [![PostgreSQL](https://skillicons.dev/icons?i=postgres)](https://www.postgresql.org/) [![Redis](https://skillicons.dev/icons?i=redis)](https://redis.io/) [![Docker](https://skillicons.dev/icons?i=docker)](https://www.docker.com/) [![HTML](https://skillicons.dev/icons?i=html)](https://html.spec.whatwg.org/) [![CSS](https://skillicons.dev/icons?i=css)](https://www.w3.org/Style/CSS/) 

</div>

> [!NOTE]
> PingTower is a learning project under active development. some components and features are planned and may not be implemented yet.

## about

PingTower is a website monitoring service. visitors can view website statuses and incident history. registered users can report incidents.

## tech stack

- **languages:** go + sql + html + css + typescript
- **backend:** gin + rest api
- **database:** postgresql + goose migrations
- **cache and rate limiting:** redis
- **security:** bcrypt
- **infrastructure:** docker + docker compose
- **testing and ci:** go tests + github actions

## api

| method | endpoint | description |
| --- | --- | --- |
| `get` | `/sites` | list sites |
| `get` | `/sites/:id` | get a site by ID |

## components

- [x] **sites** — provides api endpoints for listing and retrieving sites.
- [x] **checker** — probes enabled websites, records results, and stores statistics in the database.
- [x] **nginx** — reverse proxy and entry point for the services.
- [ ] **frontend** — displays statuses, charts, and incident reports.
- [ ] **stats** — serves historical statistics and chart data.

## run locally

start the project with docker compose:
```bash
docker compose up --build
