```markdown
# DelayedNotifier — Отложенные уведомления через очереди

DelayedNotifier — это сервис на Go для создания и отправки отложенных уведомлений через RabbitMQ и Postgres.  

Он поддерживает:
- Планирование уведомлений на конкретное время
- Проверку статуса уведомлений
- Разделение на API-сервис и воркер для обработки очередей.

---

## Структура проекта

```

.
├── cmd
│   ├── api        # HTTP-сервис для создания/просмотра/отмены уведомлений
│   └── worker     # Воркер, который обрабатывает очередь RabbitMQ
├── internal
│   ├── model      # Структуры данных, модели уведомлений
│   ├── queue      # Подключение и работа с RabbitMQ
│   ├── repo       # Работа с базой данных
│   └── service    # Основная логика отправки уведомлений
├── docker-compose.yaml
├── go.mod
└── go.sum

````

---

## Как работает сервис

1. **API-сервис** принимает HTTP-запросы:
   - `POST /notify` — создать уведомление с датой и временем отправки
   - `GET /notify/{id}` — получить статус уведомления
   - `DELETE /notify/{id}` — отменить запланированное уведомление  

2. **Worker** подписан на очередь RabbitMQ:
   - Получает уведомления
   - Ждёт, если время отправки ещё не наступило
   - Отправляет уведомление
   - При ошибках публикует повторно в очередь с экспоненциальным backoff

3. **PostgreSQL** хранит уведомления и их статусы

4. **RabbitMQ** обеспечивает асинхронную очередь уведомлений  

---

## Запуск через Docker

### Собрать и запустить сервис

```bash
docker-compose up --build
````

* API доступен на `http://localhost:8082`
* RabbitMQ Management UI: `http://localhost:15673` (логин/пароль: guest/guest)
* Postgres на порту `5433`

---

## HTTP API

### Создание уведомления

```http
POST /notify
Content-Type: application/json

{
  "recipient": "user@example.com",
  "channel": "email",
  "message": "Ваш заказ готов",
  "send_at": "2025-10-07T22:00:00Z"
}
```

Ответ:

```json
{
  "id": 1,
  "status": "scheduled"
}
```

---

### Получение статуса уведомления

```http
GET /notify/1
```

Ответ:

```json
{
  "id": 1,
  "status": "sent",
  "attempts": 1
}
```

---

### Отмена уведомления

```http
DELETE /notify/1
```

Ответ:

```json
{
  "id": 1,
  "status": "cancelled"
}
```

---

## Тестирование локально

### Через curl

**Создание уведомления:**

```bash
curl -X POST http://localhost:8082/notify \
  -H "Content-Type: application/json" \
  -d '{
        "recipient": "user@example.com",
        "channel": "email",
        "message": "Ваш заказ готов",
        "send_at": "2025-10-07T22:00:00Z"
      }'
```

**Проверка статуса уведомления:**

```bash
curl http://localhost:8082/notify/1
```

**Отмена уведомления:**

```bash
curl -X DELETE http://localhost:8082/notify/1
```

### Через Postman

1. Создать коллекцию `DelayedNotifier`
2. Добавить запросы:

    * `POST /notify` с JSON телом
    * `GET /notify/{id}`
    * `DELETE /notify/{id}`

---

## Настройки через окружение

| Переменная     | Описание                       | Пример                                                        |
| -------------- | ------------------------------ | ------------------------------------------------------------- |
| `DATABASE_URL` | URL для подключения к Postgres | `postgres://user:pass@postgres:5432/notifier?sslmode=disable` |
| `RABBITMQ_URL` | URL для подключения к RabbitMQ | `amqp://guest:guest@rabbitmq:5672/`                           |

---

## Важные моменты

* Worker и API запускаются отдельно, Worker обрабатывает очередь и retry.

```
