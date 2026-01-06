package scoring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalculate_ScalarCorrect(t *testing.T) {
	def := map[string]any{
		"fields": []any{
			map[string]any{"name": "q1", "correct": "a"},
			map[string]any{"name": "q2", "correct": 2},
		},
	}

	res, err := Calculate(def, map[string]any{"q1": "a", "q2": float64(2)})
	require.NoError(t, err)
	require.Equal(t, 2, res.Score)
	require.Equal(t, 2, res.MaxScore)
}

func TestCalculate_CheckboxExactMatch(t *testing.T) {
	def := map[string]any{
		"fields": []any{
			map[string]any{"key": "q", "correct": []any{"x", "y"}},
		},
	}

	res, err := Calculate(def, map[string]any{"q": []any{"y", "x"}})
	require.NoError(t, err)
	require.Equal(t, 1, res.Score)
	require.Equal(t, 1, res.MaxScore)

	res, err = Calculate(def, map[string]any{"q": []any{"x"}})
	require.NoError(t, err)
	require.Equal(t, 0, res.Score)
	require.Equal(t, 1, res.MaxScore)
}

func TestCalculate_NullCorrectRequiresExplicitNull(t *testing.T) {
	def := map[string]any{
		"fields": []any{
			map[string]any{"name": "q", "correct": nil},
		},
	}

	res, err := Calculate(def, map[string]any{})
	require.NoError(t, err)
	require.Equal(t, 0, res.Score)
	require.Equal(t, 1, res.MaxScore)

	res, err = Calculate(def, map[string]any{"q": nil})
	require.NoError(t, err)
	require.Equal(t, 1, res.Score)
	require.Equal(t, 1, res.MaxScore)
}
