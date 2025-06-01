package misc

import "testing"

func TestGenerateSlug(t *testing.T) {
	s := "Quartx glass"
	got := GenerateSlug(s)
	t.Log(got)
}
