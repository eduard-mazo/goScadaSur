package xmlcreator

import (
	"testing"
)

func TestGenerateDisplayName(t *testing.T) {
	headerMap := map[string]int{
		"ELEMENT": 0,
		"INFO":    1,
	}

	tests := []struct {
		name       string
		elementKey string
		row        []string
		expected   string
	}{
		{
			name:       "Normal element",
			elementKey: "PT_KV",
			row:        []string{"PT_KV", "Normal"},
			expected:   "PT_KV",
		},
		{
			name:       "Momentary movement",
			elementKey: "BREAKER_POS",
			row:        []string{"BREAKER_POS", "MvMoment"},
			expected:   "BREAKER POS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDisplayName(tt.elementKey, tt.row, headerMap)
			if result != tt.expected {
				t.Errorf("Esperado %s, obtenido %s", tt.expected, result)
			}
		})
	}
}
