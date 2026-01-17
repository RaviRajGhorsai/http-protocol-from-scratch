package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTP11HeaderEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		expectErr   bool
		expectDone  bool
		expectN     int
		expectValue map[string]string
	}{
		// ─────────────────────────────────────────────
		// VALID CASES
		// ─────────────────────────────────────────────

		{
			name:       "single valid header",
			input:      []byte("Host: localhost\r\n"),
			expectErr:  false,
			expectDone: false,
			expectN:    len("Host: localhost\r\n"),
			expectValue: map[string]string{
				"Host": "localhost",
			},
		},

		{
			name:       "multiple headers",
			input:      []byte("A: 1\r\nB: 2\r\n"),
			expectErr:  false,
			expectDone: false,
			expectN:    len("A: 1\r\nB: 2\r\n"),
			expectValue: map[string]string{
				"A": "1",
				"B": "2",
			},
		},

		{
			name:       "end of headers",
			input:      []byte("Host: a\r\n\r\n"),
			expectErr:  false,
			expectDone: true,
			expectN:    len("Host: a\r\n\r\n"),
			expectValue: map[string]string{
				"Host": "a",
			},
		},

		{
			name:       "empty header section",
			input:      []byte("\r\n"),
			expectErr:  false,
			expectDone: true,
			expectN:    2,
		},

		{
			name:       "header with extra whitespace",
			input:      []byte("Key:    value\r\n"),
			expectErr:  false,
			expectDone: false,
			expectN:    len("Key:    value\r\n"),
			expectValue: map[string]string{
				"Key": "value",
			},
		},

		{
			name:       "duplicate headers (combined with commas)", // ← Updated name
			input:      []byte("X: a\r\nX: b\r\n"),
			expectErr:  false,
			expectDone: false,
			expectN:    len("X: a\r\nX: b\r\n"),
			expectValue: map[string]string{
				"X": "a, b", // ← Changed from "b" to "a, b"
			},
		},

		// ─────────────────────────────────────────────
		// STREAMING / TCP SPLIT CASES
		// ─────────────────────────────────────────────

		{
			name:       "incomplete header",
			input:      []byte("Host: loca"),
			expectErr:  false,
			expectDone: false,
			expectN:    0,
		},

		{
			name:       "incomplete CRLF",
			input:      []byte("Host: a\r"),
			expectErr:  false,
			expectDone: false,
			expectN:    0,
		},

		// ─────────────────────────────────────────────
		// INVALID / SECURITY CASES
		// ─────────────────────────────────────────────

		{
			name:      "non ASCII header name",
			input:     []byte("H©st: a\r\n"),
			expectErr: true,
		},

		{
			name:      "space before header name",
			input:     []byte(" Host: a\r\n"),
			expectErr: true,
		},

		{
			name:      "space inside header name",
			input:     []byte("Ho st: a\r\n"),
			expectErr: true,
		},

		{
			name:      "missing colon",
			input:     []byte("Host localhost\r\n"),
			expectErr: true,
		},

		{
			name:      "obsolete line folding (request smuggling)",
			input:     []byte("A: b\r\n c\r\n"),
			expectErr: true,
		},

		{
			name:      "null byte injection",
			input:     []byte("Host:\x00evil\r\n"),
			expectErr: true,
		},

		{
			name:      "bare LF (invalid)",
			input:     []byte("Host: a\n"),
			expectErr: true,
		},

		{
			name:      "bare CR (invalid)",
			input:     []byte("Host: a\r"),
			expectErr: false, // incomplete, not invalid yet
			expectN:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHeaders()
			n, done, err := h.Parse(tt.input)

			if tt.expectErr {
				require.Error(t, err)
				assert.Equal(t, 0, n)
				assert.False(t, done)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectDone, done)
			assert.Equal(t, tt.expectN, n)

			for k, v := range tt.expectValue {
				assert.Equal(t, v, h.Get(k))
			}
		})
	}
}

func TestHeaderMultipleValues(t *testing.T) {
	headers := NewHeaders()

	data := []byte(
		"Set-Person: lane-loves-go\r\n" +
			"Set-Person: prime-loves-zig\r\n" +
			"Set-Person: tj-loves-ocaml\r\n\r\n",
	)

	n, done, err := headers.Parse(data)

	require.NoError(t, err)
	require.True(t, done)
	assert.Equal(t, len(data), n)

	expected := "lane-loves-go, prime-loves-zig, tj-loves-ocaml"
	assert.Equal(t, expected, headers.Get("Set-Person"))
}
