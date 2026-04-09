# CLAUDE.md — mcp-gateway

Gateway del protocolo MCP (Model Context Protocol). Recibe requests de `agent-executor` y los enruta al servidor MCP registrado correspondiente.

## Comandos

```bash
go build -v -o build/mcp-gateway main.go
go test -v -race -short ./...
golangci-lint run ./...
```

## Estructura

```
handlers/
  handlers.go      → Handlers HTTP de entrada para requests MCP
  middleware.go    → Middleware (auth por token, validación, body limit)

registry/
  registry.go      → Registro de servidores MCP disponibles y sus capacidades

proxy/
  proxy.go         → Proxy que enruta el request al servidor MCP destino

mcp/               → Tipos y utilidades del protocolo MCP

k8s/               → Manifests de Kubernetes para este servicio
```

## Flujo de un request

```
agent-executor → HTTP request MCP → handlers.go
  → middleware (auth SERVICE_TOKEN)
  → registry.Lookup(serverID)
  → proxy.Forward() → servidor MCP destino (ej. mcp-slack)
  → respuesta de vuelta a agent-executor
```

## Puntos clave

- **Autenticación**: usa `SERVICE_TOKEN` del env, no JWT de usuario final
- **Registro de servidores**: los servidores MCP (mcp-slack, futuros) se registran en `registry/registry.go`
- **Body limit**: middleware de tamaño máximo de request definido en `handlers/middleware.go`
- Para añadir un nuevo servidor MCP: registrarlo en `registry/registry.go` y desplegar el servicio correspondiente
