Тестирование приложения через Postman

Ваше приложение — это REST API на Go с PostgreSQL, предоставляющее эндпоинты для управления студентами (аутентификация) и дипломами (CRUD). Оно использует JWT-токены для защиты эндпоинтов дипломов. Для тестирования используйте Postman (бесплатный инструмент для API-тестов). Я опишу полный процесс: от запуска до примеров запросов.
Предварительные шаги

  Запустите приложение: Убедитесь, что Docker Compose запущен (docker-compose up --build). Приложение доступно на http://localhost:8888.
  Установите Postman: Скачайте с https://www.postman.com/downloads/ и создайте новый workspace.
  Создайте коллекцию: В Postman нажмите "New" > "Collection", назовите "Gosmol API". Добавляйте запросы в неё.
  Базовый URL: http://localhost:8888. Все запросы к этому адресу.
  Формат данных: Все тела запросов — JSON (Content-Type: application/json). Ответы тоже JSON.
  JWT-токены: Для защищённых эндпоинтов добавляйте заголовок Authorization: Bearer <access_token> (получите его через логин).

Эндпоинты и тестирование

Вот все эндпоинты из кода. Тестируйте по порядку: сначала аутентификация, затем дипломы. Используйте переменные Postman для хранения токенов (например, {{access_token}}).
1. Регистрация студента (не требует JWT)

    Метод: POST
    URL: {{base_url}}/api/auth/register
    Тело запроса (raw JSON):

{

  "firstname": "John",

  "lastname": "Doe",

  "email": "john@example.com",

  "password": "Password123!"

    }

    Ожидаемый ответ: 200 OK, JSON с данными студента (id, firstname и т.д.).
    Тест:
        Отправьте запрос.
        Если 400 — проверьте пароль (минимум 8 символов, буквы, цифры, спецсимволы) или email (уникальный).
        Повторите с другим email для второго студента.

2. Логин студента (не требует JWT)

    Метод: POST
    URL: {{base_url}}/api/auth/login
    Тело запроса (raw JSON):

{

  "email": "john@example.com",

  "password": "Password123!"

    }

    Ожидаемый ответ: 200 OK, JSON с access_token и refresh_token.
    Тест:
        Отправьте запрос.
        Сохраните access_token в переменную Postman (Tests tab: pm.collectionVariables.set("access_token", pm.response.json().access_token);).
        Если 401 — неверные данные. После 5 неудачных попыток аккаунт блокируется на 1 минуту (повторите позже).
        Если 400 — email пустой.

3. Обновление токена (не требует JWT)

    Метод: POST
    URL: {{base_url}}/api/auth/refresh
    Тело запроса (raw JSON):

{

  "refresh_token": "{{refresh_token}}"

    }

    Ожидаемый ответ: 200 OK, новые access_token и refresh_token.
    Тест:
        Сохраните новый access_token.
        Если 401 — токен истёк или недействительный.

4. Получить все дипломы (требует JWT)

    Метод: GET
    URL: {{base_url}}/api/resources
    Заголовки: Authorization: Bearer {{access_token}}
    Ожидаемый ответ: 200 OK, массив дипломов (пустой, если нет данных).
    Тест:
        Отправьте запрос.
        Если 401 — токен недействительный (обновите через refresh).

5. Создать диплом (требует JWT)

    Метод: POST
    URL: {{base_url}}/api/resources
    Заголовки: Authorization: Bearer {{access_token}}
    Тело запроса (raw JSON):

{

  "title": "My Diploma",

  "description": "Description here, up to 500 chars"

    }

    Ожидаемый ответ: 200 OK, JSON с созданным дипломом (id, title, description).
    Тест:
        Отправьте запрос.
        Сохраните id (например, pm.collectionVariables.set("diploma_id", pm.response.json().id);).
        Если 400 — title пустой или description >500 символов.

6. Получить диплом по ID (требует JWT)

    Метод: GET
    URL: {{base_url}}/api/resource/{{diploma_id}}
    Заголовки: Authorization: Bearer {{access_token}}
    Ожидаемый ответ: 200 OK, JSON с дипломом.
    Тест:
        Замените {{diploma_id}} на реальный ID.
        Если 400 — ID не найден.

7. Обновить диплом (требует JWT)

    Метод: PUT
    URL: {{base_url}}/api/resource/{{diploma_id}}
    Заголовки: Authorization: Bearer {{access_token}}
    Тело запроса (raw JSON):

{

  "title": "Updated Diploma",

  "description": "Updated description"

    }

    Ожидаемый ответ: 200 OK.
    Тест:
        Проверьте GET после обновления.

8. Удалить диплом (требует JWT)

    Метод: DELETE
    URL: {{base_url}}/api/resource/{{diploma_id}}
    Заголовки: Authorization: Bearer {{access_token}}
    Ожидаемый ответ: 200 OK.
    Тест:
        Проверьте GET — должен вернуть 400 (не найден).

Советы по тестированию

    Переменные в Postman: Используйте {{base_url}} = http://localhost:8888, {{access_token}}, {{refresh_token}}, {{diploma_id}}. Сохраняйте их в Tests tab запросов.
    Ошибки: Если 401 — проверьте токен. Если 400 — валидация данных. Логи приложения: docker-compose logs app.
    CORS: Если браузер блокирует, добавьте заголовок Origin: http://localhost:8888 в запросы (хотя в коде CORS разрешён).
    Полный сценарий: Регистрация → Логин → Создание диплома → Чтение/Обновление/Удаление → Refresh токена.
    Дополнительно: Тестируйте edge-кейсы, как пустые поля, длинные описания или блокировку аккаунта.

Если что-то не работает, поделитесь скриншотами Postman или логами!