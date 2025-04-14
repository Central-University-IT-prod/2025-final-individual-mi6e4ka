# инструкция по запуску

## предварительная подготовка

Для запуска необходимы:

- docker
- docker compose

## шаги запуска

### 1. создать файл config.yaml по пути ./api/configs/config.yaml со следующим содержимым, подставив свои данные

```yaml
db:
  host: db
  user: prod
  password: brain_rot
  db_name: prod
  port: 5432
http:
  port: 8080
s3:
  endpoint: minio:9000
  user: prod
  password: not_for_prod
ollama:
  base_url: http://ollama:11434
  model: qwen2.5:0.5b
```

### 2. заменить переменные окружения для telegram бота в docker-compose и поднять docker-compose

```bash
docker-compose up -d
```

### 3. дождаться запуска всех сервисов, api будет доступен на порту 8080
