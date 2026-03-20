// Paso 3. Definir la base de datos en memoria (Si aplica)
// Tambien seria definir las otras bases de datos que usemos.
package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates and verifies a Redis client.
// Redis is used for real-time bus position storage (low-latency reads/writes).
// No definimos el tipo de la primera, porque estamos definiendo las 2
// variables como strings, equivalente:
//  addr string, password string

func NewRedis(addr, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("Redis no responde %w", err)
	}
	return client, nil
}
