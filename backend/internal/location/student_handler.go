package location

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/umg-bus-app/backend/internal/auth"
)

type StudentMessage struct {
	CampusID string `json:"campus_id"`
}

// HandleStudent maneja la conexión WebSocket de un estudiante.
//
// 🔁 Flujo:
//
// 1. El cliente hace una petición HTTP a este endpoint con ?campus_id=XXX.
//
// 2. Se valida que venga el campusID.
//    Si no viene → error HTTP 400.
//
// 3. Se convierte la conexión HTTP a WebSocket usando websocket.Accept.
//    → Aquí ocurre el "upgrade" de HTTP → WebSocket.
//    → A partir de este punto ya NO es HTTP normal.
//
// 4. Se mantiene la conexión abierta (defer conn.Close).
//
// 5. Se obtiene el contexto de la request (ctx),
//    que servirá para detectar cuando el cliente se desconecte.
//
// 6. (Opcional) Se obtiene la última ubicación conocida desde el Hub
//    y se envía inmediatamente al cliente (estado inicial).
//
// 7. Se llama a hub.Subscribe:
//    → Se suscribe al canal de Redis del campus
//    → Escucha mensajes en tiempo real
//    → Reenvía cada mensaje al cliente por WebSocket
//
// 8. La función queda bloqueada mientras la conexión esté activa.
//
// 9. Cuando el cliente se desconecta o ocurre un error:
//    → ctx se cancela
//    → Subscribe termina
//    → Se cierra el WebSocket
//
// ------------------------------------------------------------
// 🚀 Modelo mental:
//
// HTTP request → Upgrade a WebSocket → conexión persistente
//        ↓
//   Hub (Redis Pub/Sub)
//        ↓
//   mensajes en tiempo real → cliente
//

func HandleStudent(hub *Hub, jwtSvc *auth.JWTService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenStr := r.URL.Query().Get("token")
		claims, err := jwtSvc.Verify(tokenStr)
		if err != nil || claims.Role != "student" {
			http.Error(w, "acceso denegado", 401)
		}
		campusID := r.URL.Query().Get("campus_id")
		if campusID == "" {
			http.Error(w, "campus_id requerido", 400)
			return
		}

		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Printf("ws student accept: %v", err)
		}
		defer conn.Close(websocket.StatusNormalClosure, "desconectado del servidor")

		log.Printf("estudiante conectado a: %s", campusID)

		ctx := r.Context()

		if loc, err := hub.GetLiveLocation(ctx, campusID); err != nil {
			if err := wsjson.Write(ctx, conn, loc); err != nil {
				log.Printf("error enviando la ubicacion inicial %v", err)
				return
			}
		}

		if err := hub.Suscribe(ctx, campusID, conn); err != nil {
			log.Printf("estudiante campus %s desconectado: %v", campusID, err)
		}

	}

}
