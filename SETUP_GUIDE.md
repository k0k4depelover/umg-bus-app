# UMG Bus App — Guia completa de setup desde cero

Guia paso a paso para clonar, configurar y ejecutar todo el proyecto (backend + mobile Android) en un dispositivo nuevo.

---

## Requisitos previos

Instalar todo lo siguiente **antes** de empezar:

| Herramienta | Version minima | Descarga |
|---|---|---|
| Git | 2.30+ | https://git-scm.com/downloads |
| Docker Desktop | 4.x | https://www.docker.com/products/docker-desktop/ |
| Node.js | 22+ (LTS) | https://nodejs.org/ |
| Java JDK | 17 (recomendado 21) | https://docs.aws.amazon.com/corretto/latest/corretto-21-ug/downloads-list.html |
| Android Studio | Ladybug+ | https://developer.android.com/studio |
| Go | 1.21+ | https://go.dev/dl/ |

> **Importante sobre Java**: Si tienes JDK 24 u otra version muy nueva, Gradle puede fallar. Usa JDK 17 o 21.

---

## Paso 1 — Clonar el repositorio

```bash
git clone https://github.com/k0k4depelover/umg-bus-app.git
cd umg-bus-app
```

Cambiar a la rama de desarrollo mobile:

```bash
git checkout feature/frontend-mobile
```

---

## Paso 2 — Configurar variables de entorno del sistema (Windows)

### ANDROID_HOME

1. Abrir **"Variables de entorno del sistema"** (buscar "variables de entorno" en el menu inicio)
2. En **Variables de usuario**, crear:
   - `ANDROID_HOME` = `C:\Users\TU_USUARIO\AppData\Local\Android\Sdk`
3. En **Path**, agregar:
   - `%ANDROID_HOME%\platform-tools`
   - `%ANDROID_HOME%\emulator`

### JAVA_HOME (si es necesario)

Si tienes multiples versiones de Java, asegurate de que `JAVA_HOME` apunte a JDK 17 o 21, **no** a JDK 24+.

Si `JAVA_HOME` apunta a una version incompatible, no te preocupes — el proyecto esta configurado para usar el JDK que viene con Android Studio (ver "Troubleshooting" al final).

### Verificar

Abre una **nueva terminal** (para que tome las variables) y ejecuta:

```bash
node --version    # debe ser 22+
java -version     # debe ser 17 o 21
adb --version     # debe responder (viene con Android SDK)
go version        # debe ser 1.21+
docker --version  # debe responder
```

---

## Paso 3 — Levantar el backend (Docker)

Desde la raiz del proyecto:

```bash
docker compose up -d
```

Esto levanta 3 contenedores:

| Servicio | Puerto | Descripcion |
|---|---|---|
| PostgreSQL | 5436 | Base de datos principal |
| Redis | 6381 | Posiciones de buses en tiempo real |
| Backend (Go) | 8084 (HTTP), 8085 (WebSocket) | API del servidor |

Verificar que estan corriendo:

```bash
docker compose ps
```

Deberias ver los 3 contenedores en estado `Up` o `running`.

Verificar que el backend responde:

```bash
curl http://localhost:8084/health-check
```

> **Nota**: Si prefieres correr el backend manualmente sin Docker (solo las DBs en Docker), mira la seccion "Backend manual" al final.

---

## Paso 4 — Configurar Android Studio

### 4.1 Instalar SDK y herramientas

1. Abrir Android Studio
2. Ir a **Settings** > **Languages & Frameworks** > **Android SDK**
3. En la pestana **SDK Platforms**, instalar:
   - Android 15 (API 35) o Android 14 (API 34)
4. En la pestana **SDK Tools**, verificar que estan instalados:
   - Android SDK Build-Tools 36
   - Android SDK Platform-Tools
   - Android Emulator
   - NDK 27.1.12297006 (se instala automaticamente al compilar)

### 4.2 Crear un emulador (AVD)

1. En Android Studio, abrir **Device Manager** (icono de telefono en la barra derecha)
2. Click en **Create Virtual Device**
3. Seleccionar **Pixel 7**
4. Seleccionar imagen del sistema: **API 34 o 35**, arquitectura **x86_64**
5. Darle al menos **4 GB de RAM**
6. Click en **Finish**
7. Iniciar el emulador (boton de play)

> **Tip**: Elige siempre **x86_64** para emuladores en PC — corre mucho mas rapido que ARM.

---

## Paso 5 — Instalar dependencias del proyecto mobile

```bash
cd mobile
npm install
```

---

## Paso 6 — Crear local.properties (Android)

Crear el archivo `mobile/android/local.properties` con la ruta de tu SDK:

```properties
sdk.dir=C:\\Users\\TU_USUARIO\\AppData\\Local\\Android\\Sdk
```

Reemplaza `TU_USUARIO` con tu nombre de usuario de Windows.

---

## Paso 7 — Ejecutar la app en el emulador

Asegurate de que:
- El backend esta corriendo (paso 3)
- El emulador esta abierto y funcionando (paso 4.2)

Luego, en **una terminal**:

```bash
cd mobile
npx react-native start
```

Espera a ver el mensaje `Dev server ready`.

En **otra terminal**:

```bash
cd mobile
npx react-native run-android
```

La primera compilacion tarda **15-20 minutos** (Gradle descarga dependencias y compila todo el proyecto nativo). Las siguientes compilaciones son mucho mas rapidas.

