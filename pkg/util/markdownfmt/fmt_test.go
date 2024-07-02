package markdownfmt

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrintfMd(t *testing.T) {
	testStr := "Hello, %m %s"

	tUuid := uuid.New().String()
	vals := []interface{}{tUuid, "world"}
	result := Sprintf(testStr, vals...)
	require.Equal(t, "Hello, "+tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, tUuid)+" world", result)
}
