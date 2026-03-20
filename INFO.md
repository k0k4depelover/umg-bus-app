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

# 1. Crear archivo de herramientas (desde Git Bash)

printf '//go:build tools\npackage tools\nimport \_ "github.com/99designs/gqlgen"' | gofmt > tools.go

# 2. Descargar dependencia

go mod tidy

# 3. Generar estructura del proyecto

go run github.com/99designs/gqlgen init

# 4. Limpiar dependencias

go mod tidy

Eso genera:

graph/
├── generated.go # Codigo generado (NO tocar)
├── model/models_gen.go # Modelos del schema
├── resolver.go # Struct base
├── schema.graphqls # Schema GraphQL (defines queries/mutations)
└── schema.resolvers.go # Aca escribes la logica
gqlgen.yml # Configuracion
server.go # Server de ejemplo

Tras escribir ya el schema.grapgqls de graphql, debemos regenerar el esquema con:

```bash
go run github.com/99designs/gqlgen generate
```

**El flujo completo de GraphQL en el Día 3**

Hay cuatro capas y es importante saber qué hace cada una:

```
schema.graphqls          →   contrato público (qué puede pedir el cliente)
       ↓  gqlgen generate
models_gen.go            →   structs Go generados automáticamente
       ↓  tú implementas
schema.resolvers.go      →   lógica real (va a Redis o PostgreSQL)
       ↓  gqlgen conecta
HTTP POST /graphql        →   handler que parsea, valida y ejecuta
```
