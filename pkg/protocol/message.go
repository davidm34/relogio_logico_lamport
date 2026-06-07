package protocol

import (
	"encoding/json"
	"fmt"
	"net"
)

// MessageType define os tipos de mensagem trocadas entre processos.
type MessageType string

const (
	// MsgAdvance indica que um cavalo avançou uma posição.
	MsgAdvance MessageType = "ADVANCE"
	// MsgBet indica uma aposta feita pelo apostador.
	MsgBet MessageType = "BET"
	// MsgFinish indica que um cavalo cruzou a linha de chegada.
	MsgFinish MessageType = "FINISH"
	// MsgRegister indica que um processo está se registrando no juiz.
	MsgRegister MessageType = "REGISTER"
	// MsgStart indica que a corrida deve começar.
	MsgStart MessageType = "START"
	// MsgResult indica o resultado final da corrida.
	MsgResult MessageType = "RESULT"
)

// Message representa uma mensagem trocada entre processos na corrida.
type Message struct {
	Type      MessageType `json:"type"`       // Tipo da mensagem
	Sender    string      `json:"sender"`     // Nome do processo remetente
	Timestamp uint64      `json:"timestamp"`  // Timestamp do relógio de Lamport
	Position  int         `json:"position"`   // Posição atual do cavalo (para ADVANCE/FINISH)
	BetTarget string      `json:"bet_target"` // Cavalo alvo da aposta (para BET)
	Data      string      `json:"data"`       // Dados adicionais (resultados, etc.)
}

// SendMessage envia uma mensagem JSON via conexão TCP.
func SendMessage(conn net.Conn, msg Message) error {
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return fmt.Errorf("erro ao enviar mensagem: %w", err)
	}
	return nil
}

// ReceiveMessage recebe e decodifica uma mensagem JSON via conexão TCP.
func ReceiveMessage(conn net.Conn) (Message, error) {
	var msg Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil {
		return msg, fmt.Errorf("erro ao receber mensagem: %w", err)
	}
	return msg, nil
}

// ConnectTo estabelece uma conexão TCP com o endereço especificado.
func ConnectTo(address string) (net.Conn, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar em %s: %w", address, err)
	}
	return conn, nil
}

// StartListener inicia um listener TCP no endereço especificado.
func StartListener(address string) (net.Listener, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("erro ao iniciar listener em %s: %w", address, err)
	}
	return listener, nil
}
