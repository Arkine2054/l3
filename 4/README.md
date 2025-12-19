# Image Processing Service

Микросервисная система для загрузки изображений, их обработки (ресайз + ватермарк), генерации превью и хранения метаданных.
Включает:

* **API-сервер** (загрузка, просмотр, выдача файлов)
* **Image Worker** (асинхронная обработка через Kafka)
* **PostgreSQL** (хранение данных)
* **Kafka** (очередь задач)
* **Docker Compose** для локального запуска

---

## Возможности

*  Загрузка изображений (`/upload`)
*  Асинхронная обработка:

    * ресайз самого изображения
    * добавление watermark-текста
    * создание миниатюры (thumbnail)
*  Хранение оригинала, обработанной копии и превью
*  Просмотр списка файлов (`/images`)
*  Удаление изображения
*  Отдача файлов по запросу (`/image/{id}?type=thumb|processed|original`)
*  Автообновление UI каждые 3 секунды
*  Полноценный фронтенд с HTML/JS

---

## Стек технологий

* Go 1.22
* Kafka (segmentio/kafka-go)
* PostgreSQL
* imaging (обработка изображений)
* freetype (рендеринг текста на изображении)
* Docker Compose

---

## Структура проекта

```
/cmd
  /api           — API-сервер
  /image-worker  — обработчик изображений
/internal
  /repository    — работа с БД
  /processor     — watermark + resize
  /assets        — шрифт font.ttf
/web             — HTML/JS интерфейс
docker-compose.yml
```

---

## Интерфейс

После запуска открой:
 
**(http://localhost:8087)**

##API

### Загрузка изображения

```
POST /upload
Content-Type: multipart/form-data
```

### Список изображений

```
GET /images
```

Пример ответа:

```json
[
  {
    "id": 3,
    "orig_filename": "cat.jpg",
    "stored_path": "/data/original/cat.jpg",
    "processed_path": "/data/processed/cat.jpg",
    "thumb_path": "/data/thumbs/cat.jpg",
    "status": "done",
    "format": "jpg"
  }
]
```

### Получение файла

```
GET /image/{id}?type=thumb
GET /image/{id}?type=processed
GET /image/{id}?type=original
```

### Удаление

```
DELETE /image/{id}
```