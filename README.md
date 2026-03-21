# UMG Bus App 🚌

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19.2.4-61DAFB?style=for-the-badge&logo=react&logoColor=black)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-8-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docs.docker.com/compose/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**Real-time university transport tracking system**
_GSoC Proposal · Universidad Mesoamericana de Guatemala (UMG)_

</div>

---

## Project Overview / Descripción

UMG Bus App is a real-time public transport tracking system designed for university environments. It allows students to monitor the exact location of campus buses on a live map, and enables bus pilots (drivers) to broadcast their GPS position from their mobile devices.

This project is presented as a **Google Summer of Code (GSoC) proposal**, showcasing a practical, scalable implementation applied to the **Universidad Mesoamericana de Guatemala (UMG)** as its pilot university.

> **Estado actual:** Propuesta funcional con backend completo, WebSocket en tiempo real, migraciones de base de datos y frontend base en React.

---

## Objectives / Objetivos

- **GSoC Proposal** — Demonstrate a complete, production-ready open-source transport tracking system suitable for adoption by any university.
- **Public transport control** — Give university administrators full visibility and control over their transport fleet.
- **University pilot implementation** — Deployed as a proof-of-concept at UMG (Universidad Mesoamericana de Guatemala), the first institution to adopt this system.
- **Student accessibility** — Allow students to see, in real time, where their bus is — reducing wait times and improving campus mobility.
- **Pilot autonomy** — Drivers use their own phones as GPS devices; no additional hardware required.

---

## Tech Stack / Tecnologías

### Backend

