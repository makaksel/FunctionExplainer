package main

import (
	"github.com/makaksel/FunctionExplainerGo/config"
	"github.com/makaksel/FunctionExplainerGo/internal/ai"
	"github.com/makaksel/FunctionExplainerGo/internal/files"
	"github.com/makaksel/FunctionExplainerGo/internal/parser"
)

func main() {
	env := config.Load()

	// Инициализируем claude client
	aiClient := ai.NewClient(env.Token, env.SystemPrompt, env.Proxy)

	// Загрузка файлов
	loadedFiles := files.Load()

	// Очистка изменений от лишнего и разделение на части, что бы не отдавать АИ все
	data := parser.ParsingDiff(loadedFiles, env.ExcludedPaths)

	// Отдаем аи по частям
	result := aiClient.ProcessAll(data)

	// Пишем результат txt
	files.TxtWriter(result)
}
