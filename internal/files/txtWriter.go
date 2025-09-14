package files

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

type NestedMap map[string]map[string]string

func TxtWriter(data NestedMap) {
	file, err := os.OpenFile("result.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Пишем в файл текущую дату и время
	now := time.Now()
	timestamp := now.Format("02.01.2006- 15:04:05")
	writeLine(writer, fmt.Sprintf("### %s\n", timestamp))

	for d, r := range data {
		writeLine(writer, fmt.Sprintf("File: %s\n\n", d))
		for s, c := range r {
			writeLine(writer, fmt.Sprintf("Source: %s\n%s\n", s, c))
		}
		writeLine(writer, fmt.Sprintf("---------------------------------------------------------------------\n"))
	}

	// Сбрасываем буфер в файл
	writer.Flush()
}

func writeLine(w *bufio.Writer, s string) {
	_, err := w.WriteString(s)
	if err != nil {
		panic(err)
	}
}
