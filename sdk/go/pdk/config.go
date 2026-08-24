// Package pdk (Plugin Development Kit) resuelve, para el autor de un
// plugin, las partes repetitivas de cumplir el Asterion Plugin Contract:
// leer la config que Asterion ya le inyectó como variables de entorno,
// loguear en un formato consistente, responder errores con una forma
// unificada, y exponer un health check. No es obligatorio usarlo — un
// plugin en Python o Rust cumple el contrato igual sin este paquete — pero
// para un plugin en Go es el camino de menor fricción.
package pdk

import (
	"os"
	"strconv"
	"strings"
)

const configEnvPrefix = "ASTERION_PLUGIN_CONFIG_"

// Config lee toda la configuración que internal/plugins.Start inyectó como
// variables de entorno ASTERION_PLUGIN_CONFIG_<CLAVE> y la devuelve con las
// claves en minúscula, tal como se declararon en config_schema.
func Config() map[string]string {
	out := map[string]string{}
	for _, kv := range os.Environ() {
		key, val, found := strings.Cut(kv, "=")
		if !found || !strings.HasPrefix(key, configEnvPrefix) {
			continue
		}
		out[strings.ToLower(strings.TrimPrefix(key, configEnvPrefix))] = val
	}
	return out
}

// ConfigString devuelve m[key], o def si no está o está vacío.
func ConfigString(m map[string]string, key, def string) string {
	if v, ok := m[key]; ok && v != "" {
		return v
	}
	return def
}

// ConfigBool interpreta m[key] como booleano ("true"/"1"/"yes" -> true),
// o devuelve def si no está o no se pudo interpretar.
func ConfigBool(m map[string]string, key string, def bool) bool {
	v, ok := m[key]
	if !ok || v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		switch strings.ToLower(v) {
		case "yes":
			return true
		case "no":
			return false
		}
		return def
	}
	return b
}

// ConfigInt interpreta m[key] como entero, o devuelve def si no está o no
// se pudo interpretar.
func ConfigInt(m map[string]string, key string, def int) int {
	v, ok := m[key]
	if !ok || v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// Name, Port y Dir devuelven la identidad que Asterion le asignó a este
// proceso al arrancarlo (ASTERION_PLUGIN_NAME/PORT/DIR) — lo mismo que ya
// documentaba el README de asterion-core, solo que sin que cada plugin
// tenga que leer os.Getenv a mano.
func Name() string { return os.Getenv("ASTERION_PLUGIN_NAME") }
func Port() string { return os.Getenv("ASTERION_PLUGIN_PORT") }
func Dir() string  { return os.Getenv("ASTERION_PLUGIN_DIR") }
