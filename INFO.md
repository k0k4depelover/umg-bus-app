# UMG Bus App — Guía de inicio

## Requisitos

- [Go 1.21+](https://go.dev/dl/)
- [Docker Desktop](https://www.docker.com/products/docker-desktop/)

---

## 1. Levantar la base de datos y Redis

Desde la raíz del proyecto:

```bash
docker compose up -d
```

Esto levanta:
- **PostgreSQL** en el puerto `5436`
- **Redis** en el puerto `6381`

Para verificar que están corriendo:

```bash
docker compose ps
```

---

## 2. Configurar variables de entorno

```bash
cp backend/.env.example backend/.env
```

Edita `backend/.env` con los valores reales. Para el setup local con Docker, usa esto:

```env
# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5436
POSTGRES_USER=admin
POSTGRES_PASSWORD=postgresumgtracker
POSTGRES_DB=transportation-tracker-db

# Redis
REDIS_ADDR=localhost:6381
REDIS_PASSWORD=
REDIS_DB=0

# Servidor
APP_PORT=8080
```

> Los valores de POSTGRES deben coincidir con lo definido en `docker-compose.yml`.

---

## 3. Instalar dependencias de Go

```bash
cd backend
go mod tidy
```

---

## 4. Correr el servidor

```bash
cd backend
go run main.go
```

O compilar y ejecutar:

```bash
cd backend
go build -o bin/server .
./bin/server
```

El servidor queda disponible en `http://localhost:8080`.

---

## 5. Correr pruebas

```bash
cd backend
go test ./...
```

---

## Estructura del proyecto

```
umg-bus-app/
├── backend/
│   ├── internal/
│   │   ├── config/     # Carga de variables de entorno
│   │   └── db/         # Conexiones a PostgreSQL y Redis
│   ├── main.go
│   ├── go.mod
│   └── .env.example
├── frontend/           # (pendiente)
└── docker-compose.yml
```

---

## Apagar los contenedores

```bash
docker compose down
```

Para borrar también los volúmenes (datos):

```bash
docker compose down -v
```
