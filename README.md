# Rinha de Backend 2026 - Detecção de Fraude com Busca Vetorial (Go)

Implementação da API da Rinha de Backend 2026 para score de fraude em transações de cartão usando vetorização + busca de vizinhos mais próximos.

## Visão Geral

Este repositório implementa o módulo de fraude da arquitetura da Rinha:

1. Recebe uma transação via API HTTP.
2. Converte o payload em um vetor de 14 dimensões.
3. Busca os 5 vizinhos mais próximos em uma base de referência rotulada.
4. Calcula:
   - fraud_score = fraudes entre os 5 vizinhos / 5
   - approved = fraud_score < 0.6
5. Retorna a decisão no formato esperado pelo desafio.

## Endpoints

A API expõe os seguintes endpoints na porta 9999:

- GET /ready
  - healthcheck/readiness
  - resposta: 200 OK com body "ok"

- POST /fraud-score
  - recebe payload de transação
  - retorna:

```json
{
  "approved": false,
  "fraud_score": 0.8
}
```

## Arquitetura da Solução

A arquitetura roda com docker-compose e segue o mínimo exigido no desafio:

- 1 load balancer Nginx
- 2 instâncias da API Go
- balanceamento round-robin
- rede bridge
- porta pública 9999

### Serviços

- nginx
  - imagem: nginx:1.27-alpine
  - expõe 9999
  - encaminha para backend-1:8080 e backend-2:8080
  - limite: 0.10 CPU / 20 MB

- backend-1
  - build local (Dockerfile)
  - escuta internamente em :8080
  - limite: 0.45 CPU / 165 MB

- backend-2
  - build local (Dockerfile)
  - escuta internamente em :8080
  - limite: 0.45 CPU / 165 MB

### Orçamento total de recursos

- CPU total: 1.00
- Memória total: 350 MB

## Como a Detecção Funciona no Código

### 1) Vetorização (14 dimensões)

Implementada em internal/adapters/vectorizer/vectorizer.go, usando:

- normalization.json
- mcc_risk.json

Componentes principais do vetor:

1. amount normalizado
2. installments normalizado
3. razão amount/avg_amount do cliente
4. hora do dia
5. dia da semana
6. minutos desde última transação (ou -1)
7. distância da última transação (ou -1)
8. km_from_home
9. tx_count_24h
10. is_online
11. card_present
12. merchant desconhecido
13. risco MCC
14. avg_amount do merchant

### 2) Índice vetorial

Implementado em internal/adapters/vectorindex/bruteforce.go.

- Carrega referências de resources/references.json.gz
- Usa busca brute force com distância euclidiana ao quadrado
- Mantém top-k com heap máximo
- k = 5

### 3) Regra de decisão

Implementada em internal/application/fraud_detection_service.go e internal/domain/fraud_score.go:

- fraud_score = fraudes / vizinhos
- FraudThreshold = 0.6
- approved = score < 0.6

## Estrutura de Pastas (inspirado em hexagonal)

- cmd/api/main.go
  - bootstrap da aplicação

- internal/adapters/http
  - router.go: sobe servidor e rotas
  - handler.go: parse/validação da request e response

- internal/adapters/vectorizer
  - normalização e construção do vetor

- internal/adapters/vectorindex
  - loader da base gz e índice brute force

- internal/application
  - orquestração da detecção

- internal/domain
  - entidades, interfaces e regra de aprovação

- resources
  - arquivos de apoio e exemplos

- test
  - script de carga (k6), base de testes e resultado


## Benchmark k6

O script de carga está em test/test.js e gera o relatório em test/results.json.

### Rodar o teste

Com a stack em execução:

```bash
k6 run test/test.js
```

### Resultados atuais (test/results.json)

#### Dataset

- total: 14500
- fraudes: 4812 (33.19%)
- legítimas: 9688 (66.81%)
- edge cases: 157 (1.08%)

#### Latência

- p99: 29.77 ms

#### Qualidade de detecção

- false positives: 0
- false negatives: 0
- true positives: 4682
- true negatives: 9414
- erros HTTP: 0
- failure rate: 0%

#### Pontuação

- score_p99: 1526.28
- score_det: 3000
- score final: 4526.28

## Variáveis de Ambiente

Definidas em internal/config/config.go:

- LISTEN_ADDR (default: :9999)
- REFERENCES_PATH (default: /app/resources/references.json.gz)
- NORMALIZATION_PATH (default: /app/resources/normalization.json)
- MCC_RISK_PATH (default: /app/resources/mcc_risk.json)
- HNSW_INDEX_PATH (default: /app/resources/hnsw.bin)
- EF_SEARCH (default: 300)

Observação: apesar de existirem variáveis de HNSW, a implementação atual usa índice brute force.

## Exemplo de Requisição

Payloads de exemplo em resources/example-payloads.json.

```bash
curl -X POST http://localhost:9999/fraud-score \
  -H "Content-Type: application/json" \
  -d '{
    "id": "tx-123",
    "transaction": {"amount": 384.88, "installments": 3, "requested_at": "2026-03-11T20:23:35Z"},
    "customer": {"avg_amount": 769.76, "tx_count_24h": 3, "known_merchants": ["MERC-001"]},
    "merchant": {"id": "MERC-001", "mcc": "5912", "avg_amount": 298.95},
    "terminal": {"is_online": false, "card_present": true, "km_from_home": 13.7},
    "last_transaction": {"timestamp": "2026-03-11T14:58:35Z", "km_from_current": 18.8}
  }'
```

## Tecnologias

- Go 1.25
- net/http padrão
- Nginx (load balancer)
- Docker / Docker Compose
- k6