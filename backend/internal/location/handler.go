/*

Paso no.10, agregar el handler del hub, el cual utiliza el codigo del
archivo hub.go

*/

package location

import (
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/umg-bus-app/backend/internal/auth"
)

/*
Conceptos tecnicos previos:
Una funcion que retorna otra funcion, en este caso creamos una funcion HandlePilot, la cual recibe un
puntero a la estructura Hub, y retorna una funcion que retorna un handler http.

WebSocket:
Protocolo que permite una conexión persistente entre cliente y servidor,
a diferencia de HTTP que es request/response. Aquí el cliente (piloto)
envía datos continuamente sin volver a abrir conexión.

wsjson.Read:
Función que lee directamente JSON desde el WebSocket y lo deserializa
en un struct de Go. Internamente convierte bytes → struct.

Bucle infinito (for):
Mantiene la conexión viva y escuchando datos constantemente hasta que
el cliente se desconecte o ocurra un error.

----------------------------------------------------------------------

Esta función crea un handler HTTP que luego se convierte en una conexión
WebSocket para recibir datos de ubicación en tiempo real desde un piloto.

Primero se extraen los parámetros pilotID y campusID desde la URL.
En un sistema real estos vendrían de un JWT, pero aquí se usan query params
para simplificar pruebas.

Se validan estos valores, ya que son necesarios para identificar al piloto
y su contexto dentro del sistema. Si faltan, se responde con error HTTP.

Luego se realiza el upgrade de la conexión HTTP a WebSocket usando
websocket.Accept. A partir de este punto la conexión deja de ser HTTP
y pasa a ser un canal persistente bidireccional.

Se usa defer para asegurar que la conexión se cierre correctamente cuando
el piloto se desconecte o ocurra un error.

Se obtiene el contexto de la request, el cual se reutiliza para controlar
cancelación y propagación en operaciones posteriores.

Luego entra en un bucle infinito donde:

1. Se espera un mensaje del cliente (piloto)
2. Se lee como JSON y se convierte a struct PilotPing
3. Si falla, se asume desconexión y se termina la función
4. Si es exitoso, se envía el dato al sistema mediante PublishLocation

PublishLocation se encarga de:
- Guardar el estado en Redis
- Publicar el evento en Pub/Sub

Esto desacopla completamente la recepción de datos del envío a clientes,
permitiendo escalar el sistema.



---------------------------------------------------------------------
1. Flujo real completo (como sistema)

1. Cliente (piloto) → abre conexión WebSocket
2. Envía pings constantemente (lat, lng, etc.)
3. HandlePilot recibe esos pings
4. Llama a PublishLocation

5. PublishLocation hace DOS cosas:
   A) Guarda estado actual en Redis (con TTL)
   B) Publica evento en Redis Pub/Sub

6. Hub (suscriptor):
   - Escucha Redis Pub/Sub
   - Reenvía a clientes (usuarios)

7. Cliente final ve movimiento en tiempo real

*/

func HandlePilot(hub *Hub, jwtSvc *auth.JWTService) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			http.Error(w, "Token requerido", 401)
			return
		}
		claims, err := jwtSvc.Verify(tokenStr)
		if err != nil {
			http.Error(w, "token invalido", 401)
			return
		}

		if claims.Role != "pilot" {
			http.Error(w, "acceso denegado", 403)
		}
		pilotID := r.URL.Query().Get("pilot_id")
		campusID := r.URL.Query().Get("campus_id")

		if pilotID == "" || campusID == "" {
			http.Error(w, "pilot_id y campus_id son requeridos para la peticion", http.StatusBadRequest)
			return
		}
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns: []string{"*"},
		})
		if err != nil {
			log.Printf("WS accept error: %v", err)
			return
		}
		defer conn.Close(websocket.StatusNormalClosure, "piloto desconectado")
		log.Printf("piloto %s conectado al campus: %s", pilotID, campusID)

		ctx := r.Context()

		for {
			var ping PilotPing
			if err := wsjson.Read(ctx, conn, &ping); err != nil {
				log.Printf("piloto %s desconectado: %v", pilotID, err)
				return
			}
			if err := hub.PublishLocation(ctx, pilotID, campusID, ping); err != nil {
				log.Printf("error enviando ubicacion en tiempo real.")
			}

		}

	}
}
