# Backend Service

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Version](https://img.shields.io/badge/version-1.0.0-green.svg)](https://github.com/username/project)
[![Build Status](https://img.shields.io/travis/username/project/master.svg)](https://travis-ci.org/username/project)

Масштабируемый бэкенд-сервис на Go с REST API, поддержкой PostgreSQL, Docker-оркестрацией и SSL-терминацией через nginx.

## Оглавление
- [Описание](#описание)
- [Установка](#установка)
- [Использование](#использование)
- [API](#api)
- [Тестирование](#тестирование)
- [Вклад в проект](#вклад-в-проект)
- [Лицензия](#лицензия)
- [Контакты](#контакты)

## Описание

Проект представляет собой модульный бэкенд-сервис, разработанный с использованием стандартной структуры Go-проектов (Standard Go Project Layout). Он решает задачу создания надёжной основы для веб-приложений с поддержкой контейнеризации, базы данных и безопасного соединения.

Почему он лучше аналогов:
- Чистая архитектура с разделением слоёв (http, logic, model, db)
- Готовая инфраструктура для локальной разработки через Docker Compose
- Встроенная поддержка health checks и телеметрии
- Конфигурация nginx с SSL-терминацией «из коробки»

### Основные возможности
- REST API v1 с модульной структурой (api/http/v1)
- Оркестрация сервисов через Docker Compose (backend, nginx, postgres)
- Автоматическая настройка PostgreSQL с параметрами логирования и локализации
- Обратный прокси с SSL-терминацией на порт 443
- Система health checks для мониторинга состояния сервиса
- Поддержка миграций базы данных (internal/db/migrations)
- Модуль телеметрии для сбора метрик и логирования

## Установка

### Требования
- Go 1.20+
- Docker и Docker Compose
- PostgreSQL 14+ (при локальном запуске без Docker)
- Make (опционально)

### Пошаговая установка

# Клонируйте репозиторий
git clone https://github.com/username/project.git

# Перейдите в директорию проекта
cd project

# Запустите проект через Docker Compose (рекомендуемый способ)
docker-compose up -d

# Или соберите и запустите локально
cd cmd/backend
go build -o backend .
./backend

# Настройте переменные окружения при необходимости
cp .env.example .env
# Отредактируйте .env файл под свои параметры

# Запустите миграции базы данных (если не используются автоматические)
# Пример для Go-миграций:
go run cmd/backend/main.go migrate up

## Использование

После запуска сервис доступен по адресу:
- HTTPS: https://localhost:443 (через nginx)
- HTTP: http://localhost:8080 (напрямую, если не использовать nginx)

Пример запроса к API:
curl -X GET https://localhost/api/v1/health --insecure

## API

Базовый эндпоинт: /api/v1

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| GET | /health | Проверка состояния сервиса |
| GET | /api/v1/* | Основные эндпоинты API v1 |

Подробная документация API доступна по пути: api/http/v1

## Тестирование

# Запустить юнит-тесты
go test ./...

# Запустить тесты с покрытием
go test ./... -coverprofile=coverage.out

# Запустить интеграционные тесты (требуется Docker)
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

## Вклад в проект

1. Создайте форк репозитория
2. Создайте ветку для вашей фичи: git checkout -b feature/amazing-feature
3. Внесите изменения и закоммитьте: git commit -m 'Add amazing feature'
4. Отправьте изменения: git push origin feature/amazing-feature
5. Откройте Pull Request

Пожалуйста, соблюдайте стандарты кодирования Go и добавляйте тесты для нового функционала.

## Лицензия

Распространяется под лицензией MIT. Подробности в файле [LICENSE](LICENSE).

## Контакты

Ваше Имя — [@username](https://github.com/username) — email@example.com

Ссылка на проект: [https://github.com/username/project](https://github.com/username/project)