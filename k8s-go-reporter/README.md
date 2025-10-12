# K8s Go Reporter

Aplicação Go com gRPC para reportar informações de recursos do Kubernetes, seguindo princípios de Clean Architecture e SOLID.

## Características

- **gRPC**: Comunicação eficiente e type-safe
- **Clean Architecture**: Separação clara de responsabilidades
- **SOLID**: Princípios aplicados em toda a arquitetura
- **Config Builder Pattern**: Configuração flexível via variáveis de ambiente
- **OpenTelemetry**: Observabilidade com traces e métricas (opcional)
- **Hot Reload**: Desenvolvimento rápido com Tilt.dev
- **In-Cluster**: Projetado para rodar dentro do cluster Kubernetes

## Arquitetura

```
cmd/server/              # Entry point (minimal)
internal/
  ├── bootstrap/         # Application initialization
  ├── config/            # Configuration (builder pattern)
  ├── services/          # Business logic (Service interface)
  ├── grpc/              # gRPC transport layer
  └── gen/               # Generated protobuf code (not versioned)
proto/                   # Protobuf definitions
k8s/                     # Kubernetes manifests
```

### Princípios SOLID

- **SRP**: Cada pacote tem responsabilidade única
- **OCP**: Extensível via interfaces
- **LSP**: Implementações corretas das interfaces
- **ISP**: Interface `Service` pequena e focada
- **DIP**: Camada gRPC depende de abstração `services.Service`

## Pré-requisitos

