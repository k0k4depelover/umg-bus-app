# Errores encontrados

## Error 1 — DSN de PostgreSQL mal formateado

**Archivo:** `backend/internal/config/config.go` línea 40

**Síntoma:**
```
postgres: ping failed: failed to connect to user=oskar database=
tls error: server refused TLS connection
FATAL: la autentificación password falló para el usuario «oskar»
```

Go intentaba conectarse como el usuario del sistema operativo (`oskar`) en el puerto por defecto `5432`, ignorando completamente las variables de entorno.

**Causa:**
El DSN estaba construido con formato incorrecto. `pgx` no reconoció los parámetros y cayó a los defaults del sistema:

```go
// MAL — pgx no entiende este formato
"host:%s, port:%s, user:%s, pass:%s, dbname:%s, sslmode=dissable"

// BIEN
"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable"
```

Errores específicos en el string original:
- `host:` en vez de `host=`
- `pass:` en vez de `password=`
- `sslmode=dissable` (typo) en vez de `sslmode=disable`
- Comas entre parámetros (no son parte del formato DSN de pgx)

**Fix aplicado:** `config.go` corregido con el formato `key=value` estándar de pgx.

---

## Error 2 — Archivo `.env` no existía

**Síntoma:** Las variables de entorno nunca se cargaban.

**Causa:** El archivo `backend/.env` no había sido creado a partir de `.env.example`.

**Fix aplicado:** Se creó `backend/.env` con los valores correctos para el entorno local con Docker.

---

## Error 3 — Redis: Ping() no retorna error directamente

**Archivo:** `backend/internal/db/redis.go` línea 23

**Síntoma:**
```
No se pudo cargar redis Redis no responde %!w(*redis.StatusCmd=&{...PONG})
```
Redis respondía `PONG` (estaba funcionando), pero el servidor fallaba igual.

**Causa:**
`client.Ping()` devuelve `*redis.StatusCmd`, no un `error`. Al pasarlo directo al `if`, el puntero siempre es no-nil aunque Redis responda correctamente.

```go
// MAL — evalúa el puntero, no el error
if err := client.Ping(context.Background()); err != nil {

// BIEN — extrae el error real con .Err()
if err := client.Ping(context.Background()).Err(); err != nil {
```

**Fix aplicado:** Agregado `.Err()` al resultado de `Ping()`.
