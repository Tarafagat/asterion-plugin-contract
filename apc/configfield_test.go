package apc

import "testing"

// IsSecret existe porque hay dos formas válidas de declarar un campo
// secreto y durante un tiempo el código miraba solo una. Un campo declarado
// type="secret" —la forma natural en Asterion Language— se trataba como
// público: salía en claro en 'plugin config show' y, peor, entraba en el
// .env que 'plugin export' arma para el frontend.
func TestIsSecretCuentaLasDosFormasDeDeclararlo(t *testing.T) {
	cases := []struct {
		nombre string
		field  ConfigField
		quiere bool
	}{
		{"booleano explícito", ConfigField{Key: "K", Type: "string", Secret: true}, true},
		{"type=secret", ConfigField{Key: "K", Type: "secret"}, true},
		{"las dos a la vez", ConfigField{Key: "K", Type: "secret", Secret: true}, true},
		{"string común", ConfigField{Key: "K", Type: "string"}, false},
		{"number", ConfigField{Key: "K", Type: "number"}, false},
		{"bool", ConfigField{Key: "K", Type: "bool"}, false},
		{"sin type", ConfigField{Key: "K"}, false},
	}
	for _, c := range cases {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.field.IsSecret(); got != c.quiere {
				t.Errorf("IsSecret() = %v, esperaba %v", got, c.quiere)
			}
		})
	}
}
