# GraphQL Post Comments

Backend-сервис для публикации постов и иерархических комментариев на GraphQL (аналог Habr/Reddit).

## Возможности

- **Посты** — список всех постов, просмотр поста с комментариями
- **Комментарии** — неограниченная вложенность через `parentId`, лимит текста 2000 символов
- **Пагинация** — `comments(limit, offset)` пагинирует корневые комментарии и возвращает все их ответы
- **Модерация** — автор поста может включать/отключать комментирование (`togglePostComments`)
- **Real-time** — GraphQL Subscription `commentAdded` для получения новых комментариев без polling
- **Хранилища** — in-memory или PostgreSQL, переключается через переменную окружения
- **Docker** — готовый `Dockerfile` и `docker-compose.yml`

## Архитектура

```
HTTP/WS → handler (resolvers + dataloader) → service (бизнес-логика) → storage (memory | postgres)
                                              ↓
                                         CommentEventBus (subscriptions)
```

| Слой | Ответственность |
|------|-----------------|
| `handler` | GraphQL resolvers, batched dataloader для N+1 |
| `service` | Бизнес-правила: валидация длины, проверка закрытых комментариев, parentId |
| `storage` | CRUD и пагинация, без бизнес-логики |

## Быстрый старт

### In-memory (без Docker)

```bash
STORAGE_TYPE=memory go run ./cmd/app
```

Playground: http://localhost:8080/

### Docker + PostgreSQL

```bash
docker-compose up --build
```

Сервис: http://localhost:8080/  
PostgreSQL: `localhost:5432`, БД `graphql_db`, пользователь `user` / `password`

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| `STORAGE_TYPE` | `memory` | `memory` или `postgres` |
| `PORT` | `8080` | HTTP-порт |
| `DATABASE_URL` | `postgres://user:password@localhost:5432/graphql_db?sslmode=disable` | DSN для PostgreSQL |

## GraphQL API

### Queries

```graphql
# Все посты (комментарии загружаются batched dataloader'ом)
query {
  posts {
    id
    title
    content
    commentsHidden
    comments {
      id
      content
      parentId
      createdAt
    }
  }
}

# Один пост с пагинацией корневых комментариев
query {
  post(id: "POST_ID") {
    title
    comments(limit: 10, offset: 0) {
      id
      content
      parentId
    }
  }
}
```

### Mutations

```graphql
mutation {
  createPost(title: "Hello", content: "World", commentsHidden: false) {
    id
    title
  }

  createComment(postId: "POST_ID", parentId: "PARENT_COMMENT_ID", content: "Reply") {
    id
    content
  }

  togglePostComments(id: "POST_ID", hidden: true) {
    id
    commentsHidden
  }
}
```

### Subscription

```graphql
subscription {
  commentAdded(postId: "POST_ID") {
    id
    content
    parentId
    createdAt
  }
}
```

> Subscriptions работают через WebSocket. В Playground откройте отдельную вкладку subscription.

## Пагинация комментариев

Пагинация применяется к **корневым** комментариям (`parentId = null`), отсортированным по дате (новые первые).  
Для каждого корневого комментария на странице возвращаются **все вложенные ответы** рекурсивно.

Пример: 3 корневых комментария, `limit: 1, offset: 0` → вернётся 1 корневой + все его потомки.

## Тесты

```bash
go test ./...
```

- `internal/storage` — CRUD и пагинация хранилища
- `internal/service` — бизнес-валидация (table-driven tests) и pub/sub

## Структура проекта

```
cmd/app/           — точка входа
graph/             — GraphQL schema и generated code (gqlgen)
internal/
  config/          — конфигурация из env
  handler/         — resolvers, dataloader middleware
  model/           — доменные типы и ошибки
  service/         — бизнес-логика, event bus
  storage/         — memory и postgres реализации
migrations/        — SQL-схема для PostgreSQL
```

