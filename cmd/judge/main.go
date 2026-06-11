package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"sync"
	"time"

	"lamport-horse-race/pkg/lamport"
	"lamport-horse-race/pkg/protocol"
)

const (
	raceDistance   = 10 // Distância da corrida (posições)
	totalHorses   = 3  // Número de cavalos esperados
	listenPort    = ":5000"
)

// Event armazena um evento recebido pelo juiz para ordenação.
type Event struct {
	Message   protocol.Message
	Received  time.Time
}

func main() {
	clock := lamport.NewClock()

	var mu sync.Mutex
	events := make([]Event, 0)
	horsePositions := make(map[string]int)     // posição atual de cada cavalo
	horseFinished := make(map[string]uint64)   // timestamp de chegada de cada cavalo
	bets := make([]protocol.Message, 0)        // apostas recebidas
	registered := 0
	raceStarted := false
	connections := make([]net.Conn, 0)

	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║     🏇 JUIZ DA CORRIDA DE CAVALOS 🏇     ║")
	fmt.Println("║       Relógios Lógicos de Lamport        ║")
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Printf("Aguardando conexões na porta %s...\n\n", listenPort)

	listener, err := protocol.StartListener(listenPort)
	if err != nil {
		log.Fatalf("Falha ao iniciar listener: %v", err)
	}
	defer listener.Close()

	// Goroutine para aceitar conexões
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Erro ao aceitar conexão: %v", err)
				continue
			}

			// Goroutine para tratar cada conexão
			go func(c net.Conn) {
				defer c.Close()
				decoder := json.NewDecoder(c)

				for {
					var msg protocol.Message
					if err := decoder.Decode(&msg); err != nil {
						return
					}

					mu.Lock()

					// Atualiza relógio de Lamport ao receber mensagem
					newTs := clock.Receive(msg.Timestamp)

					event := Event{
						Message:  msg,
						Received: time.Now(),
					}
					events = append(events, event)

					switch msg.Type {
					case protocol.MsgRegister:
						registered++
						connections = append(connections, c)
						fmt.Printf("[T=%d] 📋 %s registrado (%d/%d)\n",
							newTs, msg.Sender, registered, totalHorses+1)

						// Quando todos se registrarem, inicia a corrida
						if registered >= totalHorses+1 && !raceStarted {
							raceStarted = true
							fmt.Printf("\n[T=%d] 🏁 TODOS REGISTRADOS! INICIANDO CORRIDA!\n\n", clock.Tick())
							startMsg := protocol.Message{
								Type:      protocol.MsgStart,
								Sender:    "juiz",
								Timestamp: clock.Send(),
							}
							for _, conn := range connections {
								protocol.SendMessage(conn, startMsg)
							}
						}

					case protocol.MsgAdvance:
						horsePositions[msg.Sender] = msg.Position
						progress := ""
						for i := 0; i < msg.Position; i++ {
							progress += "█"
						}
						for i := msg.Position; i < raceDistance; i++ {
							progress += "░"
						}
						fmt.Printf("[T=%d] 🐎 %s avançou para posição %d/%d [%s]\n",
							newTs, msg.Sender, msg.Position, raceDistance, progress)

					case protocol.MsgFinish:
						horseFinished[msg.Sender] = msg.Timestamp
						horsePositions[msg.Sender] = msg.Position
						fmt.Printf("[T=%d] 🏆 %s CRUZOU A LINHA DE CHEGADA! (timestamp original: %d)\n",
							newTs, msg.Sender, msg.Timestamp)

						// Verifica se todos os cavalos terminaram
						if len(horseFinished) >= totalHorses {
							fmt.Println("\n⏳ Aguardando 2 segundos para apostas finais...")
							mu.Unlock()
							time.Sleep(2 * time.Second)
							mu.Lock()
							printResults(events, horseFinished, bets, raceDistance)
							mu.Unlock()
							// Dá tempo para output
							time.Sleep(1 * time.Second)
							os.Exit(0)
						}

					case protocol.MsgBet:
						bets = append(bets, msg)
						fmt.Printf("[T=%d] 🎰 %s apostou no %s (timestamp da aposta: %d)\n",
							newTs, msg.Sender, msg.BetTarget, msg.Timestamp)
					}

					mu.Unlock()
				}
			}(conn)
		}
	}()

	// Mantém o processo rodando
	select {}
}

