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

func main() {
	bettorName := os.Getenv("BETTOR_NAME")
	if bettorName == "" {
		bettorName = "apostador"
	}
	judgeAddr := os.Getenv("JUDGE_ADDR")
	if judgeAddr == "" {
		judgeAddr = "judge:5000"
	}

	// Cavalos disponíveis para apostar
	horses := []string{"cavalo1", "cavalo2", "cavalo3"}

	clock := lamport.NewClock()

	fmt.Println("🎰 Apostador iniciando...")
	fmt.Printf("   Conectando ao juiz em %s\n", judgeAddr)

	// Tenta conectar ao juiz com retries
	var conn net.Conn
	for retries := 0; retries < 30; retries++ {
		c, err := protocol.ConnectTo(judgeAddr)
		if err == nil {
			conn = c
			fmt.Println("   ✅ Conectado ao juiz!\n")
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
		Sender:    bettorName,
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
			fmt.Printf("[T=%d] 🏁 Corrida começou! Vou fazer minhas apostas...\n\n", clock.Value())
			break
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Faz uma aposta logo no início (deve ser válida)
	time.Sleep(time.Duration(500+rng.Intn(1000)) * time.Millisecond)
	earlyBetHorse := horses[rng.Intn(len(horses))]
	ts := clock.Send()
	earlyBet := protocol.Message{
		Type:      protocol.MsgBet,
		Sender:    bettorName,
		Timestamp: ts,
		BetTarget: earlyBetHorse,
	}
	encoder.Encode(earlyBet)
	fmt.Printf("[T=%d] 🎰 APOSTA INICIAL: Aposto no %s! (aposta rápida)\n", ts, earlyBetHorse)

	// Faz uma segunda aposta depois de algum tempo (pode ser válida ou não)
	time.Sleep(time.Duration(3000+rng.Intn(5000)) * time.Millisecond)

	// Incrementa o relógio várias vezes para simular eventos internos
	for i := 0; i < rng.Intn(5)+2; i++ {
		clock.Tick()
	}

	lateBetHorse := horses[rng.Intn(len(horses))]
	ts = clock.Send()
	lateBet := protocol.Message{
		Type:      protocol.MsgBet,
		Sender:    bettorName,
		Timestamp: ts,
		BetTarget: lateBetHorse,
	}
	encoder.Encode(lateBet)
	fmt.Printf("[T=%d] 🎰 APOSTA TARDIA: Aposto no %s! (aposta tardia)\n", ts, lateBetHorse)

	// Faz uma terceira aposta bem tarde (provavelmente inválida)
	time.Sleep(time.Duration(5000+rng.Intn(5000)) * time.Millisecond)

	for i := 0; i < rng.Intn(10)+5; i++ {
		clock.Tick()
	}

	veryLateBetHorse := horses[rng.Intn(len(horses))]
	ts = clock.Send()
	veryLateBet := protocol.Message{
		Type:      protocol.MsgBet,
		Sender:    bettorName,
		Timestamp: ts,
		BetTarget: veryLateBetHorse,
	}
	encoder.Encode(veryLateBet)
	fmt.Printf("[T=%d] 🎰 APOSTA MUITO TARDIA: Aposto no %s! (provavelmente inválida)\n",
		ts, veryLateBetHorse)

	fmt.Println("\n🎰 Apostador terminou de apostar. Aguardando resultado...")

	// Mantém conexão aberta
	time.Sleep(10 * time.Second)
	os.Exit(0)
}
