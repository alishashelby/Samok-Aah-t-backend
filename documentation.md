# Финальный проект - дока

## Запуск локально

```
docker-compose up --build
```

Далее приложение будет доступен на http://localhost:8080.

## Coverage юнит-тестов
```
go test ./internal/domain/service/service/... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

## Curl-запросы
В openapi-public.yaml - курлы можно сгенерить для авторизации; в openapi-authorized.yml - курлы уже для всего, что с авторизацией соответственно.

## Описание эндпоинтов
Постаралась нормально сделать в самой спецификации

* Замечание: Сначала идет регистрация, потом уже создается сама роль (клиент или модель или админ) и доп данные прилагаются.

## Первый запуск
*.env специально вытащила из gitignore для удобной проверки

Как создать самого супер админа в системе (+ представим что их права(permissions) учитываются - на самом деле бутафория, которые вводятся при создании админа):

```bash
docker exec -it service_postgres psql -U postgres -d postgres_db
```

```postgresql
 INSERT INTO auth (email, password_hash, role) VALUES ('admin1234@mail.ru', '$2a$10$4DVZlP2kyoV1lytLkUzsmeqXTDlP.1NIWk4GxfHKXh5B6YCWfZDtS', 'ADMIN') RETURNING auth_id;
```
admin1234 - пароль для входа

Смотрим под каким auth_id:
```postgresql
 SELECT auth_id, email, role FROM auth;
```

Сюда этот auth_id
```postgresql
INSERT INTO admins (auth_id, permissions) VALUES (1, '{"test permission": true}');
```

Проверить:
```postgresql
SELECT a.auth_id, a.email, ad.permissions FROM auth a JOIN admins ad ON ad.auth_id = a.auth_id;
```

## Логика фоновых задач
Разрешено было чисто ограничиться метриками, кстати доступно по http://localhost:8080/metrics
Но я добавила фонового воркера для обновления метрик общего числа моделей и клиентов в системе.

Также касательно "booking" - задумывалось также добавить воркер для фоновой проверки expiration и перевод брони с зарезервированного 
на свободный статус - добью в будущем, так как проект этот еще имеет нормально спроектированную бд отказоустойчивую, так что
он со мной надолго.
