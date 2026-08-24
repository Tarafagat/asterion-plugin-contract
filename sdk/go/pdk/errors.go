package pdk

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse es la forma unificada de error que se espera que un plugin
// devuelva — así el dashboard puede mostrar el mismo componente de error
// sin importar qué plugin respondió.
type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail,omitempty"`
}

// WriteError escribe una respuesta de error con la forma de ErrorResponse.
func WriteError(w http.ResponseWriter, status int, message, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: message, Detail: detail})
}
