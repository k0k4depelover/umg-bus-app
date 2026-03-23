/*

	Paso no.9: Hub de ubicaciones
	Publicar la ubicacion a un canal al que los estudiantes se suscriben.
	Metodo de tipo Hub para acceder a todos los pilotos pertenecientes a un campus.

*/
/*

Concepto previo.
func (h *Hub) GetLiveLocation(ctx context.Context, campusID string) (*LiveLocation, error) {}

En funciones que tienen un tercer parametro, el primer parametro, antes del nombre de la funcion indica que es un metodo
que pertenece a la estructura HUB
-----------------------------------------------------------------------


Un hub es un punto central que recibe mensajes, que es justo lo que estamos
haciendo en este archivo. Esta es la torre de control, define quien necesita
los datos, y los reenvia a clientes, y usuarios.

Un multiplexor (MUX) es un circuito combinacional que selecciona una de
varias señales de entrada y la dirige hacia una única salida.


Lat -> Latitud (norte o sur)
Lng -> Longitud (este o oeste)
Bearing -> Angulo de direccion (0, 360),define la direccion del movimiento.const
Speed -> Que tan rapido se mueve en la direccion



*/

package location

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/coder/websocket"
	"github.com/redis/go-redis/v9"
)

type PilotPing struct {
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Bearing float64 `json:"bearing"`
	Speed   float64 `json:"speed"`
}

type LiveLocation struct {
	PilotoID  string    `json:"pilot_id"`
	CampusID  string    `json:"campus_id"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Bearing   float64   `json:"bearing"`
	Speed     float64   `json:"speed"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Hub struct {
	rdb *redis.Client
}

func NewHub(rdb *redis.Client) *Hub {
	return &Hub{rdb: rdb}
}

/*

CONCEPTOS PREVIOS:
Inyeccion de eventos en el sistema:
Acto de introducir un dato o mensaje al flujo de ejecucion
para que el sistema reaccione a el.
Dejamos que un evento llegue y deja que los componentes
reaccionen.

ctx -> Un paquete de GO para el control de flujo y la transmision
de metadatos, sirve para controlar timeouts, y liberar recursos
Ademas previene la propapagacion, lo que quiere decir que si una
funcion llamaba a 5 mas, envia una señal para liberar esas otras
5 funciones.

----------------------------------------------------------------------
Guarda en redis y publica en el canal de campus
======================================================================


Esta funcion es el componente que maneja el tiempo real.
Aca estamos inyectando un evento al sistema, ademas utilizamos
ctx para manejar el control de cancelacion y timeouts, identificadores
del piloto y el campus, y un ping que viene del cliente.
Recibimos esos datos y los manejamos, normalizamos y añadimos
campos extras de control.

Una vez transformados los datos definimos una clave de Redis
mediante un namespace jerarquico, en este caso:
"campus:live:2:4"
Despues creamos un pipeline en la que mandamos un contexto,
mandamos la key, y un tiempo de vida, reduciendo la latencia,
dado agrupa multiples comandos para enviarlos en un solo roadtrip.
Despues utilizamos un Hash que es un objeto de tipo diccionario
para guardar en memoria la ultima ubicacion del piloto,
sin depender del flujo de tiempo real.

Despues ejecutamos el pipeline en el if.
Despues serializamos el objeto, que basicamente significa
convertirlo en JSON, esto es necesario ya que Redis maneja bytes
no estructuras, este payload es lo que vamos a distrubuir como evento.

Despues creamos un canal Pub/Sub, para que un cliente unicamente
reciba eventos de ese campus, creando un sistema de mensajeria entre
procesos.

Finalmente publicamos el evento, enviando un mensaje a Redis Pub/Sub.
Cualquier parte suscrita recibira este mensaje inmediatamente.

*/

func (h *Hub) PublishLocation(ctx context.Context, pilotID, campusID string, ping PilotPing) error {
	loc := LiveLocation{
		PilotoID: pilotID,
		CampusID: campusID,
		Lat:      ping.Lat,
		Lng:      ping.Lng,
		Bearing:  ping.Bearing,
		Speed:    ping.Speed,
	}
	key := fmt.Sprintf("campus:live:%s:%s", campusID, pilotID)
	pipe := h.rdb.Pipeline()
	//pipeline de Redis
	// Guardar estado actual con TTL 30s, con un hash
	// de redis
	pipe.HSet(ctx, key,
		"pilot_id", loc.PilotoID,
		"lat", loc.Lat,
		"lng", loc.Lng,
		"bearing", loc.Bearing,
		"speed", loc.Speed,
		"updated_at", loc.UpdatedAt.Unix(),
	)

	pipe.Expire(ctx, key, 30*time.Second)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis HSet: %w", err)
	}
	payload, _ := json.Marshal(loc)
	channel := fmt.Sprintf("campus:channel:%s", campusID)
	return h.rdb.Publish(ctx, channel, payload).Err() // .Err() -> Extrae el error de la ejecucion
}

