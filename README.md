# Notification Service

Servicio en Go que consume mensajes desde una cola de RabbitMQ y los procesa.

## Requisitos

- Go 1.24.4
- RabbitMQ accesible desde la URL configurada

## Variables de entorno

Copiá el archivo de ejemplo y ajustá los valores según tu entorno:

```bash
cp .env.example .env
```

Variables disponibles:

- `RABBIT_URL`: URL de conexión a RabbitMQ.
- `RABBIT_QUEUE`: nombre de la cola a consumir.

## Levantar el proyecto

1. Instalá las dependencias del proyecto:

```bash
go mod download
```

2. Creá tu archivo de entorno:

```bash
cp .env.example .env
```

3. Configurá `RABBIT_URL` y `RABBIT_QUEUE` en `.env`.

4. Ejecutá el servicio:

```bash
go run ./cmd
```

El servicio va a conectarse a RabbitMQ, empezar a consumir mensajes y escribir logs en formato JSON a stdout.

## Notas

- Si ejecutás el servicio desde otra carpeta, la carga de `.env` intenta resolver tanto `.env` como `../../.env`.
- Si `RABBIT_QUEUE` está vacío, el servicio falla al iniciar por validación de configuración.