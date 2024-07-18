package markdownfmt

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

func Sprintf(format string, a ...interface{}) string {
	outStr := strings.Builder{}
	valLen := len(a)
	valPos := 0

	//Escape the format string in case it has markdown characters
	format = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, format)

	artifacts := make([]interface{}, 0)
	for i := 0; i < len(format); {
		if valPos > valLen {
			break
		}

		if format[i] == '%' {
			i++
			if format[i] != 'm' {
				outStr.WriteString(format[i-1 : i+1])
				artifacts = append(artifacts, a[valPos])
				i++
				continue
			}
			i++

			mkdStr, ok := a[valPos].(string)
			if !ok {
				continue
			}

			mkdStr = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, mkdStr)

			outStr.WriteString(mkdStr)
			valPos++
		} else {
			outStr.WriteByte(format[i])
			i++
		}
	}

	return fmt.Sprintf(outStr.String(), artifacts...)
}
