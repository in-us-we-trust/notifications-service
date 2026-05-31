# Notification Service

Servicio en Go que consume mensajes desde RabbitMQ.

Hoy el proceso de arranque crea el cliente de RabbitMQ y empieza a consumir mensajes de una cola. El procesamiento actual solamente registra el contenido del mensaje en logs JSON.

## Requisitos

- Go 1.25.0
- RabbitMQ accesible desde la URL configurada

## Configuración

Creá tu archivo de entorno a partir del ejemplo:

```bash
cp .env.example .env
```

Variables requeridas para levantar el servicio:

- `RABBIT_URL`: URL de conexión a RabbitMQ.
- `RABBIT_QUEUE`: nombre de la cola a consumir.

Variables definidas en el código pero no utilizadas en el arranque actual:

- `FCM_CREDENTIALS_PATH`: path al JSON de credenciales de Firebase.
- `USERS_BASE_URL`: URL base del servicio de usuarios.

## Cómo levantarlo

1. Descargá las dependencias:

```bash
go mod download
```

2. Creá el archivo de entorno:

```bash
cp .env.example .env
```

3. Completá al menos estas variables en `.env`:

```dotenv
RABBIT_URL=amqp://guest:guest@localhost:5672/
RABBIT_QUEUE=notifications
```

4. Ejecutá el servicio desde la raíz del proyecto:

```bash
go run ./cmd
```

## Comportamiento esperado

- El servicio abre una conexión con RabbitMQ.
- Empieza a consumir la cola configurada.
- Por cada mensaje recibido, escribe un log JSON con el body.

## Notas

- La carga de entorno intenta resolver `.env` y también `../../.env`, así que puede funcionar aunque el comando se ejecute desde un directorio distinto.
- Si `RABBIT_QUEUE` está vacío, el cliente de RabbitMQ falla al inicializar.
- `RABBIT_PREFETCH_COUNT` aparece en `.env.example`, pero hoy no se lee desde entorno: el valor de prefetch está fijo en `1` en la configuración actual.