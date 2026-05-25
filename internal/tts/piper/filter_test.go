package piper

import (
	"testing"
)

func TestDefaultTextFilter(t *testing.T) {
	f := &DefaultTextFilter{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty text",
			input:    "",
			expected: "",
		},
		{
			name:     "plain text no changes",
			input:    "El dragón rojo exhala fuego.",
			expected: "El dragón rojo exhala fuego.",
		},
		{
			name: "single markdown table",
			input: `El dragón rojo exhala fuego.
| Stat | Valor |
|------|-------|
| HP   | 250   |
| AC   | 19    |
El dragón ruge triunfante.`,
			expected: "El dragón rojo exhala fuego.\nEl dragón ruge triunfante.",
		},
		{
			name: "table inline in narrative",
			input: `La batalla comienza.
| Arma    | Daño |
|---------|------|
| Espada  | 1d8  |
El enemigo cae derrotado.`,
			expected: "La batalla comienza.\nEl enemigo cae derrotado.",
		},
		{
			name:     "thinking block single line",
			input:    `El salón está oscuro.<thinking>razonamiento interno</thinking>Una antorcha parpadea.`,
			expected: "El salón está oscuro.Una antorcha parpadea.",
		},
		{
			name: "thinking block multiline",
			input: `El salón está oscuro.
<thinking>
Voy a describir la escena con detalle.
Necesito mencionar las antorchas.
</thinking>
Una antorcha parpadea en la pared.`,
			expected: "El salón está oscuro.\n\nUna antorcha parpadea en la pared.",
		},
		{
			name: "multiple thinking blocks",
			input: `Texto uno.
<thinking>primero</thinking>
Texto dos.
<thinking>segundo</thinking>
Texto tres.`,
			expected: "Texto uno.\n\nTexto dos.\n\nTexto tres.",
		},
		{
			name: "table and thinking combined",
			input: `El guardia se acerca.
<thinking>NPC: Guardia de la puerta</thinking>
| Stat | Valor |
|------|-------|
| HP   | 30    |
"¡Alto! ¿Quién va?"`,
			expected: "El guardia se acerca.\n\n\"¡Alto! ¿Quién va?\"",
		},
		{
			name:     "table separator line only",
			input:    "Texto antes.\n|------|-------|\nTexto después.",
			expected: "Texto antes.\nTexto después.",
		},
		{
			name:     "line with pipe not at start preserved",
			input:    `El | palo está roto.`,
			expected: `El | palo está roto.`,
		},
		{
			name:     "thinking with no closing tag preserved",
			input:    `Texto <thinking> sin cerrar.`,
			expected: `Texto <thinking> sin cerrar.`,
		},
		{
			name: "multiple blank lines collapsed",
			input: `Uno.



Dos.`,
			expected: "Uno.\n\nDos.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := f.Filter(tt.input)
			if got != tt.expected {
				t.Errorf("Filter() = %q, want %q", got, tt.expected)
			}
		})
	}
}
