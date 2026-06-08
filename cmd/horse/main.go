package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	"lamport-horse-race/pkg/lamport"
	"lamport-horse-race/pkg/protocol"
)

const (
	raceDistance = 10 // Distância total da corrida
)

func main() {
	// Configuração via variáveis de ambiente
	horseName := os.Getenv("HORSE_NAME")
	if horseName == "" {
		horseName = "cavalo1"
	}
	judgeAddr := os.Getenv("JUDGE_ADDR")
	if judgeAddr == "" {
		judgeAddr = "judge:5000"
	}

	clock := lamport.NewClock()

	fmt.Printf("🐎 %s iniciando...\n", horseName)
	fmt.Printf("   Conectando ao juiz em %s\n", judgeAddr)

	// Tenta conectar ao juiz com retries
	var conn net.Conn
	for retries := 0; retries < 30; retries++ {
		c, err := protocol.ConnectTo(judgeAddr)
		if err == nil {
			conn = c
			fmt.Printf("   ✅ Conectado ao juiz!\n\n")
			break
		}
		fmt.Printf("   ⏳ Tentativa %d - aguardando juiz...\n", retries+1)
		time.Sleep(1 * time.Second)
	}
	if conn == nil {
		log.Fatal("Não foi possível conectar ao juiz")
	}
	defer conn.Close()

	encoder := json.NewEncoder(conn)
	decoder := json.NewDecoder(conn)

	// Registra no juiz
	regMsg := protocol.Message{
		Type:      protocol.MsgRegister,
		Sender:    horseName,
		Timestamp: clock.Send(),
	}
	encoder.Encode(regMsg)
	fmt.Printf("[T=%d] 📋 Registro enviado ao juiz\n", regMsg.Timestamp)

	// Aguarda sinal de START do juiz
	fmt.Println("⏳ Aguardando sinal de largada...")
	for {
		var msg protocol.Message
		if err := decoder.Decode(&msg); err != nil {
			log.Fatalf("Erro ao receber mensagem: %v", err)
		}
		if msg.Type == protocol.MsgStart {
			clock.Receive(msg.Timestamp)
			fmt.Printf("[T=%d] 🏁 LARGADA! Corrida começou!\n\n", clock.Value())
			break
		}
	}

	// Simula a corrida
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	position := 0

	for position < raceDistance {
		// Espera aleatória entre 500ms e 2000ms simulando velocidade do cavalo
		delay := time.Duration(500+rng.Intn(1500)) * time.Millisecond
		time.Sleep(delay)

		// Avança uma posição
		position++

		// Incrementa relógio e envia avanço
		ts := clock.Send()

		if position < raceDistance {
			msg := protocol.Message{
				Type:      protocol.MsgAdvance,
				Sender:    horseName,
				Timestamp: ts,
				Position:  position,
			}
			encoder.Encode(msg)

			progress := ""
			for i := 0; i < position; i++ {
				progress += "█"
			}
			for i := position; i < raceDistance; i++ {
				progress += "░"
			}
			fmt.Printf("[T=%d] 🐎 Avancei para posição %d/%d [%s]\n",
				ts, position, raceDistance, progress)
		} else {
			// Cruzou a linha de chegada!
			msg := protocol.Message{
				Type:      protocol.MsgFinish,
				Sender:    horseName,
				Timestamp: ts,
				Position:  position,
			}
			encoder.Encode(msg)
			fmt.Printf("[T=%d] 🏆 CRUZEI A LINHA DE CHEGADA!\n", ts)
		}
	}

	fmt.Printf("\n🐎 %s terminou a corrida!\n", horseName)

	// Mantém conexão aberta por um tempo para o juiz processar
	time.Sleep(5 * time.Second)
}
