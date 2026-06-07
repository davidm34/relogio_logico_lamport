# Lamport Horse Race - Corrida de Cavalos Distribuída

Sistema distribuído que demonstra o funcionamento de **Relógios Lógicos de Lamport** através de uma corrida de cavalos com apostas.

## 🏗️ Arquitetura

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   cavalo1   │     │   cavalo2   │     │   cavalo3   │     │  apostador  │
│  (horse)    │     │  (horse)    │     │  (horse)    │     │  (bettor)   │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │                   │
       │    TCP/JSON       │    TCP/JSON       │    TCP/JSON       │
       │                   │                   │                   │
       └───────────────────┼───────────────────┼───────────────────┘
                           │                   │
                    ┌──────┴───────────────────┴──────┐
                    │           JUIZ (judge)           │
                    │   - Coleta eventos              │
                    │   - Ordena por Lamport          │
                    │   - Valida apostas              │
                    │   - Exibe classificação         │
                    └─────────────────────────────────┘
```

## 📁 Estrutura do Projeto

```
lamport/
├── cmd/
│   ├── judge/          # Processo Juiz
│   │   └── main.go
│   ├── horse/          # Processo Cavalo
│   │   └── main.go
│   └── bettor/         # Processo Apostador
│       └── main.go
├── pkg/
│   ├── lamport/        # Implementação do Relógio de Lamport
│   │   └── clock.go
│   └── protocol/       # Protocolo de comunicação (mensagens TCP/JSON)
│       └── message.go
├── Dockerfile          # Build multi-stage em Go
├── docker-compose.yml  # Orquestração dos containers
├── go.mod              # Módulo Go
└── README.md
```

## 🚀 Como Executar

### Pré-requisitos
- Docker e Docker Compose instalados

### Executar a corrida

```bash
# Build e execução de todos os containers
docker-compose up --build

# Para ver os logs de um processo específico
docker-compose logs -f judge
docker-compose logs -f horse1

# Para parar tudo
docker-compose down
```

## 🔧 Conceitos Demonstrados

### Relógio Lógico de Lamport
- **Evento interno**: `clock++`
- **Envio de mensagem**: `clock++`, envia `clock` junto com a mensagem
- **Recebimento de mensagem**: `clock = max(clock_local, clock_recebido) + 1`

### Validação de Apostas por Causalidade
- Uma aposta é **válida** se seu timestamp de Lamport é **menor** que o timestamp de chegada do vencedor
- Uma aposta é **inválida** se seu timestamp indica que ela foi feita **depois** do vencedor cruzar a linha (em termos de ordem causal)

### Comunicação Distribuída
- Cada processo mantém seu próprio relógio de Lamport
- Mensagens são trocadas via **TCP** com encoding **JSON**
- O juiz coleta e ordena todos os eventos pela ordem de Lamport

## 📊 Tipos de Mensagens

| Tipo      | Descrição                              | Remetente  |
|-----------|----------------------------------------|------------|
| REGISTER  | Processo se registra no juiz           | Todos      |
| START     | Sinal de largada da corrida            | Juiz       |
| ADVANCE   | Cavalo avançou uma posição             | Cavalo     |
| FINISH    | Cavalo cruzou a linha de chegada       | Cavalo     |
| BET       | Apostador faz uma aposta               | Apostador  |

## 🎯 Saída Esperada

O juiz exibe:
1. **Progresso em tempo real** de cada cavalo
2. **Apostas recebidas** com timestamps
3. **Classificação final** dos cavalos
4. **Validação das apostas** (válidas vs. inválidas por causalidade)
5. **Log completo** de eventos ordenados por timestamp de Lamport
