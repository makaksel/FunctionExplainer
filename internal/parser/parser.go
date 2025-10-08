package parser

import (
	"fmt"
	"strings"

	"github.com/makaksel/FunctionExplainerGo/internal/files"
	"github.com/waigani/diffparser"
)

var lineLimit = 300

func checkFileName(s string, excludedPaths []string) bool {
	skip := false
	for _, dis := range excludedPaths {
		if strings.Contains(s, dis) {
			skip = true
			break
		}
	}
	return skip
}

func hasImportExport(s string) bool {
	prefixes := []string{"import", "export *", "export {"}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

type ParsedFiles struct {
	DiffFile string
	SrcFile  string
	Data     string
}

func ParsingDiff(data []files.FileData, excludedPaths []string) []ParsedFiles {
	var result []ParsedFiles
	var fileChunk = make(map[string]string)
	for _, diffFile := range data {

		diff, _ := diffparser.Parse(diffFile.Data)

		for _, srcFile := range diff.Files {
			skip := checkFileName(srcFile.NewName, excludedPaths) || srcFile.Mode == 0
			if skip {
				continue
			}

			if srcFile.NewName == "" {
				fmt.Printf("New file name is empty. Skip hunks!!!\n Diff FileName: %s;\n Diff Mode: %v;\n OrigName: %v\n\n", diffFile.FileName, srcFile.Mode, srcFile.OrigName)
				continue
			}

			for _, hunk := range srcFile.Hunks {
				lines := 0

				for _, line := range hunk.NewRange.Lines {
					switch {
					// Если строка пустая и превышен лимит строк, то сделать отсечку
					case lines >= lineLimit && len(line.Content) == 0:
						result = append(result, ParsedFiles{
							DiffFile: diffFile.FileName,
							SrcFile:  srcFile.NewName,
							Data:     fileChunk[srcFile.NewName],
						})
						fileChunk[srcFile.NewName] = ""

					// Если строка не пустая и это не импорт и не экспорт, то добавляем изменения
					case len(line.Content) != 0 && !hasImportExport(line.Content):
						fileChunk[srcFile.NewName] += line.Content
					}

					lines++
				}
			}

			if fileChunk[srcFile.NewName] != "" {
				result = append(result, ParsedFiles{
					DiffFile: diffFile.FileName,
					SrcFile:  srcFile.NewName,
					Data:     fileChunk[srcFile.NewName],
				})
			}

		}
	}
	return result
}
