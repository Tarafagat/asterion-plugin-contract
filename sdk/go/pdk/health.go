package pdk

import (
	"encoding/json"
	"net/http"
)

// Status son los tres estados que el Asterion Plugin Contract reconoce para
// un health check — los mismos tres que ya usa internal/plugins.Status en
// asterion-core (running se mapea a Healthy, unhealthy a Unhealthy).
type Status string

const (
	Healthy   Status = "healthy"
	Degraded  Status = "degraded"
	Unhealthy Status = "unhealthy"
)

// CheckFunc evalúa el estado real del plugin en el momento en que se le
// pregunta (ej.: ¿puedo escribir en mi directorio de datos?, ¿responde la
// API externa que envuelvo?). Detail es un mensaje humano opcional.
type CheckFunc func() (Status, string)

// HealthHandler arma el handler HTTP para el health_path declarado en
// plugin.yaml, a partir de una CheckFunc provista por el plugin. Devuelve
// 200 para healthy/degraded (el proceso está vivo y puede responder, solo
// degradado) y 503 para unhealthy — internal/plugins.waitHealthy en
// asterion-core solo mira el código HTTP, así que ese contrato (2xx=vivo)
// es lo único que realmente importa para el arranque.
func HealthHandler(check CheckFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, detail := check()
		code := http.StatusOK
		if status == Unhealthy {
			code = http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": string(status),
			"detail": detail,
		})
	}
}

// SimpleHealthHandler es un health check que siempre reporta healthy — para
// el plugin más simple posible, o como punto de partida del scaffold.
func SimpleHealthHandler() http.HandlerFunc {
	return HealthHandler(func() (Status, string) { return Healthy, "" })
}
