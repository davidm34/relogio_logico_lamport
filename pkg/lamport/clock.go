package lamport

import "sync"

// Clock implementa um relógio lógico de Lamport.
// É thread-safe para uso concorrente.
type Clock struct {
	mu    sync.Mutex
	value uint64
}

// NewClock cria um novo relógio de Lamport inicializado em 0.
func NewClock() *Clock {
	return &Clock{value: 0}
}

// Tick incrementa o relógio local (evento interno).
// Retorna o novo valor do timestamp.
func (c *Clock) Tick() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	return c.value
}

// Send incrementa o relógio e retorna o timestamp para envio de mensagem.
// Regra de Lamport: antes de enviar, incrementar o relógio.
func (c *Clock) Send() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	return c.value
}

// Receive atualiza o relógio ao receber uma mensagem.
// Regra de Lamport: clock = max(local, recebido) + 1
func (c *Clock) Receive(received uint64) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	if received > c.value {
		c.value = received
	}
	c.value++
	return c.value
}

// Value retorna o valor atual do relógio sem incrementar.
func (c *Clock) Value() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}
