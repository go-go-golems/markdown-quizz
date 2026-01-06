package quizdsl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtractFormsFromContent(t *testing.T) {
	content := `
hello
<form id="quiz-a">
fields:
  - name: q1
    correct: "a"
</form>
middle
<form id='quiz-b'>
fields:
  - key: q2
    correct: [x, y]
</form>
bye
`

	forms, err := ExtractFormsFromContent(content)
	require.NoError(t, err)
	require.Len(t, forms, 2)
	require.Equal(t, "quiz-a", forms[0].FormID)
	require.Equal(t, "quiz-b", forms[1].FormID)
}

func TestExtractFormsFromContent_SkipsInvalidYAML(t *testing.T) {
	content := `
<form id="ok">
fields:
  - name: q1
    correct: "a"
</form>
<form id="bad">
fields:
  - name: q2
    correct: [unclosed
</form>
`

	forms, err := ExtractFormsFromContent(content)
	require.Error(t, err)
	require.Len(t, forms, 1)
	require.Equal(t, "ok", forms[0].FormID)
}
