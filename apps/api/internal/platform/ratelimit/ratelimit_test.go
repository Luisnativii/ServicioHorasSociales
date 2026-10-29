package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func reloj(fijos ...time.Time) func() time.Time {
	idx := 0
	return func() time.Time {
		if len(fijos) == 0 {
			return time.Now()
		}
		t := fijos[idx]
		if idx < len(fijos)-1 {
			idx++
		}
		return t
	}
}

func TestLimiterPermiteYRechazaPorUmbral(t *testing.T) {
	inicio := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	l := NewLimiter(2, 5*time.Minute, 16)
	l.ahora = reloj(inicio)

	if retraso, ok := l.Allow("10.0.0.1"); retraso != 0 || !ok {
		t.Fatal("primer intento debe permitirse")
	}
	l.RegistrarFallo("10.0.0.1")
	l.RegistrarFallo("10.0.0.1")
	if retraso, ok := l.Allow("10.0.0.1"); retraso == 0 || ok {
		t.Fatal("tercer intento debe rechazarse tras dos fallos")
	} else if retraso > 5*time.Minute {
		t.Fatalf("retraso fuera de ventana: %v", retraso)
	}
}

func TestLimiterExpiraVentana(t *testing.T) {
	inicio := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	l := NewLimiter(1, 5*time.Minute, 16)
	l.ahora = reloj(inicio, inicio, inicio, inicio, inicio.Add(6*time.Minute))
	l.RegistrarFallo("10.0.0.2")
	if retraso, ok := l.Allow("10.0.0.2"); retraso == 0 || ok {
		t.Fatal("fallo dentro de ventana debe rechazarse")
	}
	if retraso, ok := l.Allow("10.0.0.2"); retraso != 0 || !ok {
		t.Fatal("ventana expirada debe permitir de nuevo")
	}
}

func TestLimiterExitoLimpiaIP(t *testing.T) {
	l := NewLimiter(1, time.Hour, 16)
	l.RegistrarFallo("10.0.0.3")
	if _, ok := l.Allow("10.0.0.3"); ok {
		t.Fatal("fallo registrado debe bloquear")
	}
	l.RegistrarExito("10.0.0.3")
	if _, ok := l.Allow("10.0.0.3"); !ok {
		t.Fatal("éxito debe limpiar el contador")
	}
}

func TestLimiterCapacidadAcutada(t *testing.T) {
	l := NewLimiter(1, time.Hour, 2)
	for i := 1; i <= 3; i++ {
		l.RegistrarFallo("10.0.0." + string(rune(i+48)))
	}
	if _, ok := l.Allow("10.0.0.9"); !ok {
		t.Fatal("mapa acotado no debe bloquear IP nueva sin fallos")
	}
}

func TestLimiterConcurrente(t *testing.T) {
	l := NewLimiter(100, time.Minute, 1024)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				l.RegistrarFallo("10.0.0.1")
				l.Allow("10.0.0.1")
			}
		}()
	}
	wg.Wait()
}

func TestLimiterIPsIndependientes(t *testing.T) {
	l := NewLimiter(1, time.Hour, 16)
	l.RegistrarFallo("10.0.0.1")
	if _, ok := l.Allow("10.0.0.1"); ok {
		t.Fatal("IP con fallo no debe pasar")
	}
	if _, ok := l.Allow("10.0.0.2"); !ok {
		t.Fatal("otra IP sin fallos debe pasar")
	}
}
