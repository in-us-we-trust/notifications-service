# Notification Service

Servicio en Go que consume notificaciones desde RabbitMQ y las envía como push al dueño del dispositivo. Si el push no se puede entregar, manda el mismo aviso por email.

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
- `FCM_CREDENTIALS_PATH`: path al JSON de credenciales de Firebase.
- `USERS_BASE_URL`: URL base del servicio de usuarios.
- `MAILER_FROM_EMAIL`, `MAILER_FROM_PSW`, `MAILER_FRONTEND_URL`, `MAILER_DEV_EMAIL` y `ENV`: email de respaldo cuando no hay push.

Al arrancar, el servicio declara el exchange topic durable `notifications`, las colas `notification.intruder` y `notification.unknown-person`, y el binding de cada cola con esa misma routing key.

## Cómo levantarlo

1. Descargá las dependencias:

```bash
go mod download
```

2. Creá el archivo de entorno:

```bash
cp .env.example .env
```

3. Completá las variables requeridas en `.env`. RabbitMQ:

```dotenv
RABBIT_URL=amqp://guest:guest@localhost:5672/
```

4. Ejecutá el servicio desde la raíz del proyecto:

```bash
go run ./cmd
```

## Comportamiento esperado

- El servicio abre una conexión con RabbitMQ y declara el exchange, las dos colas y los bindings
- Consume `notification.intruder` (alerta de peligro) y `notification.unknown-person` (sugerencia para confirmar una persona desconocida).
- El push sale solo a los usuarios owner del `deviceId`.
- Intruso: el push usa `title`, `message` e `imageUrl` como vienen. Al tocarlo, la app abre esa alerta con esa foto. Si no hay token, o el push falla, se manda por email.
- Persona desconocida: el texto se arma con `distinctDays`. El push muestra la primera foto de `imageUrls` y, al tocarlo, la app abre la confirmación de esa `personId` con todas las fotos. Si `imageUrls` viene vacío, el push se envía igual. Esta sugerencia sale solo por push: si el owner no tiene token, no se manda.

## Notas

- La carga de entorno intenta resolver `.env` y también `../../.env`, así que puede funcionar aunque el comando se ejecute desde un directorio distinto.
- `RABBIT_PREFETCH_COUNT` aparece en `.env.example`, pero hoy no se lee desde entorno: el valor de prefetch está fijo en `1` en la configuración actual.
- `RABBIT_QUEUE` ya no se usa. La cola anterior del exchange por defecto deja de consumirse.