// printResults exibe os resultados finais com ordenação por relógio de Lamport.
func printResults(events []Event, horseFinished map[string]uint64, bets []protocol.Message, raceDistance int) {
	fmt.Println("\n╔══════════════════════════════════════════╗")
	fmt.Println("║          📊 RESULTADO FINAL 📊            ║")
	fmt.Println("╚══════════════════════════════════════════╝")

	// Ordena eventos por timestamp de Lamport
	sort.Slice(events, func(i, j int) bool {
		if events[i].Message.Timestamp == events[j].Message.Timestamp {
			return events[i].Message.Sender < events[j].Message.Sender
		}
		return events[i].Message.Timestamp < events[j].Message.Timestamp
	})

	// Classificação dos cavalos (por timestamp de chegada)
	type HorseResult struct {
		Name      string
		Timestamp uint64
	}
	results := make([]HorseResult, 0)
	for name, ts := range horseFinished {
		results = append(results, HorseResult{Name: name, Timestamp: ts})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Timestamp < results[j].Timestamp
	})

	fmt.Println("\n🏆 CLASSIFICAÇÃO:")
	fmt.Println("─────────────────────────────────────")
	medals := []string{"🥇", "🥈", "🥉"}
	for i, r := range results {
		medal := ""
		if i < len(medals) {
			medal = medals[i]
		}
		fmt.Printf("  %s %dº lugar: %s (timestamp de chegada: %d)\n",
			medal, i+1, r.Name, r.Timestamp)
	}

	// Valida apostas
	winner := results[0].Name
	winnerFinishTs := results[0].Timestamp

	fmt.Println("\n🎰 VALIDAÇÃO DAS APOSTAS:")
	fmt.Println("─────────────────────────────────────")
	for _, bet := range bets {
		valid := bet.Timestamp < winnerFinishTs
		status := "✅ VÁLIDA"
		if !valid {
			status = "❌ INVÁLIDA (aposta após o vencedor cruzar a linha)"
		}
		acertou := ""
		if bet.BetTarget == winner && valid {
			acertou = " 🎉 ACERTOU!"
		} else if bet.BetTarget != winner && valid {
			acertou = " 😢 Errou!"
		}
		fmt.Printf("  %s apostou em %s (T_aposta=%d, T_chegada_vencedor=%d) → %s%s\n",
			bet.Sender, bet.BetTarget, bet.Timestamp, winnerFinishTs, status, acertou)
	}

	// Log de todos os eventos ordenados por Lamport
	fmt.Println("\n📜 LOG DE EVENTOS (ordenados por Relógio de Lamport):")
	fmt.Println("─────────────────────────────────────")
	for _, e := range events {
		if e.Message.Type == protocol.MsgRegister || e.Message.Type == protocol.MsgStart {
			continue
		}
		icon := "  "
		switch e.Message.Type {
		case protocol.MsgAdvance:
			icon = "🐎"
		case protocol.MsgFinish:
			icon = "🏁"
		case protocol.MsgBet:
			icon = "🎰"
		}
		detail := ""
		switch e.Message.Type {
		case protocol.MsgAdvance:
			detail = fmt.Sprintf("posição %d", e.Message.Position)
		case protocol.MsgFinish:
			detail = fmt.Sprintf("CHEGOU! posição %d", e.Message.Position)
		case protocol.MsgBet:
			detail = fmt.Sprintf("apostou em %s", e.Message.BetTarget)
		}
		fmt.Printf("  [T=%03d] %s %-10s → %s\n",
			e.Message.Timestamp, icon, e.Message.Sender, detail)
	}
	fmt.Println()
}
