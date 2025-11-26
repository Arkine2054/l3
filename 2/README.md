# URL Shortener Service

Cервис сокращения ссылок на Go с поддержкой:

* кастомных alias’ов
* редиректов
* сбора аналитики (IP, User-Agent, Referer, время перехода)
* Web UI
* Docker/Compose для быстрого запуска

## Стек

* Go 1.22+
* Gorilla Mux
* PostgreSQL
* Docker + docker-compose
* HTML/JS UI

---

## API

### Создать короткую ссылку

`POST /shorten`

#### Request JSON

```json
{
  "target": "https://golang.org",
  "alias": "go"
}
```

#### Response

```json
{
  "alias": "go",
  "url": "http://localhost:8085/s/go"
}
```

---

### Редирект

`GET /s/{alias}`

Пример:
`GET http://localhost:8085/s/go`

→ HTTP 302 Redirect на `https://golang.org`

---

### Получить аналитику

`GET /analytics/{alias}`

#### Response

```json
[
  {
    "id": 1,
    "alias": "go",
    "user_agent": "Mozilla/5.0...",
    "ip": "172.18.0.1",
    "referer": "",
    "created_at": "2025-02-14T15:12:03Z"
  }
]
```

## Web UI

Функции UI:

* создание коротких ссылок
* копирование результата в буфер обмена
* просмотр аналитики
* автоподстановка alias в аналитике

---

## Структура проекта

```
.
├── cmd/
│   └── app/main.go
├── internal/
│   ├── handlers/
│   ├── services/
│   ├── repository/
│   └── router/
├── ui/
│   └── index.html
├── docker-compose.yml
├── Dockerfile
└── README.md
```