| Technology                                                  | Version | Purpose                                                   |
| ----------------------------------------------------------- | ------- | --------------------------------------------------------- |
| [Go](https://go.dev/)                                       | 1.25    | Primary backend language                                  |
| [Gofiber v2](https://gofiber.io/)                           | v2      | HTTP REST API framework                                   |
| [PostgreSQL](https://www.postgresql.org/)                   | 15      | Relational database (campus, pilots, students, sessions)  |
| [pgx/v5](https://github.com/jackc/pgx)                      | v5      | PostgreSQL driver for Go                                  |
| [Redis](https://redis.io/)                                  | 8       | Real-time state cache + Pub/Sub for location broadcasting |
| [coder/websocket](https://github.com/coder/websocket)       | latest  | WebSocket protocol (pilot GPS streaming)                  |
| [golang-migrate](https://github.com/golang-migrate/migrate) | latest  | Database schema migrations                                |

### Frontend

| Technology                  | Version | Purpose                   |
| --------------------------- | ------- | ------------------------- |
| [React](https://react.dev/) | 19.2.4  | UI framework              |
| [Vite](https://vitejs.dev/) | 8.0.0   | Build tool and dev server |

### Infrastructure

| Technology                                         | Purpose                                                       |
| -------------------------------------------------- | ------------------------------------------------------------- |
| [Docker Compose](https://docs.docker.com/compose/) | Orchestrates PostgreSQL (port `5436`) and Redis (port `6381`) |

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                          Clients                                │
│                                                                  │
│   📱 Pilot App (mobile)          🖥️  Student Browser            │
│   WebSocket → /ws/pilot          HTTP + WebSocket               │
└──────────────┬──────────────────────────┬───────────────────────┘
               │                          │
               ▼                          ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go Backend                               │
│                                                                  │
│  ┌──────────────────┐    ┌──────────────────────────────────┐   │
│  │  Gofiber (REST)  │    │  net/http WebSocket Server       │   │
│  │  Port :8084      │    │  Port :8085                      │   │
│  │                  │    │                                  │   │
│  │  GET /health     │    │  /ws/pilot  → HandlePilot()      │   │
│  │  GET /campus     │    │               ↓                  │   │
│  │  GET /campus/:id │    │           Hub.PublishLocation()  │   │
│  │  GET /campus/:id/│    │               ↓                  │   │
│  │       pilot      │    │        Redis Pub/Sub             │   │
│  └──────────────────┘    └──────────────────────────────────┘   │
└────────────────┬────────────────────────────────────────────────┘
                 │
       ┌─────────▼─────────┐    ┌───────────────────┐
       │   PostgreSQL 15   │    │     Redis 8        │
       │   Port 5436       │    │     Port 6381      │
       │                   │    │                    │
       │  campus           │    │  GPS state (TTL)   │
       │  pilots           │    │  Pub/Sub channels  │
       │  students         │    │                    │
       │  sessions         │    └───────────────────┘
       │  campus_changes   │
       └───────────────────┘
```

**Data flow:**

1. The pilot opens a WebSocket to `/ws/pilot?pilot_id=X&campus_id=Y`
2. The pilot's app streams GPS pings (`lat`, `lng`, timestamp) continuously
3. The Hub saves current state in Redis (with TTL) and publishes to a Pub/Sub channel
4. Students receive real-time bus position updates via the Hub subscriber

---

## Database Schema

The PostgreSQL schema (3NF, fully relational) includes:

- **`campus`** — University campuses with geographic bounding boxes and optional route GeoJSON
- **`pilots`** — Authorized drivers per campus with secure secret-code onboarding
- **`students`** — Student users linked to a campus
- **`student_campus_changes`** — Immutable audit log of campus transfers
- **`sessions`** — Session audit log (active tokens live in Redis with TTL)

---

## API Endpoints

| Method | Path                             | Description                                 |
| ------ | -------------------------------- | ------------------------------------------- |
| `GET`  | `/health-check`                  | Server liveness check                       |
| `GET`  | `/campus`                        | List all active campuses                    |
| `GET`  | `/campus/:id`                    | Get campus by ID                            |
| `GET`  | `/campus/:id/pilot`              | Get pilot assigned to a campus              |
| `WS`   | `/ws/pilot?pilot_id=&campus_id=` | WebSocket — pilot GPS streaming (port 8085) |

> **Ports:** REST API → `:8084` · WebSocket → `:8085`

---

## Prerequisites

Make sure you have the following installed:

- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) and npm
- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (includes Docker Compose)
- [GitHub CLI `gh`](https://cli.github.com/) _(optional, for contributors)_

---

## Getting Started / Instalación

### 1. Clone the repository

```bash
git clone https://github.com/k0k4developer/umg-bus-app.git
cd umg-bus-app
```

### 2. Configure environment variables

```bash
# Backend
cp backend/.env.example backend/.env
# Edit backend/.env with your actual values
```

> ⚠️ **NEVER commit `.env` files.** They are listed in `.gitignore` and contain real credentials.

### 3. Start infrastructure (PostgreSQL + Redis)

```bash
docker compose up -d
```

This starts:

- **PostgreSQL 15** on `localhost:5436`
- **Redis 8** on `localhost:6381`

### 4. Database migrations

Migrations run **automatically on backend startup** via `golang-migrate`. No manual step needed.

To run manually:

```bash
migrate -path ./backend/migrations \
  -database "postgres://admin:PASSWORD@localhost:5436/transportation-tracker-db?sslmode=disable" up
```

### 5. Start the backend

```bash
cd backend
go run ./cmd/server
```

Server starts on:

- **REST API:** `http://localhost:8084`
- **WebSocket:** `ws://localhost:8085`

### 6. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend runs at `http://localhost:5173`

---

## Branch Strategy / Estrategia de Branches

This project follows a structured branching model inspired by Git Flow:

| Branch                       | Purpose                                                             |
| ---------------------------- | ------------------------------------------------------------------- |
| `main`                       | Production-ready code. Only stable, reviewed code merges here.      |
| `develop`                    | Integration branch. All features merge here first. Base de trabajo. |
| `feature/backend-core`       | Go backend (REST API, DB, config, repositories)                     |
| `feature/frontend-ui`        | React frontend development                                          |
| `feature/websocket-realtime` | WebSocket Hub, pilot handler, real-time location                    |
| `feature/docker-infra`       | Docker Compose, infrastructure configuration                        |

**Merge workflow:**

```
feature/* ──► develop ──► main
```

All feature branches are created from `develop` and merge back via Pull Request. Only `develop` merges into `main` after review.

---

## Project Structure

```
umg-bus-app/
├── backend/
│   ├── cmd/server/         # Entry point (main.go)
│   ├── internal/
│   │   ├── config/         # Environment variable loading
│   │   ├── db/             # PostgreSQL + Redis connectors, migrations
│   │   ├── domain/         # Domain structs (Campus, Pilot)
│   │   ├── location/       # WebSocket Hub + Pilot handler
│   │   └── repository/     # Data access layer (campus_repo, pilot_repo)
│   ├── migrations/         # SQL migration files (golang-migrate)
│   ├── .env.example        # Environment template (safe to commit ✅)
│   └── go.mod / go.sum
├── frontend/
│   ├── src/
│   │   ├── App.jsx         # Root component
│   │   └── main.jsx        # React entry point
│   ├── public/
│   └── vite.config.js
├── docker-compose.yml      # PostgreSQL 15 + Redis 8
├── .gitignore              # Excludes .env, binaries, node_modules, etc.
└── README.md
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch from `develop`:
   ```bash
   git checkout develop
   git checkout -b feature/your-feature-name
   ```
3. Commit following [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: add campus search endpoint
   fix: correct Redis TTL calculation
   docs: update API endpoint table
   chore: update docker-compose ports
   ```
4. Push and open a Pull Request targeting `develop`
5. Request review and address feedback

---

## License

This project is licensed under the **MIT License**. See [LICENSE](LICENSE) for details.

---

<div align="center">

Made with ❤️ for the Google Summer of Code Community
_GSoC Proposal · Open Source University Transport Tracking_

</div>
