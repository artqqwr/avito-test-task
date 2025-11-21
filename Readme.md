# Avito Test Task 

[Ссылка на тз](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/Backend/Backend-trainee-assignment-autumn-2025/Backend-trainee-assignment-autumn-2025.md)


Запуск:

```bash
docker compose up

# или

make compose-up
```


Стек:
- Go
- OpenAPI (oapi-codegen)
- Postgres (pgx)
- Migrations (goose)
- HTTP router (chi)


Стуркута проекта:
```text
avito-test-task
├── api (openapi spec and oapi-codegen config) 
├── cmd
│    └── server
├── internal
│    ├── app 
│    ├── controller
│    │    └── http
│    ├── domain (models)
│    ├── repository
│    │    ├── postgres
│    └── service (core logic) 
├── migrations
├── pkg
│    └── api (oapi-codegen generated code)
```


Комманды Makefile:
```bash
make run
make build
make generate
make clean
make compose-up
make compose-down
make load-test
```

Результаты нагрузочного тестирование Grafana k6 ([load_test_results.txt](./load_test_results.txt)):
```text
SLI времени ответа = 16.27 ms 
RPS = ~70 rps 
SLI успешности = 100%
```

**Буду благадарен за любой фитбек**

Вопросы:

1. Раньше пытался на круд применять чистую архитектуру (не напороться бы на холивар) - разделение слоёв, у кажого слоя свои модели с которыми он работает.
У контроллера - свои dto, у сервиса - *ModelName*Input/*ModelName*Output, у репозитория - для каждой бд свои модели (при том что у меня используется ток postgres) 
В итоге, сам себя топил бесконечными мапперами.
Поэтому в этому проекте один слой моделей - domain.
2. Текущая реализация сервисного слоя не исключает других, уже назначенных ревьюверов.
Мы можем назначить одного и того же человека дважды, если в команде осталось мало активных пользователей. 
Учитывая требование низкой задержки, оправдана ли такая ошибка (повторное назначение) в обмен на то что нам не нужно делать дополнительный селект в бд для каждого PR, чтобы узнать всех текущих ревьюверов?