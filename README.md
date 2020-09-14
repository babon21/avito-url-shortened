# avito-url-shortened

## Задача:
   
   Нужно сделать HTTP сервис для сокращения URL наподобие [Bitly](https://bitly.com/) и других сервисов.
   
   UI не нужен, достаточно сделать JSON API сервис.  
   Должна быть возможность: 
   - сохранить короткое представление заданного URL
   - перейти по сохраненному ранее короткому представлению и получить redirect на соответствующий исходный URL

### Усложнения:

- Добавлена валидация URL
- Добавлена возможность задавать кастомные ссылки, чтобы пользователь мог сделать их человекочитаемыми

### Информация по проекту

Для генерации сокращенного url используется Base62 кодирование id, полученный из БД. 

- Язык программирования Golang
- Docker
- PostgreSQL
- Конфигурация хранится yaml файле

## Инструкция по запуску

#### Сборка проекта

Для создания образа из Dockerfile выполняется следующая команда из того же каталога, в котором находится файл docker-compose.yml:

`$ docker-compose build`


#### Запуск проекта

Теперь, когда проект собран, его можно запустить с помощью команды:

`$ docker-compose up`

После выполнения этой команды в терминале должен появиться текст: `"Start http server on [значение порта] port`

### API методы 

***Метод сокращения заданного исходного URL***

Запрос:

```
curl --header "Content-Type: application/json"   
    --request POST   
    --data '{"origin_url": "<URL>"}'   
    http://localhost:8080/api
```

Ответ в случае успеха:
```
{
    "data": [
        {
            "short_url": "some_short_url"
        }
    ]
}
```

Формат ответа в случае ошибки:
```
{
    "errors": [{
        "status": "500",
        "detail": "some description"
    }]
}
```

***Метод задания кастомного URL по заданному сокращенногому URL***

Запрос:

```
curl --header "Content-Type: application/json"   
    --request PATCH   
    --data '{"short_url": "<SHORT_URL>", "custom_url": "<CUSTOM_URL>"}'   
    http://localhost:8080/api
```

Ответ в случае успеха:
```
{
    "data": [
        {
            "short_url": "some_short_url",
            "custom_url": "some_custom_url"
        }
    ]
}
```

Формат ответа в случае ошибки:
```
{
    "errors": [{
        "status": "500",
        "detail": "some description"
    }]
}
```

Сокращение URL по заданному исходному URL, test curl:
```
curl --header "Content-Type: application/json" --request POST --data '{"origin_url": "google.com"}' http://localhost:8080/api
```

Задание кастомного URL по заданному сокращенному URL, test curl:
```
curl --header "Content-Type: application/json" --request PATCH --data '{"short_url": "localhost:8080/1", "custom_url": "some_url"}' http://localhost:8080/api
```