// Package ratelimit implementa el control de intentos de inicio de sesión por
// dirección de conexión. No bloquea ni modifica cuentas: solo retrasa nuevos
// intentos desde una IP mientras supera el umbral dentro de su ventana.
package ratelimit

import (
	"sync"
	"time"
)

type registro struct {
	fallos int
	inicio time.Time
	ultimo time.Time
}

// Limiter cuenta fallos autenticados por IP dentro de ventanas deslizantes
// fijas, con mapa acotado para no crecer sin límite.
type Limiter struct {
	mu        sync.Mutex
	entries   map[string]*registro
	maxFallos int
	ventana   time.Duration
	capacidad int
	ahora     func() time.Time
}

// NewLimiter crea un límite de maxFallos fallos por IP por ventana, con al
// más capacidad entradas simultáneas. La más antigua se expulsa al llenarse.
func NewLimiter(maxFallos int, ventana time.Duration, capacidad int) *Limiter {
	return &Limiter{
		entries:   make(map[string]*registro, capacidad),
		maxFallos: maxFallos,
		ventana:   ventana,
		capacidad: capacidad,
		ahora:     time.Now,
	}
}

// Allow informa si la IP puede intentar un nuevo inicio de sesión. Cuando no,
// retraso indica cuánto queda de ventana actual.
func (l *Limiter) Allow(clave string) (retraso time.Duration, permitido bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	ahora := l.ahora()
	reg, existe := l.entries[clave]
	if !existe {
		if len(l.entries) >= l.capacidad {
			l.expulsarMasAntiguo()
		}
		reg = &registro{inicio: ahora, ultimo: ahora}
		l.entries[clave] = reg
	} else {
		reg.ultimo = ahora
		if ahora.Sub(reg.inicio) >= l.ventana {
			reg.inicio = ahora
			reg.fallos = 0
		}
	}
	if reg.fallos >= l.maxFallos {
		if d := reg.inicio.Add(l.ventana).Sub(ahora); d > 0 {
			return d, false
		}
		return 0, false
	}
	return 0, true
}

// RegistrarFallo cuenta un intento fallido de autenticación para la IP.
func (l *Limiter) RegistrarFallo(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	reg, existe := l.entries[clave]
	if !existe {
		if len(l.entries) >= l.capacidad {
			l.expulsarMasAntiguo()
		}
		reg = &registro{inicio: l.ahora(), ultimo: l.ahora()}
		l.entries[clave] = reg
	}
	reg.fallos++
	reg.ultimo = l.ahora()
}

// RegistrarExito elimina el contador de la IP tras un inicio de sesión válido.
func (l *Limiter) RegistrarExito(clave string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.entries, clave)
}

func (l *Limiter) expulsarMasAntiguo() {
	var masAntiguo string
	var tocado time.Time
	for clave, reg := range l.entries {
		if masAntiguo == "" || reg.ultimo.Before(tocado) {
			masAntiguo = clave
			tocado = reg.ultimo
		}
	}
	delete(l.entries, masAntiguo)
}
