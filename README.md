<div align="center">

# PingTower

**website uptime monitoring, built with go**

---

track website health, report incidents, and explore uptime history.

[![Go tests](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml/badge.svg?branch=main)](https://github.com/mbrysin/PingTower/actions/workflows/go-test.yml)

[![Go](https://skillicons.dev/icons?i=go)](https://go.dev/) [![PostgreSQL](https://skillicons.dev/icons?i=postgres)](https://www.postgresql.org/) [![Redis](https://skillicons.dev/icons?i=redis)](https://redis.io/) [![Docker](https://skillicons.dev/icons?i=docker)](https://www.docker.com/) [![HTML](https://skillicons.dev/icons?i=html)](https://html.spec.whatwg.org/) [![CSS](https://skillicons.dev/icons?i=css)](https://www.w3.org/Style/CSS/) [![TypeScript](https://skillicons.dev/icons?i=ts)](https://www.typescriptlang.org/) [![JavaScript](https://skillicons.dev/icons?i=js)](https://developer.mozilla.org/en-US/docs/Web/JavaScript) [![TypeScript](https://skillicons.dev/icons?i=ts)](https://www.typescriptlang.org/)

</div>

> [!NOTE]
> this project is under active development. some components and features are planned and may not be implemented yet.

## about

pingtower is a website monitoring service. visitors can view website statuses and incident history. registered users can report incidents.

## tech stack

- **languages:** go + sql + html + css + typescript
- **backend:** gin + rest api
- **database:** postgresql + goose migrations
- **cache and rate limiting:** redis
- **security:** bcrypt
- **infrastructure:** docker + docker compose
- **testing and ci:** go test + github actions

## preview

![PingTower dashboard dark theme](docs/screenshots/dashboard_dark_theme.png)
![PingTower dashboard light theme](docs/screenshots/dashboard_light_theme.png)
![PingTower site dark theme](docs/screenshots/site_dark_theme.png)
![PingTower site light theme](docs/screenshots/site_light_theme.png)

## components

- [x] **sites** — manages monitored websites.
- [x] **checker** — probes enabled websites and records results and daily statistics.
- [x] **nginx** — reverse proxy and entry point for the services.
- [x] **stats** — serves historical statistics and chart data.
- [x] **frontend** — displays statuses, charts, and incident reports.
