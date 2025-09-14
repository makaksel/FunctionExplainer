### Function Explainer

Инструмент на Go, который:
1. Парсит `diff` файлов (например, из `git diff`).
2. Отправляет код в **Claude AI** для анализа.
3. Формирует результат в `result.txt`, где содержится объяснение функций и изменений.

## Структура проекта
```
.
├── cmd/                # Точка входа (main.go)
│ └── explainer/        # CLI-приложение
├── config/             # Конфигурация
├── diffs/              # Папка для diff-файлов
├── internal/           # Внутренние пакеты
│ ├── ai/               # Работа с AI (Claude)
│ ├── files/            # Чтение/запись файлов
│ └── parser/           # Парсер diff
├── .env                # Переменные окружения
├── .env.example        # Пример env-конфига
├── .gitignore
├── go.mod
├── README.md
└── result.txt          # Результат работы
```

## Установка и запуск

```bash
git clone https://github.com/username/FunctionExplainerGo.git
cd FunctionExplainerGo

go mod tidy
```

Создайте файл .env на основе .env.example и укажите ключ API для Claude:
```
CLAUDE_API_KEY=your_api_key_here    
PROXY=http://user:pass@ip:port      # необходимо если запускаете проект из заблокированого региона
EXCLUDED_PATHS=                     # части пути, что бы исключать файлы из обработки
SYS_PROMPT=                         # промпт для AI
```

Запуск
```
go run ./cmd/explainer
```