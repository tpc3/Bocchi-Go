package chat

import (
	"context"
	"log"

	"github.com/bounoable/deepl"
	"github.com/tpc3/Bocchi-Go/lib/config"
)

func Deepl(msg *string) string {
	client := deepl.New(config.CurrentConfig.Chat.DeepLToken)
	translated, _, err := client.Translate(
		context.TODO(),
		*msg,
		deepl.Japanese,
	)
	if err != nil {
		log.Fatal("Translating response error: ", err)
	}
	return translated
}
