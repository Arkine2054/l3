# Event Booking System

Создания событий и бронирования мест.
Backend — Go, база данных — PostgreSQL, миграции — migrate/migrate, UI — простые HTML-страницы.

## Функционал

### Создание события

* название
* дата
* количество мест

### Просмотр события

Показывает количество свободных мест и активные брони.

### Бронирование места

* временная бронь на `BOOKING_EXPIRATION_MINUTES`
* если не подтвердить — бронь автоматически очищается

### Подтверждение брони

Финальное закрепление места.

### Автоматический очиститель

Удаляет просроченные брони.

# Структура проекта

```
.
├── cmd/server        # main.go и запуск сервера
├── internal/
│   ├── repository    # SQL работа с БД
│   ├── service       # бизнес-логика
│   └── handlers      # HTTP-ручки
├── migrations        # SQL миграции
├── web/              # HTML UI
├── Dockerfile
├── docker-compose.yml
└── README.md
```

---

# API Endpoints

## Создать событие

`POST /events`

Request:

```json
{
  "title": "Rock Concert",
  "date": "2025-12-01T20:00:00Z",
  "total_seats": 100
}
```

Response:

```json
{
  "id": 1,
  "title": "Rock Concert",
  "total_seats": 100,
  "free_seats": 100
}
```

---

## Получить событие

`GET /events/{id}`

---

## Забронировать место

`POST /events/{id}/book`

Request:

```json
{ "user_name": "Alice" }
```

---

## Подтвердить бронь

`POST /events/{event_id}/confirm`

Request:

```json
{ "booking_id": 5 }
```

---

# Web UI

Открыть в браузере:

**(http://localhost:8080/web)**

Доступные страницы:

| Страница            | URL                      |
| ------------------- | ------------------------ |
| Список действий     | `/web/index.html`        |
| Создать событие     | `/web/create.html`       |
| Открыть событие     | `/web/event.html?id=1`   |
| Забронировать место | `/web/book.html?id=1`    |
| Подтвердить бронь   | `/web/confirm.html?id=1` |

# Тестирование

## Тесты API (ручное)

```bash
curl -X POST http://localhost:8080/events -d ...
```

## Тесты UI

1. Открыть (http://localhost:8080/web)
2. Создать событие
3. Забронировать
4. Подтвердить
