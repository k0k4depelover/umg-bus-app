package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims define la información que se guardará dentro del JWT.
//
// 🧠 Contenido del token:
// - UserID: identifica al usuario
// - CampusID: a qué campus pertenece
// - Role: rol del usuario ("pilot" o "student")
// - RegisteredClaims: campos estándar del JWT (expiración, emisión, etc.)
//
// 📦 Esto es el "payload" del token (lo que viaja dentro del JWT)
type Claims struct {
	UserID   string `json:"user_id"`
	CampusID string `json:"campus_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTService encapsula la lógica de generación y verificación de tokens.
//
// 🔐 Usa un "secret" para firmar y validar tokens.
// Solo quien tenga este secret puede generar tokens válidos.
type JWTService struct {
	secret []byte
}

// NewJWTService crea una nueva instancia del servicio JWT.
//
// 🔁 Flujo:
// - Recibe el secret como string
// - Lo convierte a []byte (requerido por la librería)
// - Retorna el servicio listo para usar
func NewJWTService(secret string) *JWTService {
	return &JWTService{secret: []byte(secret)}
}

// GenerateAccess crea un JWT firmado para un usuario.
//
// 🔁 Flujo:
// 1. Se construyen los "claims" (datos del usuario).
// 2. Se agregan claims estándar:
//   - ExpiresAt → el token expira en 15 minutos
//   - IssuedAt → momento en que fue creado
//
// 3. Se crea el token con algoritmo HMAC (HS256).
// 4. Se firma usando el secret.
// 5. Se retorna el token como string.
//
// 📦 Salida:
// Un string tipo:
// "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
//
// 🧠 Nota:
// Este token luego se envía al cliente (frontend)
// y se usa en cada request (Authorization header o WebSocket)
func (s *JWTService) GenerateAccess(userID, campusID, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		CampusID: campusID,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Se crea el token con el método de firma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Se firma el token con el secret
	return token.SignedString(s.secret)
}

// GenerateRefresh crea un refresh token de larga duración (7 días).
// No incluye CampusID ya que al refrescar se consulta de la DB.
func (s *JWTService) GenerateRefresh(userID, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// Verify valida un JWT recibido y extrae sus datos.
//
// 🔁 Flujo:
// 1. Se recibe el token como string.
// 2. Se intenta parsear y validar:
//   - Se verifica que el algoritmo sea HMAC (seguridad)
//   - Se valida la firma con el secret
//
// 3. Se extraen los claims (payload).
// 4. Se valida que el token sea correcto y no esté expirado.
// 5. Se retornan los datos del usuario.
//
// 📥 Entrada:
// "Bearer eyJhbGciOi..."
//
// 📤 Salida:
// Claims con:
// - UserID
// - CampusID
// - Role
//
// ❌ Casos de error:
// - Firma inválida
// - Token expirado
// - Token mal formado
func (s *JWTService) Verify(tokenStr string) (*Claims, error) {

	// Parsea el token y valida firma + estructura
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {

		// Verifica que el método de firma sea HMAC (seguridad)
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("metodo de firma inesperado")
		}

		return s.secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extrae los claims del token
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("token invalido")
	}

	return claims, nil
}