- Go 1.24+
- Protocol Buffers compiler (`protoc`)
- Docker
- Kubernetes local (kind, minikube, Docker Desktop)
- [Tilt](https://docs.tilt.dev/install.html) (para desenvolvimento)
- Make

## Instalação

### 1. Instalar ferramentas necessárias

```bash
# macOS
brew install go protobuf docker kubectl tilt-dev/tap/tilt

# Linux (Ubuntu/Debian)
sudo apt-get install -y golang-go protobuf-compiler docker.io

# Verificar instalação
make check-tools
```

### 2. Instalar plugins protoc

```bash
make install-proto-tools
```

### 3. Baixar dependências

```bash
make deps
```

## Desenvolvimento Local

### Usando Makefile (Recomendado)

```bash
# Ver todos os comandos disponíveis
make help

# Build completo (proto + binary)
make all

# Apenas gerar proto
make proto

# Build local
make build

# Build e executar
make run

# Testes
make test

# Testes com cobertura
make test-coverage

# Formatar e validar código
make lint
```

### Usando Tilt (Hot Reload)

```bash
# Iniciar cluster Kubernetes local
kind create cluster
# ou
minikube start

# Iniciar Tilt
make tilt-up
# ou
tilt up
```

O Tilt irá:
1. Gerar arquivos protobuf automaticamente
2. Compilar o binário Go localmente (~1-2s)
3. Construir imagem Docker leve
4. Aplicar manifests Kubernetes
5. Configurar port-forward para `:50051`
6. Habilitar live reload em mudanças de código

**Hot Reload**: Edite qualquer arquivo `.go` e veja as mudanças em ~2-3s!

**Tilt UI**: http://localhost:10350

### Parar Tilt

```bash
make tilt-down
# ou
tilt down
```

## Configuração

A aplicação usa o **Config Builder Pattern** e é configurada via variáveis de ambiente:

### Configuração do Servidor

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `PORT` | `50051` | Porta do servidor gRPC |

### Configuração do OpenTelemetry (Opcional)

| Variável | Padrão | Descrição |
|----------|--------|-----------|
| `OTEL_ENABLED` | `false` | Habilita OpenTelemetry |
| `OTEL_SERVICE_NAME` | `k8s-go-reporter` | Nome do serviço |
| `OTEL_SERVICE_VERSION` | `unknown` | Versão do serviço |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | - | Endpoint OTLP (obrigatório se habilitado) |
| `OTEL_EXPORTER_OTLP_INSECURE` | `false` | Conexão insegura (sem TLS) |
| `OTEL_TRACES_ENABLED` | `true` | Habilita distributed tracing |
| `OTEL_METRICS_ENABLED` | `true` | Habilita métricas |

**Nota**: Ao habilitar OpenTelemetry (`OTEL_ENABLED=true`), pelo menos um de `OTEL_TRACES_ENABLED` ou `OTEL_METRICS_ENABLED` deve ser `true`.

### Exemplos de Configuração

#### Sem OpenTelemetry (padrão)

```bash
export PORT=50051
make run
```

#### Com OpenTelemetry + Jaeger (local)

```bash
# Iniciar Jaeger localmente
docker run -d --name jaeger \
  -p 4317:4317 \
  -p 16686:16686 \
  jaegertracing/all-in-one:latest

# Configurar aplicação
export PORT=50051
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=k8s-go-reporter
export OTEL_SERVICE_VERSION=v0.0.1
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_INSECURE=true
export OTEL_TRACES_ENABLED=true
export OTEL_METRICS_ENABLED=true

make run

# Acessar Jaeger UI: http://localhost:16686
```

#### Com Grafana Cloud

```bash
export OTEL_ENABLED=true
export OTEL_SERVICE_NAME=k8s-go-reporter
export OTEL_SERVICE_VERSION=v0.0.1
export OTEL_EXPORTER_OTLP_ENDPOINT=tempo-prod-10-prod-us-east-0.grafana.net:443
export OTEL_EXPORTER_OTLP_INSECURE=false
export OTEL_TRACES_ENABLED=true
export OTEL_METRICS_ENABLED=false

make run
```

#### Apenas Métricas (sem Traces)

```bash
export OTEL_ENABLED=true
export OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
export OTEL_EXPORTER_OTLP_INSECURE=true
export OTEL_TRACES_ENABLED=false
export OTEL_METRICS_ENABLED=true

make run
```

### Métricas Coletadas

Quando `OTEL_METRICS_ENABLED=true`, as seguintes métricas são coletadas:

- `grpc.server.requests.total` (Counter): Total de requisições gRPC
- `grpc.server.request.duration` (Histogram): Duração das requisições em segundos
- `grpc.server.requests.in_flight` (UpDownCounter): Requisições em processamento

### Traces Coletados

Quando `OTEL_TRACES_ENABLED=true`, todos os métodos gRPC são automaticamente instrumentados com distributed tracing usando o padrão W3C Trace Context.

## Serviços gRPC

### ListNamespaces

Lista todos os namespaces do cluster.

**Request**: `ListNamespacesRequest` (vazio)

**Response**: `ListNamespacesResponse`
```protobuf
message ListNamespacesResponse {
  repeated string namespaces = 1;
}
```

### ListPods

Lista pods de um namespace específico.

**Request**: `ListPodsRequest`
```protobuf
message ListPodsRequest {
  string namespace = 1;
}
```

**Response**: `ListPodsResponse`
```protobuf
message ListPodsResponse {
  repeated Pod pods = 1;
}

message Pod {
  string name = 1;
  string namespace = 2;
  string status = 3;
  string node_name = 4;
}
```

## Testando com grpcurl

```bash
# Instalar grpcurl
brew install grpcurl

# Listar serviços
grpcurl -plaintext localhost:50051 list

# Listar namespaces
grpcurl -plaintext localhost:50051 k8sreporter.v1.K8SReporterService/ListNamespaces

# Listar pods no namespace default
grpcurl -plaintext -d '{"namespace":"default"}' localhost:50051 k8sreporter.v1.K8SReporterService/ListPods
```

## Docker

### Build para produção

```bash
make docker-build
```

### Build para desenvolvimento

```bash
make docker-build-dev
```

## Kubernetes

### RBAC Permissions

A aplicação requer:
- ServiceAccount: `k8s-go-reporter`
- ClusterRole com permissões: `get`, `list`, `watch` em `namespaces` e `pods`

### Deploy manual

```bash
# Aplicar manifests
make k8s-apply

# Ver status
make k8s-status

# Ver logs
make k8s-logs

# Restart deployment
make k8s-restart

# Remover recursos
make k8s-delete
```

## Estrutura do Projeto

```
k8s-go-reporter/
├── cmd/server/              # Entry point (main.go)
├── internal/
│   ├── bootstrap/           # App initialization & DI
│   ├── config/              # Config with builder pattern
│   ├── services/            # Business logic
│   │   ├── service.go       # Service interface
│   │   └── k8s_service.go   # Kubernetes implementation
│   ├── grpc/                # gRPC transport layer
│   │   ├── server.go        # Server lifecycle
│   │   └── handler.go       # Request handlers
│   └── gen/                 # Generated proto (gitignored)
├── proto/                   # Protobuf definitions
│   └── k8sreporter/v1/
│       └── k8s_reporter.proto
├── k8s/                     # Kubernetes manifests
│   ├── rbac.yaml
│   ├── deployment.yaml
│   └── service.yaml
├── Dockerfile               # Production build
├── Dockerfile.dev           # Development build
├── Tiltfile                 # Tilt configuration
├── Makefile                 # Build automation
└── README.md
```

## Desenvolvimento

### Adicionar novo serviço gRPC

1. Definir no `.proto`:
```protobuf
service K8sReporterService {
  rpc NewMethod(NewRequest) returns (NewResponse);
}
```

2. Regenerar código:
```bash
make proto
```

3. Implementar no service:
```go
// internal/services/k8s_service.go
func (s *K8sService) NewMethod(ctx context.Context, ...) {...}
```

4. Implementar no handler:
```go
// internal/grpc/handler.go
func (h *Handler) NewMethod(ctx context.Context, req *pb.NewRequest) (*pb.NewResponse, error) {
    return h.service.NewMethod(ctx, ...)
}
```

### Adicionar nova configuração

```go
// internal/config/config.go
type Config struct {
    server ServerConfig
    newConfig NewConfig  // adicionar aqui
}

// internal/config/builder.go
func (b *Builder) WithNewConfig(value string) *Builder {
    b.config.newConfig.value = value
    return b
}

func (b *Builder) WithEnv() *Builder {
    // ler variável de ambiente
    if val, ok := os.LookupEnv("NEW_VAR"); ok {
        b.config.newConfig.value = val
    }
    return b
}
```

## Troubleshooting

### Erro de permissão RBAC

```bash
# Verificar RBAC
kubectl get serviceaccount k8s-go-reporter
kubectl get clusterrole k8s-go-reporter-reader
kubectl get clusterrolebinding k8s-go-reporter-binding

# Reaplicar
make k8s-delete
make k8s-apply
```

### Proto não compila

```bash
# Limpar e regenerar
make clean-proto
make proto
```

### Rebuild completo

```bash
make clean-all
make all
```

### Container em CrashLoopBackOff

```bash
# Ver logs
make k8s-logs

# Verificar recursos
kubectl describe pod -l app=k8s-go-reporter

# Restart
make k8s-restart
```

## Scripts úteis

```bash
# Limpar tudo
make clean-all

# Atualizar dependências
make deps-upgrade

# Verificar código
make lint

# Executar testes
make test

# Ver versão
make version

# Formatar código
make fmt
```

## Produção

### Build otimizado

```bash
# Build para Linux
make build-linux

# Build da imagem Docker
make docker-build

# Push da imagem
make docker-push
```

### Deploy

```bash
# Aplicar manifests
kubectl apply -f k8s/

# Verificar status
kubectl get pods -l app=k8s-go-reporter
kubectl get svc k8s-go-reporter
```

## Licença

MIT

## Contribuindo

1. Fork o projeto
2. Crie sua feature branch (`git checkout -b feature/amazing-feature`)
3. Commit suas mudanças (`git commit -m 'Add amazing feature'`)
4. Push para a branch (`git push origin feature/amazing-feature`)
5. Abra um Pull Request

## Recursos

- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://developers.google.com/protocol-buffers)
- [Kubernetes Go Client](https://github.com/kubernetes/client-go)
- [Tilt Documentation](https://docs.tilt.dev/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)