package files

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileData struct {
	DiffFile string
	FileName string
	Data     string
}

func Load() []FileData {
	var result []FileData

	err := filepath.Walk("diffs",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil // Проигнорируем директории
			}

			data, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("ошибка чтения файла: %s", err)
			}

			result = append(result, FileData{
				FileName: info.Name(),
				Data:     string(data),
			})
			return nil
		})
	if err != nil {
		fmt.Println("Ошибка обхода папки loadFiles:", err)
		panic(err)
		return nil
	}

	return result
}
