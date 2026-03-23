// Handler maneja autenticación completa:
//
// 🔐 Funcionalidades:
// - Login → valida credenciales y genera tokens
// - Refresh → genera nuevo access token usando refresh token
// - Logout → invalida sesión
//
// 🧠 Arquitectura:
//
// Cliente → Login → recibe tokens
//        ↓
// usa access_token en requests (corto)
//        ↓
// expira → usa refresh_token
//        ↓
// servidor valida en Redis/DB
//        ↓
// genera nuevo access_token
//
// 🚨 IMPORTANTE:
// - NO usa cookies → usa tokens manuales (JSON)
// - Redis → sesiones rápidas
// - DB → persistencia y auditoría

package auth

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Handler struct {
	db  *pgxpool.Pool
	rdb *redis.Client
	jwt *JWTService
}

func NewHandler(db *pgxpool.Pool, rdb *redis.Client, jwt *JWTService) *Handler {
	return &Handler{db: db, rdb: rdb, jwt: jwt}
}

// Login autentica al usuario y genera tokens.
//
// 🔁 Flujo:
//
// 1. Lee el body JSON (username, password, role).
// 2. Valida que el rol sea válido.
// 3. Consulta la base de datos según el rol:
//    - pilot → tabla pilots
//    - student → tabla students
//
// 4. Obtiene:
//    - userID
//    - campusID
//    - password_hash
//
// 5. Compara password con bcrypt:
//    → NO se guarda password plano
//    → bcrypt usa hashing seguro con salt
//
// 6. Genera:
//    - access token (corto, 15 min)
//    - refresh token (largo)
//
// 7. Hashea el refresh token con SHA-256:
//    → seguridad: nunca guardar token plano
//
// 8. Guarda sesión en Redis:
//    key: refresh:<hash>
//    value: userID|role
//    TTL: 7 días
//
// 9. Guarda sesión en DB:
//    → auditoría / control / revocación
//
// 10. Retorna tokens al cliente.

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}
	if req.Role != "pilot" && req.Role != "student" {
		return c.Status(400).JSON(fiber.Map{"error": "invalid role"})
	}

	var userID, campusID, passwordHash string
	var err error

	switch req.Role {
	case "pilot":
		err = h.db.QueryRow(c.Context(),
			`
		SELECT pilot_id, campus_id, password_hash
		FROM pilots WHERE pilot_id = $1 AND active=TRUE
	`, req.Username).Scan(&userID, &campusID, &passwordHash)

	case "student":
		err = h.db.QueryRow(c.Context(),
			`
		SELECT student_id, campus_id, password_hash
		FROM students WHERE student_id = $1 AND active=TRUE
	`, req.Username).Scan(&userID, &campusID, &passwordHash)

	}
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "credenciales invalidas"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "credenciales invalidas"})
	}

	accessToken, err := h.jwt.GenerateAccess(userID, campusID, req.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "error generando el token"})
	}
	refreshToken, err := h.jwt.GenerateRefresh(userID, req.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "error generando refresh token"})
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshToken)))
	key := fmt.Sprintf("refresh:%s", tokenHash)
	h.rdb.Set(c.Context(), key, userID+"|"+req.Role, 7*24*time.Hour)

	h.db.Exec(c.Context(),
		`
	INSERT INTO sessions(user_id, user_role, token, expires_at)
	VALUES ($1, $2, $3, $4)
	`, userID, req.Role, tokenHash, time.Now().Add(7*24*time.Hour))

	return c.JSON(LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// Refresh genera un nuevo access token usando un refresh token.
//
// 🔁 Flujo:
//
// 1. Recibe refresh_token del cliente.
// 2. Hashea el token (SHA-256).
// 3. Busca en Redis:
//    → si no existe → inválido o expirado
//
// 4. Obtiene userID y role.
// 5. Consulta DB para obtener campusID.
// 6. Genera nuevo access token.
// 7. Retorna access_token nuevo.

func (h *Handler) Refresh(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "body invalido"})
	}

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(body.RefreshToken)))
	key := fmt.Sprintf("refresh:%s", tokenHash)
	val, err := h.rdb.Get(c.Context(), key).Result()
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "refresh token invalido o expirado"})
	}
	var userID, role, campusID string
	fmt.Sscanf(val, "%s|%s", &userID, &role)

	if role == "pilot" {
		h.db.QueryRow(c.Context(),
			`
			SELECT campus_id FROM pilots WHERE pilot_id =$1
		`, userID).Scan(&campusID)
	} else {
		h.db.QueryRow(c.Context(),
			`
		SELECT campus_id FROM students WHERE student_id=$1
		`, userID).Scan(&campusID)
	}

	accessToken, err := h.jwt.GenerateAccess(userID, campusID, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "error generando token",
		})
	}
	return c.JSON(fiber.Map{"access_token": accessToken})

}

// Logout invalida un refresh token.
//
// 🔁 Flujo:
//
// 1. Recibe refresh_token.
// 2. Hashea el token.
// 3. Elimina la key en Redis.
//    → ya no se puede usar
//
// 4. Marca sesión como revocada en DB.
//    → auditoría / seguridad
//
// 5. Retorna OK.

func (h *Handler) Logout(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "body invalido"})
	}
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(body.RefreshToken)))
	key := fmt.Sprintf("refresh:%s", tokenHash)

	h.rdb.Del(c.Context(), key)
	h.db.Exec(c.Context(),
		`
	UPDATE sessions SET revoked_at= now(), revoke_reason='logout'
	WHERE token_hash=$1
	`, tokenHash)

	return c.JSON(fiber.Map{
		"ok": true,
	})

}
