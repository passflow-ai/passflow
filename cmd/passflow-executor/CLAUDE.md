# CLAUDE.md — agent-executor

Motor de ejecución de agentes de IA. Consume jobs del Redis Stream, ejecuta el loop ReAct y reporta resultados a passflow-api.

## Comandos

```bash
go build -v -o build/agent-executor main.go
go test -v -race -short ./...
go test -v -run TestFunctionName ./path/to/package/...
golangci-lint run ./...
```

## Estructura

```
engine/       → Modos de ejecución del agente
  react.go    → Loop ReAct (LLM + tool calls hasta max_iterations o sin más tool calls)
  oneshot.go  → Modo one-shot (single LLM call)
  helpers.go  → Utilidades compartidas del motor
  options.go  → Opciones de configuración de ejecución
  result.go   → Tipos de resultado

llm/          → Capa de proveedores LLM
  factory/    → Factory multi-proveedor
  anthropic/  → Cliente Anthropic
  openai/     → Cliente OpenAI
  bedrock/    → Cliente AWS Bedrock
  gemini/     → Cliente Google Gemini
  azure/      → Cliente Azure OpenAI
  registry/   → Registro de proveedores
  client.go   → Interfaz común de cliente LLM
  resilient.go → ResilientClient con fallback automático entre proveedores

tools/        → Definiciones y ejecutor de tools (HTTP tools, integration tools)
mcp/          → Cliente del protocolo MCP (se conecta a mcp-gateway)
job/          → Specs de definición de jobs
reporter/     → Reporta resultados a passflow-api
pkg/          → Utilidades internas
acf/          → ACF (Agent Coordination Framework)
config/       → Configuración del servicio
```

## Patrones clave

**Loop ReAct** (`engine/react.go`):
```
1. LLM call con contexto + tools disponibles
2. Si hay tool calls → ejecutar tools → añadir resultado al contexto → repetir
3. Si no hay tool calls o se alcanza max_iterations → fin
```

**LLM Factory** (`llm/factory/`):
- Crea clientes específicos por proveedor basado en config del job
- `ResilientClient` (`llm/resilient.go`) envuelve cualquier cliente con fallback configurable
- Registro de proveedores en `llm/registry/`

**Ciclo de vida de un job**:
```
Redis Stream → Dequeue job → engine.Execute() → reporter.Report() → passflow-api
```

## Nota de despliegue

`agent-executor` NO se despliega como Deployment de K8s. El orquestador de `passflow-api` lo crea como pod dinámicamente usando un ConfigMap template. Los manifests están en `k8s/`.
