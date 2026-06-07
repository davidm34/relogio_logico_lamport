# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copia arquivos do módulo Go
COPY go.mod ./

# Copia o código fonte
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/

# Argumento para selecionar qual binário compilar
ARG SERVICE_NAME

# Compila o binário
RUN CGO_ENABLED=0 GOOS=linux go build -o /service ./cmd/${SERVICE_NAME}/

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /service .

CMD ["./service"]
