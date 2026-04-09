# CLAUDE.md — channels-service

Servicio de triggers de eventos. Normaliza eventos de entrada (cron, Slack, email, webhooks) y los despacha hacia el Redis Stream según reglas configuradas con expresiones CEL.

## Comandos

```bash
go build -v -o build/channels-service main.go
go test -v -race -short ./...
golangci-lint run ./...
```

## Estructura

```
input/              → Handlers de entrada, uno por tipo de canal
  cron.go           → Triggers por cron schedule
  email.go          → Lectura IMAP, normalización de emails
  slack.go          → Eventos de Slack (mensajes, menciones)
  webhook.go        → HTTP webhooks entrantes
  processed_set.go  → Deduplicación de eventos ya procesados

trigger/            → Dispatcher de eventos y matching de reglas
domain/             → Modelos e interfaces del dominio
middleware/         → Middleware HTTP (auth, logging)
output/             → Acciones de salida tras match
store/              → Persistencia de estado
config/             → Configuración del servicio
```

## Flujo de un evento

```
Input Handler (cron / slack / email / webhook)
  → normaliza a Event{Channel, WorkspaceID, Fields, Raw}
  → trigger.Dispatcher.Dispatch(event)
    → evalúa TriggerRules con expresiones CEL
      → si match → fire action (enqueue al Redis Stream de chronos-api)
```

## Patrones clave

- **Un handler por canal**: cada archivo en `input/` es independiente y auto-contenido
- **CEL expressions**: las reglas de trigger se evalúan con el paquete `cel/` compartido del monorepo
- **Deduplicación**: `input/processed_set.go` evita procesar el mismo evento dos veces (crítico para email/IMAP)
- **1 réplica en K8s**: a diferencia de otros servicios, channels-service corre con una sola réplica para evitar procesamiento duplicado de eventos

## Agregar un nuevo tipo de canal

1. Crear `input/{canal}.go` implementando la interfaz de handler
2. Registrar el handler en `main.go`
3. Añadir tests en `input/{canal}_test.go`
