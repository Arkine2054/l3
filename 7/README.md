# Warehouse App

Современное мини-приложение для управления складскими товарами.
Поддерживает роли пользователей, историю изменений, JWT-аутентификацию и удобный web-интерфейс.

### Авторизация

* JWT-аутентификация
* Роли пользователей:

    * **admin** — полный доступ
    * **manager** — создавать/редактировать
    * **viewer** — только просмотр

### Управление товарами

* Добавление / редактирование / удаление
* Поиск и сортировка
* Постраничный вывод

### 🌐 Web-интерфейс (Bootstrap + JS)

* Темная тема
* Модальное окно редактирования
* Поиск / сортировка / пагинация
* Просмотр истории товара

---

## Структура проекта

```
.
├── cmd/main.go
├── internal/
│   ├── handlers/
│   ├── models/
│   ├── repository/
│   ├── middleware/
│   ├── server/
│   └── utils/
├── migrate/               # SQL миграции
├── web/                   # Фронтенд
└── docker-compose.yml
```
### Запуск через Docker

```sh
docker-compose up --build
```

После запуска:

* Backend → [http://localhost:8080]
* Frontend → [http://localhost:8091]
* PostgreSQL → `localhost:5432`

## 🔥 API

### Auth

```
POST /login
POST /register
```

### Items

```
GET    /items
POST   /items
PUT    /items/{id}
DELETE /items/{id}
```

### History

```
GET /history/{item_id}
```
Функции UI:

* авторизация
* фильтры
* сортировка
* CRUD
* просмотр diff
* экспорт CSV
