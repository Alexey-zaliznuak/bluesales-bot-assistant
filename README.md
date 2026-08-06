# BlueSales Bot Assistant

SPA-ассистент по настройке и созданию ботов в BlueSales. Документы базы знаний объединяются
в системный контекст, который отправляется модели вместе с сообщениями чата.

Стек: Vite + React 19 + TypeScript, Go 1.25 + chi, PostgreSQL 17, docker compose.
Модель по умолчанию — `openai/gpt-5.6-luna` с `reasoning.effort = max`.

## Быстрый старт

```bash
cp .env.example .env          # Windows: Copy-Item .env.example .env
# впишите OPENROUTER_API_KEY и смените пароли
docker compose up -d --build
```

- Интерфейс: http://localhost:5173
- API: http://localhost:8088/api/health (порт задаётся `API_EXPOSED_PORT`)
- Логин и пароль — из `SEED_USER_LOGIN` / `SEED_USER_PASSWORD`

Порядок запуска сервисов: `db` → `migrate` → `seed` → `api` и `web`. Миграции и сид выполняются
одноразовыми сервисами, оба идемпотентны. Все переменные окружения сервисы получают через
`env_file: .env` — inline-переменных в compose нет.

## Как устроены снимки базы знаний

Кнопка **Синхронизировать базу знаний** делает следующее:

1. Собирает все документы в один текст стабильного формата (XML-разметка, порядок по `created_at, id`).
2. Считает SHA-256 от результата, сохраняет снимок в `kb_snapshots` и делает его активным.

Формат склейки:

```
<system_instructions>Ты — ассистент по настройке и созданию ботов в сервисе BlueSales...</system_instructions>

<knowledge_base format="xml/v1" documents="12">
<document id="..." title="Автоворонка приветствия" categories="воронки, старт">
...тело документа...
</document>
</knowledge_base>
```

Запрос к модели выглядит так:

```json
{
  "model": "openai/gpt-5.6-luna",
  "reasoning": { "effort": "max" },
  "messages": [
    { "role": "system", "content": "<контекст базы знаний>" },
    { "role": "user", "content": "..." }
  ]
}
```

Каждый чат фиксирует за собой снимок базы знаний в момент создания: пересборка базы позже
не меняет контекст уже начатых диалогов.

## Прокси

`OPENROUTER_PROXY_URL` поддерживает схемы `http`, `https`, `socks5`. Если переменная пустая,
бэкенд ходит напрямую и игнорирует системные `HTTP_PROXY`. Проверить связность:

```bash
curl -b cookies.txt http://localhost:8080/api/health/openrouter
```

## Разделы интерфейса

**База знаний** — список документов с фильтром по категориям и поиском, редактор
(заголовок, категории тегами, тело), панель состояния синхронизации с просмотром готового
префикса и кнопкой синхронизации.

**Чаты** — список чатов, обмен сообщениями со стримингом токенов, прикрепление текстовых файлов,
раскрывающийся блок рассуждений модели и статистика токенов под каждым ответом.

Вложения принимаются только текстовые (`ALLOWED_UPLOAD_EXTENSIONS`), содержимое проверяется на
UTF-8 и вставляется в сообщение блоками `<attachment filename="...">`.

## API

| Метод | Путь | Назначение |
| --- | --- | --- |
| POST | `/api/auth/login` | вход, ставит httpOnly cookie с токеном сессии |
| POST | `/api/auth/logout` | выход |
| GET | `/api/auth/me` | текущий пользователь |
| GET | `/api/health` | состояние сервиса и БД |
| GET | `/api/health/openrouter` | проверка ключа и связности через прокси |
| GET/POST | `/api/documents` | список и создание документов |
| GET | `/api/documents/categories` | все категории |
| GET/PATCH/DELETE | `/api/documents/{id}` | чтение, правка, удаление |
| GET | `/api/kb/status` | активный снимок и признак устаревания |
| GET | `/api/kb/preview` | готовый префикс промпта |
| POST | `/api/kb/sync` | пересборка активного снимка |
| GET/POST | `/api/chats` | список и создание чатов |
| GET/PATCH/DELETE | `/api/chats/{id}` | чат с сообщениями, переименование, удаление |
| POST | `/api/chats/{id}/messages` | отправка сообщения, ответ SSE-потоком |

События SSE: `user_message`, `chat`, `reasoning`, `delta`, `usage`, `error`, `done`.

## Разработка без docker

```bash
# БД
docker compose up -d db

# бэкенд (переменные из .env, DATABASE_URL с localhost и POSTGRES_EXPOSED_PORT)
cd backend && go run ./cmd/migrate && go run ./cmd/seed && go run ./cmd/api

# фронтенд
cd frontend && npm install && npm run dev
```

## Тесты

```bash
cd backend && go test ./...
cd frontend && npm run typecheck && npm run build
```

Тесты покрывают детерминированность склейки базы знаний, экранирование разметки в теле
документа и разбор SSE-потока.

## Продакшн-сборка фронтенда

В `frontend/Dockerfile` есть стадия `prod`: собирает статику и отдаёт её nginx, который
проксирует `/api` на сервис `api` с отключённой буферизацией (иначе стриминг ответов встанет).
Для использования замените в compose `target: dev` на `target: prod` и порт на `80`.

## Безопасность

- Пароли хранятся в bcrypt, в БД лежит только SHA-256 от токена сессии.
- Cookie — httpOnly, SameSite=Lax; при работе по HTTPS выставьте `SESSION_COOKIE_SECURE=true`.
- Истёкшие сессии подчищаются фоновой задачей раз в час.
- `.env` в репозиторий не попадает, шаблон — `.env.example`.
