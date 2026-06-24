package llm

import "testing"

func TestIsTemplateText(t *testing.T) {
	if !IsTemplateText("Migrando mais um serviço para Go hoje") {
		t.Fatal("expected template marker match")
	}
	if IsTemplateText("Hoje compartilhei aprendizados sobre liderança técnica no time.") {
		t.Fatal("expected non-template text")
	}
}

func TestLooksPortuguese(t *testing.T) {
	if !LooksPortuguese("Ótima reflexão sobre carreira em tecnologia.") {
		t.Fatal("expected PT-BR heuristics to pass")
	}
	if LooksPortuguese("ok") {
		t.Fatal("expected short ascii to fail")
	}
}
