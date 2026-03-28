# UMG Bus — Mobile App Developer Guide

Guía completa para desarrolladores que trabajan con la app móvil de UMG Bus.

---

## Tabla de Contenidos

1. [Requisitos previos](#requisitos-previos)
2. [Setup inicial](#setup-inicial)
3. [Estructura del proyecto](#estructura-del-proyecto)
4. [Arquitectura](#arquitectura)
5. [Cómo funciona cada módulo](#cómo-funciona-cada-módulo)
6. [Cómo hacer cambios](#cómo-hacer-cambios)
7. [Debugging](#debugging)
8. [Errores comunes](#errores-comunes)
9. [Guía de estilos](#guía-de-estilos)
10. [Testing](#testing)

---

## Requisitos Previos

### Software necesario

| Herramienta | Versión mínima | Para qué |
|-------------|---------------|----------|
| Node.js | 18+ | Runtime de JavaScript |
| npm | 9+ | Gestor de paquetes |
| Java JDK | 17+ | Compilar Android |
| Android Studio | Hedgehog+ | SDK de Android, emulador |
| Xcode | 15+ | Solo macOS — compilar iOS |
| Git | 2.30+ | Control de versiones |

### Variables de entorno (Android)

Agrega estas a tu perfil de shell (`~/.bashrc`, `~/.zshrc`, o System Environment Variables en Windows):

```bash
export ANDROID_HOME=$HOME/Android/Sdk
# Windows: set ANDROID_HOME=C:\Users\TU_USUARIO\AppData\Local\Android\Sdk

export PATH=$PATH:$ANDROID_HOME/emulator
export PATH=$PATH:$ANDROID_HOME/platform-tools
```

### Verificar setup

```bash
# Node
node --version   # debe ser 18+

# Java
java -version    # debe ser 17+

# Android SDK
adb --version    # debe responder

# Verificar React Native
npx react-native doctor
```

---

## Setup Inicial

### 1. Clonar e instalar

```bash
cd mobile
npm install
```

### 2. Android: instalar dependencias nativas

```bash
cd android
./gradlew clean
cd ..
```

### 3. iOS (solo macOS)

```bash
cd ios
bundle install        # Instalar CocoaPods via Bundler
bundle exec pod install
cd ..
```

### 4. Ejecutar

```bash
# Android (con emulador corriendo o dispositivo conectado)
npx react-native run-android

# iOS (solo macOS)
npx react-native run-ios
```

### 5. Backend

Asegúrate de que el backend esté corriendo (puertos 8084 y 8085):

```bash
cd ../backend
docker compose up -d   # PostgreSQL + Redis
go run ./cmd/server
```

### Conexión desde emulador

- **Android emulator** → El backend en `localhost` se accede via `10.0.2.2`
  - Ya está configurado en `src/api/config.ts`
- **Dispositivo físico** → Cambia la IP en `src/api/config.ts` a la IP de tu máquina
- **iOS simulator** → Usa `localhost` directamente

---

## Estructura del Proyecto

```
mobile/
├── App.tsx                     ← Punto de entrada principal
├── android/                    ← Proyecto nativo Android (no tocar salvo necesario)
├── ios/                        ← Proyecto nativo iOS (no tocar salvo necesario)
├── src/
│   ├── api/                    ← Comunicación con el backend
│   │   ├── config.ts           ← URLs y endpoints
│   │   ├── http.ts             ← HTTP client con auto-refresh de tokens
│   │   ├── auth.ts             ← Login, logout, refresh
│   │   ├── campus.ts           ← Queries GraphQL (campus, live location)
│   │   └── index.ts
│   ├── components/             ← Componentes reutilizables de UI
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   ├── Card.tsx
│   │   ├── StatusBadge.tsx
│   │   ├── LoadingScreen.tsx
│   │   ├── EmptyState.tsx
│   │   ├── MapHeader.tsx
│   │   └── index.ts
│   ├── hooks/                  ← Lógica reutilizable
│   │   ├── useAuth.ts          ← Estado de autenticación (Zustand)
│   │   ├── useWebSocket.ts     ← Conexión WebSocket (piloto/estudiante)
│   │   ├── useLocation.ts      ← GPS tracking (piloto)
│   │   └── index.ts
│   ├── navigation/             ← Estructura de navegación
│   │   ├── RootNavigator.tsx   ← Switch: Auth vs Pilot vs Student
│   │   ├── AuthNavigator.tsx   ← Stack: Login (futuro: Register)
│   │   ├── PilotNavigator.tsx  ← Tabs: Mapa, Perfil, Ajustes
│   │   ├── StudentNavigator.tsx← Tabs: Mapa, Campus, Perfil, Ajustes
│   │   └── index.ts
│   ├── screens/                ← Pantallas de la app
│   │   ├── auth/
│   │   │   └── LoginScreen.tsx
│   │   ├── pilot/
│   │   │   └── PilotMapScreen.tsx
│   │   ├── student/
│   │   │   └── StudentMapScreen.tsx
│   │   └── shared/             ← Pantallas compartidas (mocks por ahora)
│   │       ├── ProfileScreen.tsx
│   │       ├── SettingsScreen.tsx
│   │       └── CampusChangeScreen.tsx
│   ├── theme/                  ← Sistema de diseño
│   │   ├── colors.ts           ← Paleta de colores (rojo Anthropic)
│   │   ├── spacing.ts          ← Espaciado consistente
│   │   ├── typography.ts       ← Estilos de texto
│   │   └── index.ts
│   └── types/                  ← Tipos TypeScript
│       └── index.ts            ← Todos los tipos (API, navigation, etc.)
```

---

## Arquitectura

### Flujo de datos

```
┌────────────────────────────────────────────────────────┐
│                        App.tsx                          │
│  GestureHandler → SafeArea → Auth Init → Navigator     │
└──────────────────────┬─────────────────────────────────┘
                       │
          ┌────────────┴────────────┐
          │                         │
    ┌─────▼──────┐           ┌─────▼──────┐
    │  AuthStack  │           │  Main Tabs  │
    │  (Login)    │           │  (Pilot/    │
    │             │           │   Student)  │
    └─────────────┘           └──────┬──────┘
                                     │
                        ┌────────────┴────────────┐
                        │                         │
                  ┌─────▼──────┐           ┌─────▼──────┐
                  │ MapScreen  │           │   Mocks    │
                  │ (WebSocket │           │ (Profile,  │
                  │  + MapLib) │           │  Settings) │
                  └──────┬─────┘           └────────────┘
                         │
              ┌──────────┴──────────┐
              │                     │
        ┌─────▼──────┐       ┌─────▼──────┐
        │  useWebSocket│      │ useLocation │
        │  (real-time) │      │ (GPS pilot) │
        └──────┬───────┘      └──────┬──────┘
               │                     │
         ┌─────▼─────────────────────▼─────┐
         │          Backend API             │
         │  HTTP :8084  │  WebSocket :8085  │
         └─────────────────────────────────┘
```

### Estado

- **useAuth** (Zustand store): Estado global de autenticación
  - `isAuthenticated`, `claims`, `login()`, `logout()`
  - Los tokens se persisten en AsyncStorage
- **Estado local** (useState): Datos de pantalla (campus, location, etc.)
- **No usamos Redux** — Zustand es más simple y suficiente

### Navegación

La app usa **React Navigation** con este flujo:

1. Al abrir → `useAuth.init()` restaura sesión desde AsyncStorage
2. Si no hay sesión → muestra `AuthStack` (LoginScreen)
3. Si hay sesión con `role: "pilot"` → muestra `PilotNavigator` (tabs)
4. Si hay sesión con `role: "student"` → muestra `StudentNavigator` (tabs)

---

## Cómo Funciona Cada Módulo

### Autenticación (`src/api/auth.ts` + `src/hooks/useAuth.ts`)

1. Usuario ingresa credentials en LoginScreen
2. Se envía POST `/auth/login` con `{username, password, role}`
3. Backend responde con `{access_token, refresh_token}`
4. Tokens se guardan en AsyncStorage
5. Se decodifica el JWT para extraer `claims` (user_id, campus_id, role)
6. El navigator reacciona y muestra la pantalla correcta

**Auto-refresh**: El `httpClient` detecta respuestas 401 y automáticamente:
1. Envía el refresh_token a `/auth/refresh`
2. Obtiene nuevo access_token
3. Reintenta la petición original

### WebSocket (`src/hooks/useWebSocket.ts`)

**Piloto**:
1. Se conecta a `ws://host:8085/ws/pilot?token=JWT`
2. `useLocation` obtiene GPS → genera `PilotPing`
3. `sendPing()` envía `{lat, lng, bearing, speed}` al WebSocket
4. Backend almacena en Redis y difunde a estudiantes

**Estudiante**:
1. Se conecta a `ws://host:8085/ws/student?token=JWT`
2. Recibe `LiveLocation` con posición del bus
3. Actualiza el marcador en el mapa
4. Acumula coordenadas para dibujar la ruta (trail)

### Mapa (`@maplibre/maplibre-react-native`)

- **Campus bounds**: Se dibujan como polígono semi-transparente rojo
- **Bus marker**: PointAnnotation con icono de bus
- **Route trail** (estudiante): LineString acumulado de las posiciones recibidas
- **Camera**: Centrada en el centro del campus bounds

### GPS (`src/hooks/useLocation.ts`)

Solo para piloto. Usa `react-native-geolocation-service`:
- Pide permiso `ACCESS_FINE_LOCATION`
- `watchPosition` con `enableHighAccuracy: true`
- `distanceFilter: 5m` — solo envía si se movió 5+ metros
- Convierte `speed` de m/s a km/h
- Convierte `heading` a `bearing`

---

## Cómo Hacer Cambios

### Agregar una nueva pantalla

1. Crea el archivo en `src/screens/[categoria]/MiScreen.tsx`
2. Usa el template:

```tsx
import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { colors } from '../../theme/colors';
import { spacing } from '../../theme/spacing';
import { typography } from '../../theme/typography';

export function MiScreen() {
  return (
    <SafeAreaView style={styles.container}>
      <Text style={styles.title}>Mi Pantalla</Text>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.background },
  title: { ...typography.h2, color: colors.textPrimary, padding: spacing.lg },
});
```

3. Registra la pantalla en el navigator correspondiente
4. Agrega el tipo a `types/index.ts` en el ParamList correcto

### Agregar un nuevo componente

1. Crea en `src/components/MiComponente.tsx`
2. Exporta desde `src/components/index.ts`
3. Usa colores de `theme/colors`, no valores hardcodeados

### Agregar una nueva query GraphQL

1. Escribe la query en `src/api/campus.ts` (o crea un nuevo archivo de servicio)
2. Usa `httpClient.graphql<TipoRespuesta>(query, variables)`
3. Define los tipos en `src/types/index.ts`

### Agregar un endpoint REST

1. Usa `httpClient.fetch<TipoRespuesta>('/mi/endpoint', { method: 'POST', body: ... })`
2. El token se agrega automáticamente

### Implementar el registro de usuarios (futuro)

1. Agregar endpoint en backend: `POST /auth/register`
2. Crear `src/screens/auth/RegisterScreen.tsx`
3. Agregar a `AuthStackParamList` en `types/index.ts`
4. Descomentar la ruta en `AuthNavigator.tsx`
5. Agregar servicio en `src/api/auth.ts`

---

## Debugging

### Metro Bundler

```bash
# Si Metro no arranca automáticamente
npx react-native start

# Limpiar cache de Metro
npx react-native start --reset-cache
```

### React Native Debugger

1. En el emulador Android: `Ctrl + M` → "Debug"
2. En iOS simulator: `Cmd + D` → "Debug"
3. Usa Chrome DevTools o Flipper

### Logs

```bash
# Android
adb logcat | grep -i "ReactNative\|UMGBus"

# iOS (desde Xcode console)
# O usa: npx react-native log-ios
```

### Problemas de red desde emulador

Si no puedes conectar al backend:

```bash
# Verificar que el backend está corriendo
curl http://localhost:8084/health-check

# Android emulator usa 10.0.2.2 en vez de localhost
# Verificar en src/api/config.ts que la URL es correcta

# Si usas dispositivo físico, necesitas la IP de tu máquina
# Cambia en src/api/config.ts
```

### Hot Reload no funciona

```bash
# Reiniciar Metro
npx react-native start --reset-cache

# Rebuild completo
cd android && ./gradlew clean && cd ..
npx react-native run-android
```

### Errores de native modules

```bash
# Android
cd android && ./gradlew clean && cd ..
npx react-native run-android

# iOS
cd ios && pod install --repo-update && cd ..
npx react-native run-ios
```

---

## Errores Comunes

### "Unable to load script from assets"
```bash
# Android: crear directorio de assets
mkdir -p android/app/src/main/assets
npx react-native bundle --platform android --dev false --entry-file index.js --bundle-output android/app/src/main/assets/index.android.bundle
npx react-native run-android
```

### "SDK location not found"
Crea `android/local.properties`:
```
sdk.dir=C:\\Users\\TU_USUARIO\\AppData\\Local\\Android\\Sdk
```

### "Could not determine the dependencies of task ':app:compileDebugJavaWithJavac'"
```bash
cd android && ./gradlew --stop && ./gradlew clean && cd ..
```

### WebSocket no conecta
1. Verifica que el backend WS está en puerto 8085
2. Verifica que el token JWT no ha expirado
3. Verifica la URL en `src/api/config.ts`
4. En Android emulator: usa `10.0.2.2` no `localhost`

### Mapa no carga
1. MapLibre usa tiles demo por defecto — funcionan sin API key
2. Para producción, configura un tile server propio o usa MapTiler
3. Cambia `styleURL` en los componentes de mapa

### "Invariant Violation: requireNativeComponent"
Un módulo nativo no está linkeado. En RN 0.84+ el autolinking debería funcionar:
```bash
cd android && ./gradlew clean && cd ..
npx react-native run-android
```

---

## Guía de Estilos

### Colores
**NUNCA** uses colores hardcodeados. Siempre importa de `theme/colors`:

```tsx
// ✅ Correcto
import { colors } from '../theme/colors';
backgroundColor: colors.primary

// ❌ Incorrecto
backgroundColor: '#DC2626'
```

### Espaciado
Usa los valores de `theme/spacing` para consistencia:

```tsx
import { spacing } from '../theme/spacing';
padding: spacing.md    // 16
marginBottom: spacing.lg // 24
```

### Tipografía
Usa los presets de `theme/typography`:

```tsx
import { typography } from '../theme/typography';
...typography.h2     // Para títulos
...typography.body   // Para texto normal
...typography.caption // Para texto secundario
```

### Nomenclatura
- Pantallas: `PascalCase` + `Screen` sufijo → `LoginScreen.tsx`
- Componentes: `PascalCase` → `Button.tsx`
- Hooks: `camelCase` con prefijo `use` → `useAuth.ts`
- Servicios API: `camelCase` + `Service` → `authService`
- Tipos: `PascalCase` interfaces → `LoginRequest`

---

## Testing

### Ejecutar tests
```bash
npm test
```

### Agregar tests
Crea archivos `__tests__/MiComponente.test.tsx` junto al componente, o en un directorio `__tests__/`.

### Testing manual checklist

1. **Login**:
   - [ ] Login con credenciales válidas de estudiante
   - [ ] Login con credenciales válidas de piloto
   - [ ] Error con credenciales inválidas
   - [ ] Persistencia de sesión (cerrar y abrir app)

2. **Piloto**:
   - [ ] Mapa muestra bounds del campus
   - [ ] Botón inicia/detiene transmisión GPS
   - [ ] WebSocket se conecta correctamente
   - [ ] Stats se actualizan (velocidad, dirección, pings)

3. **Estudiante**:
   - [ ] Mapa muestra bounds del campus
   - [ ] Bus aparece cuando hay piloto activo
   - [ ] Ruta trail se dibuja con las posiciones
   - [ ] Info panel muestra datos correctos
   - [ ] Mensaje "sin buses" cuando no hay piloto

4. **General**:
   - [ ] Logout funciona desde perfil
   - [ ] App maneja background/foreground correctamente
   - [ ] Token refresh funciona (esperar 15 min)

---

## Stack Tecnológico

| Librería | Versión | Uso |
|----------|---------|-----|
| React Native | 0.84 | Framework móvil |
| React Navigation | 7.x | Navegación (stack + tabs) |
| MapLibre React Native | latest | Mapas (open source) |
| Zustand | 5.x | Estado global (auth) |
| AsyncStorage | latest | Persistencia de tokens |
| Geolocation Service | latest | GPS tracking |
| React Native Gesture Handler | latest | Gestos nativos |
| React Native Reanimated | latest | Animaciones |
| React Native Safe Area Context | latest | Safe area insets |
| React Native Screens | latest | Navegación nativa |
