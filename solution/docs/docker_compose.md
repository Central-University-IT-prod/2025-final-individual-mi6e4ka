# структура docker compose

В docker compose прописаны 3 необходимых и 2 дополнительных сервиса для работы api рекламного сервиса

```yaml
db:
  image: postgres:17.3-alpine
  restart: always
  volumes:
    - postgres_data:/var/lib/postgresql/data
  environment:
    POSTGRES_DB: prod
    POSTGRES_USER: prod
    POSTGRES_PASSWORD: brain_rot
  healthcheck:
    test: ["CMD-SHELL", "pg_isready", "--quiet"]
    interval: 2s
    timeout: 5s
    retries: 10
```

Тут я использую бд Postgresql конкретной версии, чтобы не получить ошибок совместимости в проде  
Далее подключаю volume для сохранение данных при перезагрузках и прописываю в environment параметры инициализации бд  
Также прописан healthcheck для ожидания работоспособности бд у api

```yaml
grafana:
  image: grafana/grafana-oss:11.3.4
  container_name: grafana
  restart: unless-stopped
  # admin:proood
  user: "0"
  ports:
    - "3000:3000"
  environment:
    - GF_AUTH_ANONYMOUS_ENABLED=true
    - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    - GF_SECURITY_ALLOW_EMBEDDING=true
    - GF_AUTH_DISABLE_LOGIN_FORM=true
    - GF_PATHS_PROVISIONING=/etc/grafana/provisioning
  volumes:
    - ./grafana:/etc/grafana/provisioning
```

Также загружаю фиксированную версию grafana, которая будет слушать на 3000 порту и загружать dashboard доступный без авторизации

```yaml
minio:
  image: minio/minio:RELEASE.2025-02-18T16-25-55Z
  container_name: minio
  entrypoint: sh
  command: -c 'mkdir -p /data/prod && /usr/bin/minio server /data --console-address ":9090"'
  environment:
    MINIO_ROOT_USER: prod
    MINIO_ROOT_PASSWORD: not_for_prod
    MINIO_DEFAULT_BUCKETS: prod
    MINIO_BROWSER_REDIRECT_URL: "http://localhost:9090"
    MINIO_API_REQUESTS_MAX: 0
    MINIO_DOMAIN: localhost
  volumes:
    - minio_data:/data
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
    interval: 2s
    timeout: 5s
    retries: 10
```

minio, при запуске которого автоматически создается bucket prod

```yaml
ollama:
  image: mi6e4kadev/ollama-lightweight:0.5.11-qwen2.5-0.5b
  # build: ./llm
  container_name: ollama
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:11434"]
    interval: 2s
    timeout: 5s
    retries: 10
```

Собственный билд легковесной ollama со встроенной моделью qwen2.5-0.5b. Dockerfile по которому был собран образ лежит в папке solution/llm

```yaml
api:
  build: ./api
  ports:
    - 8080:8080
  container_name: prod_api
  volumes:
    - ./api/configs:/app/configs
  depends_on:
    db:
      condition: service_healthy
    minio:
      condition: service_healthy
    ollama:
      condition: service_healthy
  healthcheck:
    test: ["CMD", "curl", "-f", "http://localhost:8080"]
    interval: 2s
    timeout: 5s
    retries: 10
  restart: always
```

И наконец сам api. Собирается по dockerfil'у из папки api. Подгружает конфигурационный файл из папки api/configs, и ожидает когда поднимутся все контейнеры необходимые для его работы. Есть healthcheck который проверяет работоспособность по корневому эндпоинту api

```yaml
bot:
  build: ./bot
  environment:
    TOKEN: REDACTED
    API_URL: http://api:8080
```

Telegram бот. Через переменные среды задается токен telegram и api_url (адрес самого api)

```yaml
volumes:
  postgres_data:
  minio_data:
```

Инициализация хранилищ данных для сервисов которым необходимо сохранять данные между перезапусками