/*


--------------------------------------------------------------------------------
Obtiene la ubicacion desde Redis
=================================================================================


CONCEPTOS PREVIOS:
Redis KEYS:
Permite buscar claves por patron usando comodines (*), pero no es eficiente
en produccion porque recorre toda la base de datos. Se usa aqui con fines
didacticos.

Redis Hash:
Estructura tipo diccionario (mapa clave-valor) donde almacenamos los datos
de un piloto (lat, lng, speed, etc).

strconv.ParseFloat:
Convierte valores string (como los devuelve Redis) a float64 para poder
trabajarlos en Go.

----------------------------------------------------------------------

Esta funcion consulta Redis para obtener la ubicacion en vivo de un piloto
dentro de un campus.

Primero se construye un patron de busqueda usando el mismo namespace
jerarquico definido anteriormente: "campus:live:{campusID}:*".
Esto permite obtener todas las keys de pilotos activos en ese campus.

Luego se consulta Redis usando KEYS, lo que devuelve una lista de claves
que coinciden con ese patron.

Si ocurre un error o no hay resultados, se retorna un error indicando que
no hay pilotos activos.

Luego se selecciona una key (keys[0]) y se usa HGetAll para obtener todos
los campos del hash almacenado en Redis. Esto devuelve un mapa de
[string]string.

Dado que Redis almacena todo como string, se convierten los valores
numericos (lat, lng, bearing, speed) a float64 usando strconv.ParseFloat.

Finalmente se reconstruye el objeto LiveLocation con los datos obtenidos
y se retorna al caller.

*/

func (h *Hub) GetLiveLocation(ctx context.Context, campusID string) (*LiveLocation, error) {
	pattern := fmt.Sprintf("campus:live:%s:*", campusID)

	keys, err := h.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("No hay pilotos activos en el campus: %s", campusID)
	}

	vals, err := h.rdb.HGetAll(ctx, keys[0]).Result()
	if err != nil {
		return nil, err
	}

	lat, _ := strconv.ParseFloat(vals["lat"], 64)
	lng, _ := strconv.ParseFloat(vals["lng"], 64)
	bearing, _ := strconv.ParseFloat(vals["bearing"], 64)
	speed, _ := strconv.ParseFloat(vals["speed"], 64)

	loc := &LiveLocation{
		PilotoID: vals["pilot_id"],
		CampusID: campusID,
		Lat:      lat,
		Lng:      lng,
		Bearing:  bearing,
		Speed:    speed,
	}

	return loc, nil
}

/*
PASO 11.
Ya tenemos un handler para el repository de REDIS, el que maneja las peticiones
HTTP que llegan, en este archivo habiamos dicho que manejamos la entrada de datos,
estas funciones se llaman cuando se quieren subir datos a canales, guardamos las keys
informacion que queremos guardar para mostrarla despues, y canales para la suscripcion.
--------------------------------------------------------------------------------------

// Subscribe maneja la suscripción en tiempo real de un cliente (WebSocket)
// a un canal de Redis asociado a un campus.
//
//  Flujo normal:
// 1. El cliente se conecta por WebSocket y envía el campusID.
// 2. Esta función se suscribe al canal de Redis: "campus:channel:{campusID}".
// 3. Mientras el cliente esté conectado:
//    - Espera mensajes publicados en Redis.
//    - Cada mensaje recibido se reenvía inmediatamente al cliente.
// 4. Este proceso continúa indefinidamente hasta que el cliente se desconecta
//    o ocurre un error.
//
// ⚙️ Concurrencia:
// Se usa `select` para esperar múltiples eventos al mismo tiempo:
// - Mensajes desde Redis (channel `ch`).
// - Cancelación del contexto (`ctx.Done()`).
//
// 🧩 Casos manejados:
// - ✔ Flujo normal:
//     Llegan mensajes → se envían al cliente → loop infinito.
//
// - ❌ Canal cerrado (`ok == false`):
//     Redis dejó de enviar datos → se termina.
//
// - ❌ Error en WebSocket:
//     Cliente desconectado o fallo de red → se retorna error.
//
// - ❌ Contexto cancelado (`ctx.Done()`):
//     El cliente se fue o se canceló la operación → salida limpia.
//
// 🧠 Nota:
// Esta función es bloqueante y vive durante toda la conexión.
//
// ------------------------------------------------------------
// 📚 GUÍA DE NOTACIONES (Go concurrency)
//
// 🔹 channel (chan)
//   Es una “tubería” para enviar/recibir datos entre goroutines.
//
// 🔹 <- (operador de channel)
//
//   📥 Recibir:
//     msg := <-ch
//     → espera hasta que llegue un dato desde el canal
//
//   📥 Recibir con estado:
//     msg, ok := <-ch
//     → ok == false → canal cerrado
//
//   📤 Enviar:
//     ch <- dato
//     → envía un dato al canal
//
// 🔹 select
//   Permite esperar múltiples operaciones de channel al mismo tiempo.
//
//   select {
//     case msg := <-ch:
//         // llegó un mensaje
//     case <-ctx.Done():
//         // cancelación
//   }
//
//   → ejecuta el primer caso que esté listo (no bloquea todos)
//
// 🔹 ctx (context.Context)
//   Maneja cancelación, timeouts y ciclo de vida.
//
//   <-ctx.Done()
//   → señal de “terminar operación”
//
// 🔹 defer
//   Ejecuta una instrucción al final de la función (cuando termina).
//
//   defer sub.Close()
//   → asegura liberar recursos aunque haya error
//
// 🔹 for + select (loop reactivo)
//
//   for {
//     select { ... }
//   }
//
//   → patrón típico para:
//     - sistemas en tiempo real
//     - sockets
//     - listeners
//
// ------------------------------------------------------------
// 🚀 Modelo mental:
//
// Redis → (channel Go) → select → WebSocket → cliente
//
// Esta función actúa como puente en tiempo real entre backend y frontend.
*/

func (h *Hub) Suscribe(ctx context.Context, campusID string, conn *websocket.Conn) error {
	channel := fmt.Sprintf("campus:channel:%s", campusID)
	sub := h.rdb.Subscribe(ctx, channel)
	defer sub.Close()

	ch := sub.Channel()

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if err := conn.Write(ctx, websocket.MessageText, []byte(msg.Payload)); err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}