Una vez que termine, la app se instala automaticamente en el emulador.

---

## Paso 8 — Verificar que todo funciona

1. La app debe abrir en el emulador mostrando la pantalla de **Login**
2. El backend debe estar accesible desde el emulador en `http://10.0.2.2:8084` (ya configurado en el codigo)
3. Puedes ver logs de la app con:

```bash
adb logcat -s ReactNativeJS
```

---

## Estructura del proyecto

```
umg-bus-app/
├── backend/                 # API en Go (Fiber + GraphQL)
│   ├── cmd/server/          # Entry point del servidor
│   ├── internal/            # Logica interna (config, db, auth, handlers)
│   ├── graph/               # Schema y resolvers de GraphQL
│   ├── Dockerfile           # Imagen Docker del backend
│   ├── go.mod
│   └── go.sum
├── mobile/                  # App React Native
│   ├── android/             # Proyecto nativo Android
│   ├── src/
│   │   ├── api/             # Comunicacion con backend (HTTP + WebSocket)
│   │   ├── components/      # Componentes de UI reutilizables
│   │   ├── hooks/           # Logica reutilizable (auth, GPS, WebSocket)
│   │   ├── navigation/      # Navegacion (Auth, Pilot, Student)
│   │   ├── screens/         # Pantallas de la app
│   │   ├── theme/           # Colores, tipografia, espaciado
│   │   └── types/           # Tipos TypeScript
│   ├── App.tsx              # Punto de entrada
│   ├── package.json
│   └── metro.config.js
├── docker-compose.yml       # Orquestacion de servicios
└── CLAUDE.md
```

---

## Puertos

| Servicio | Puerto | Notas |
|---|---|---|
| PostgreSQL | 5436 | Mapeado desde 5432 del contenedor |
| Redis | 6381 | Mapeado desde 6379 del contenedor |
| Backend HTTP | 8084 | API REST + GraphQL |
| Backend WebSocket | 8085 | Posiciones en tiempo real |
| Metro Bundler | 8081 | Dev server de React Native |

---

## Backend manual (opcional)

Si prefieres correr solo las bases de datos en Docker y el backend manualmente:

```bash
# Solo DBs
docker compose up -d postgres redis

# Crear archivo de entorno
cp backend/.env.example backend/.env
# Editar backend/.env con:
#   POSTGRES_HOST=localhost
#   POSTGRES_PORT=5436
#   POSTGRES_USER=admin
#   POSTGRES_PASSWORD=postgresumgtracker
#   POSTGRES_DB=transportation-tracker-db
#   REDIS_ADDR=localhost:6381
#   APP_PORT=8084

# Instalar dependencias e iniciar
cd backend
go mod tidy
go run main.go
```

---

## Troubleshooting

### Error: "Unsupported class file major version 68"

Tu `JAVA_HOME` apunta a JDK 24+. Solucion: el archivo `mobile/android/gradle.properties` ya incluye la linea:

```
org.gradle.java.home=C:\\Program Files\\Android\\Android Studio\\jbr
```

Esto fuerza a Gradle a usar el JDK 21 que viene con Android Studio. Si tu Android Studio esta en otra ruta, ajusta este valor.

### Error: "Could not find org.asyncstorage.shared_storage:storage-android:1.0.0"

El archivo `mobile/android/build.gradle` ya incluye el repositorio local de AsyncStorage. Si reaparece, verifica que `allprojects.repositories` contiene:

```gradle
maven { url "${rootProject.projectDir}/../node_modules/@react-native-async-storage/async-storage/android/local_repo" }
```

### Error: "react-native-worklets library not found"

Instalar la dependencia faltante:

```bash
cd mobile
npm install react-native-worklets
```

### Error: "Unable to load script" (pantalla roja en emulador)

Metro no esta corriendo o esta en otro puerto. Solucion:

```bash
# Cerrar todos los procesos de Metro
taskkill /IM node.exe /F

# Reiniciar
cd mobile
npx react-native start
# En otra terminal:
npx react-native run-android
```

### Error: "SDK location not found"

Crear `mobile/android/local.properties` con:

```
sdk.dir=C:\\Users\\TU_USUARIO\\AppData\\Local\\Android\\Sdk
```

### Error: "Another process is running on port 8081"

Cerrar el proceso viejo:

```bash
# Ver que proceso usa el puerto
netstat -ano | findstr 8081

# Matar el proceso (reemplazar PID con el numero real)
taskkill /PID <PID> /F
```

### El teclado hace saltar la pantalla de login

Verificar que en `LoginScreen.tsx` el `KeyboardAvoidingView` tenga:

```tsx
behavior={Platform.OS === 'ios' ? 'padding' : undefined}
```

**No** usar `'height'` en Android.

### Conexion desde dispositivo fisico

Si usas un telefono fisico en vez del emulador, cambia la IP en `mobile/src/api/config.ts`:

```typescript
// Cambiar de 10.0.2.2 a la IP de tu maquina
const DEV_HOST_ANDROID = 'http://192.168.X.X';
```

Tu PC y el telefono deben estar en la misma red WiFi.

---

## Tests

```bash
cd mobile
npm test
```

Los tests usan mocks para los modulos nativos (gesture handler, safe area, etc.) configurados en `jest.setup.js`.

---

## Apagar todo

```bash
# Detener contenedores
docker compose down

# Detener contenedores Y borrar datos de las DBs
docker compose down -v
```
