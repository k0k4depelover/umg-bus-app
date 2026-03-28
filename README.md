# UMG Bus App

<div align="center">

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![React Native](https://img.shields.io/badge/React_Native-0.84-61DAFB?style=for-the-badge&logo=react&logoColor=black)](https://reactnative.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=for-the-badge&logo=postgresql&logoColor=white)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-8-DC382D?style=for-the-badge&logo=redis&logoColor=white)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=for-the-badge&logo=docker&logoColor=white)](https://docs.docker.com/compose/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)

**Sistema de monitoreo de transporte universitario en tiempo real**
_GSoC Proposal - Universidad Mesoamericana de Guatemala (UMG)_

</div>

---

## Descripcion del proyecto

UMG Bus App es un sistema de rastreo de transporte en tiempo real para universidades. Permite a los estudiantes ver la ubicacion exacta de los buses del campus en un mapa, y a los pilotos (conductores) transmitir su posicion GPS desde su dispositivo movil.

El sistema se compone de:
- **Backend en Go** con API REST, GraphQL y WebSocket
- **App movil nativa** en React Native (Android/iOS) para pilotos y estudiantes
- **PostgreSQL** para datos persistentes y **Redis** para posiciones en tiempo real

---

## Tech Stack

### Backend

| Tecnologia | Version | Uso |
|---|---|---|
| [Go](https://go.dev/) | 1.25 | Lenguaje del backend |
| [Gofiber v2](https://gofiber.io/) | v2 | Framework HTTP REST |
| [gqlgen](https://gqlgen.com/) | latest | GraphQL server |
| [PostgreSQL](https://www.postgresql.org/) | 15 | Base de datos relacional |
| [pgx/v5](https://github.com/jackc/pgx) | v5 | Driver PostgreSQL para Go |
| [Redis](https://redis.io/) | 8 | Cache en tiempo real + Pub/Sub |
| [coder/websocket](https://github.com/coder/websocket) | latest | Protocolo WebSocket |
| [golang-migrate](https://github.com/golang-migrate/migrate) | latest | Migraciones de base de datos |

### Mobile

| Tecnologia | Version | Uso |
|---|---|---|
| [React Native](https://reactnative.dev/) | 0.84 | Framework movil nativo |
| [TypeScript](https://www.typescriptlang.org/) | 5.8 | Tipado estatico |
| [React Navigation](https://reactnavigation.org/) | 7.x | Navegacion (stack + tabs) |
| [MapLibre React Native](https://github.com/maplibre/maplibre-react-native) | latest | Mapas open source |
| [Zustand](https://zustand-demo.pmnd.rs/) | 5.x | Estado global (auth) |
| [AsyncStorage](https://react-native-async-storage.github.io/async-storage/) | latest | Persistencia de tokens |
| [Geolocation Service](https://github.com/Agontuk/react-native-geolocation-service) | latest | GPS tracking |

### Infraestructura

| Tecnologia | Uso |
|---|---|
| [Docker Compose](https://docs.docker.com/compose/) | Orquesta PostgreSQL, Redis y Backend |

---

## Arquitectura

```
┌──────────────────────────────────────────────────────────────┐
│                        Clientes                               │
│                                                               │
│   App Piloto (mobile)            App Estudiante (mobile)      │
│   WebSocket -> /ws/pilot         WebSocket -> /ws/student     │
│   Envia GPS                      Recibe posiciones            │
└──────────────┬──────────────────────────┬────────────────────┘
               │                          │
               v                          v
┌──────────────────────────────────────────────────────────────┐
│                       Go Backend                              │
│                                                               │
│  ┌─────────────────────┐    ┌──────────────────────────────┐ │
│  │  Gofiber (REST)     │    │  WebSocket Server            │ │
│  │  Puerto :8084       │    │  Puerto :8085                │ │
│  │                     │    │                              │ │
│  │  /health-check      │    │  /ws/pilot  -> GPS in        │ │
│  │  /auth/login        │    │  /ws/student -> GPS out      │ │
│  │  /auth/refresh      │    │                              │ │
│  │  /graphql           │    │  Hub.PublishLocation()       │ │
│  └─────────────────────┘    │       -> Redis Pub/Sub       │ │
│                              └──────────────────────────────┘ │
└────────────────┬─────────────────────────────────────────────┘
                 │
       ┌─────────┴─────────┐    ┌───────────────────┐
       │   PostgreSQL 15   │    │     Redis 8        │
       │   Puerto 5436     │    │     Puerto 6381    │
       │                   │    │                    │
       │  campus           │    │  GPS state (TTL)   │
       │  pilots           │    │  Pub/Sub channels  │
       │  students         │    │  Sessions (JWT)    │
       │  sessions         │    │                    │
       │  campus_changes   │    └───────────────────┘
       │  location_log     │
       └───────────────────┘
```

**Flujo de datos:**

1. El piloto abre la app y se autentica (JWT)
2. Se conecta por WebSocket a `/ws/pilot` y transmite GPS continuamente
3. El Hub guarda la posicion en Redis (con TTL) y publica en Pub/Sub
4. Los estudiantes conectados a `/ws/student` reciben actualizaciones en tiempo real
5. El mapa en la app del estudiante muestra la posicion del bus

---

## Estructura del proyecto

```
umg-bus-app/
├── backend/
│   ├── cmd/server/              # Entry point (main.go)
│   ├── internal/
│   │   ├── auth/                # Autenticacion (JWT, login, middleware)
│   │   ├── config/              # Carga de variables de entorno
│   │   ├── db/                  # Conexiones PostgreSQL + Redis, migraciones
│   │   ├── domain/              # Structs de dominio (Campus, Pilot, Student)
│   │   ├── location/            # WebSocket Hub + handlers (pilot, student)
│   │   ├── mcp/                 # Servidor MCP
│   │   ├── middleware/          # Rate limiter
│   │   ├── repository/          # Capa de acceso a datos
│   │   └── testutil/            # Utilidades para tests
│   ├── graph/                   # GraphQL (schema, resolvers, modelos)
│   ├── migrations/              # SQL migrations (golang-migrate)
│   ├── scripts/                 # Scripts auxiliares (smoke test)
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod
│   └── gqlgen.yml
├── mobile/
│   ├── android/                 # Proyecto nativo Android (Gradle)
│   ├── ios/                     # Proyecto nativo iOS (CocoaPods)
│   ├── src/
│   │   ├── api/                 # Comunicacion con backend
│   │   │   ├── config.ts        # URLs y endpoints
│   │   │   ├── http.ts          # HTTP client con auto-refresh de tokens
│   │   │   ├── auth.ts          # Login, logout, refresh
│   │   │   └── campus.ts        # Queries GraphQL
│   │   ├── components/          # Componentes de UI reutilizables
│   │   │   ├── Button.tsx
│   │   │   ├── Input.tsx
│   │   │   ├── Card.tsx
│   │   │   ├── StatusBadge.tsx
│   │   │   ├── LoadingScreen.tsx
│   │   │   ├── EmptyState.tsx
│   │   │   └── MapHeader.tsx
│   │   ├── hooks/               # Logica reutilizable
│   │   │   ├── useAuth.ts       # Estado de autenticacion (Zustand)
│   │   │   ├── useWebSocket.ts  # Conexion WebSocket
│   │   │   └── useLocation.ts   # GPS tracking (piloto)
│   │   ├── navigation/          # Estructura de navegacion
│   │   │   ├── RootNavigator.tsx
│   │   │   ├── AuthNavigator.tsx
│   │   │   ├── PilotNavigator.tsx
│   │   │   └── StudentNavigator.tsx
│   │   ├── screens/             # Pantallas
│   │   │   ├── auth/LoginScreen.tsx
│   │   │   ├── pilot/PilotMapScreen.tsx
│   │   │   ├── student/StudentMapScreen.tsx
│   │   │   └── shared/          # ProfileScreen, SettingsScreen, CampusChangeScreen
│   │   ├── theme/               # Sistema de diseno (colors, spacing, typography)
│   │   └── types/               # Tipos TypeScript
│   ├── __tests__/
│   ├── App.tsx                  # Punto de entrada
│   ├── package.json
│   ├── metro.config.js
│   ├── babel.config.js
│   ├── tsconfig.json
│   └── jest.config.js
├── docker-compose.yml           # PostgreSQL + Redis + Backend
├── SETUP_GUIDE.md               # Guia completa de instalacion
├── CLAUDE.md
└── README.md
```

---

## API Endpoints

### REST (puerto 8084)

| Metodo | Ruta | Descripcion |
|---|---|---|
| `GET` | `/health-check` | Verificacion de estado del servidor |
| `POST` | `/auth/login` | Login (devuelve access + refresh token) |
| `POST` | `/auth/refresh` | Refrescar access token |
| `POST` | `/auth/logout` | Cerrar sesion |
| `POST` | `/graphql` | Endpoint GraphQL |

### WebSocket (puerto 8085)

| Ruta | Descripcion |
|---|---|
| `/ws/pilot?token=JWT` | Piloto transmite posicion GPS |
| `/ws/student?token=JWT` | Estudiante recibe posiciones en tiempo real |

### GraphQL

Queries disponibles en `/graphql`:
- `campuses` - Lista todos los campus activos
- `campus(id)` - Detalle de un campus
- `liveLocation(campusId)` - Posicion actual del bus en un campus

---

## Database Schema

| Tabla | Descripcion |
|---|---|
| `campus` | Campus universitarios con bounding box geografico y ruta GeoJSON |
| `pilots` | Conductores autorizados por campus |
| `students` | Estudiantes vinculados a un campus |
| `sessions` | Registro de sesiones (audit log) |
| `student_campus_changes` | Log inmutable de cambios de campus |
| `location_log` | Historial de posiciones GPS |

---

## Quick Start

### 1. Clonar el repositorio

```bash
git clone https://github.com/k0k4depelover/umg-bus-app.git
cd umg-bus-app
```

### 2. Levantar todo con Docker

```bash
docker compose up -d
```

Esto inicia PostgreSQL (5436), Redis (6381) y el Backend (8084/8085).

### 3. Verificar

```bash
docker compose ps
curl http://localhost:8084/health-check
```

### 4. App mobile

```bash
cd mobile
npm install
npx react-native start         # Terminal 1: Metro bundler
npx react-native run-android   # Terminal 2: Compilar e instalar
```

> Para la guia completa de setup mobile (Android Studio, emulador, variables de entorno, troubleshooting), ver **[SETUP_GUIDE.md](SETUP_GUIDE.md)**.

---

## Branch Strategy

| Branch | Proposito |
|---|---|
| `main` | Codigo estable, listo para produccion |
| `feature/frontend-mobile` | Desarrollo de la app movil React Native |
| `feature/middleware-mcp-impl` | Middleware de autenticacion y MCP |

```
feature/* --> main (via Pull Request)
```

---

## Comandos utiles

### Backend

```bash
cd backend
go run ./cmd/server       # Ejecutar servidor
go test ./...             # Correr tests
go build -o bin/server .  # Compilar
```

### Mobile

```bash
cd mobile
npm start                       # Metro bundler
npx react-native run-android   # Compilar Android
npm test                        # Tests con Jest
adb logcat -s ReactNativeJS    # Ver logs de la app
```

### Docker

```bash
docker compose up -d      # Levantar servicios
docker compose ps         # Ver estado
docker compose logs -f    # Ver logs en tiempo real
docker compose down       # Apagar
docker compose down -v    # Apagar y borrar datos
```

---

## Anexos

- **[SETUP_GUIDE.md](SETUP_GUIDE.md)** - Guia paso a paso para configurar el proyecto desde cero en un dispositivo nuevo. Incluye requisitos, configuracion de Android Studio, creacion de emulador, troubleshooting de errores comunes de Gradle/Java/Metro, y como debuggear la app mobile.
- **[DEVELOPER_GUIDE.md](mobile/DEVELOPER_GUIDE.md)** - Guia para desarrolladores de la app mobile: arquitectura, como agregar pantallas/componentes, debugging, errores comunes, guia de estilos y testing.

---

## Contribuir

1. Fork del repositorio
2. Crear branch desde `main`:
   ```bash
   git checkout -b feature/nombre-del-feature
   ```
3. Commits siguiendo [Conventional Commits](https://www.conventionalcommits.org/):
   ```
   feat: agregar pantalla de registro
   fix: corregir refresh de token
   docs: actualizar guia de setup
   ```
4. Push y abrir Pull Request hacia `main`

---

## Licencia

Este proyecto esta bajo la licencia **MIT**. Ver [LICENSE](LICENSE) para detalles.

---

<div align="center">

Made for the Google Summer of Code Community
_GSoC Proposal - Open Source University Transport Tracking_

</div>
