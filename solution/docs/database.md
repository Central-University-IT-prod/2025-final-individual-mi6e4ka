# структура базы данных

### таблица Clients

информация о клиентах

- `client_id` uuid - uuid клиента
- `login` string - login клиента. уникальный
- `age` uint - возраст клиента, >= 0
- `location` string - локация клиента
- `gender` string - гендер клиента. принимает только MALE или FEMALE

### таблица Advertisers

информация о рекламодателях

- `advertiser_id` uuid - uuid рекламодателя
- `name` string - название рекламодателя

### таблица Campaigns

информация о кампаниях

- `campaign_id` uuid - uuid рекламной кампании
- `advertiser_id` uuid - uuid рекламодателя
- `impressions_limit` uint - лимит (желаемое количество) просмотров. всегда >= `clicks_limit`
- `clicks_limit` uint - лимит (желаемое количество) кликов
- `cost_per_impression` float - цена за уникальный просмотр
- `cost_per_click` float - цена за уникальный клик
- `ad_title` string - заголовок рекламы
- `ad_text` string - текст рекламы
- `start_date` uint - начальный день рекламной компании
- `end_date` uint - конечный день рекламной компании. всегда >= `start_date`
- `targeting` - jsonb структура настроек таргетинга объявления
  - `gender` string - необязательное. пол клиента, MALE/FEMALE/ALL
  - `age_from` uint - необязательное. возраст клиента с
  - `age_to` uint - необязательное. возраст клиента по. всегда >= `age_from`
  - `location` string - необязательное. локация клиента
- `image` string - необязательное. используется эндпоинтом получения изображения из S3
- `moderated` bool - используется при включенной модерации объявлений. можно ли показывать объявление (прошло ли оно модерацию)

### таблица MLScores

сопоставления ml скоров рекламодателей и пользователей

- `client_id` uuid - часть составного первичного ключа, uuid клиента
- `advertiser_id` uuid - часть составного первичного ключа, uuid кампании
- `score` uint - ML скор

### таблица Settings

динамические параметры сервиса

- `id` uint - используется всегда id 1
- `day` uint - текущий день в системе
- `moderation` bool - активна ли система модерации
