
# Sales Tracker — CRUD сервис для учёта продаж + аналитика

Cервис для хранения информации о продажах, получения агрегированной аналитики и построения отчетов.

Проект использует:
- Go  
- Chi Router  
- PostgreSQL  
- SQL миграции  
- Docker + docker-compose  

---

## Структура проекта

```

.
│   docker-compose.yml
│   Dockerfile
│   go.mod
│   go.sum
│   README.md
│
├───cmd
│   └───server
│           main.go          # запуск HTTP API
│
├───internal
│   ├───models
│   │       models.go        # структуры данных
│   │
│   └───repository
│           repository.go    # SQL-операции с продажами
│
├───migrations
│       001_create_sales_table.up.sql   # миграция таблицы sales

````

---

## Функциональность

### CRUD операции по продажам
- Создать продажу  
- Получить список всех продаж  
- Получить конкретную продажу  
- Обновить  
- Удалить  

### Аналитика (SQL агрегирование)
- общая выручка  
- количество продаж  
- средняя сумма покупки  
- продажи по дням/месяцам  
- топовые товары  

### Поддержка миграций
Миграции автоматически применяются при старте сервера.

---

## Технологии

- **Go 1.22+**
- **PostgreSQL 15**
- **Chi Router**
- **Docker**
- **SQL миграции**

---

## 🔌 API Endpoints

### Создать продажу

```
POST /sales
```

Body:

```json
{
  "product": "Laptop",
  "price": 1900.00,
  "quantity": 2
}
```

---

### Получить все продажи

```
GET /sales
```

---

### Получить одну продажу

```
GET /sales/{id}
```

---

### Аналитика

```
GET /analytics/summary
```

Пример ответа:

```json
{
  "total_revenue": 54000,
  "total_sales": 120,
  "average_check": 450,
  "daily": [
    { "date": "2024-01-10", "revenue": 1200 }
  ]
}
```