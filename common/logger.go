package common

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	LogKeyRequestId = "request_id"
)

func InitLogger() {
	consoleWriter := zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stdout
		w.TimeFormat = "2006-01-02 15:04:05"
		w.FormatLevel = func(i interface{}) string {
			s := fmt.Sprintf("%s", i)
			if len(s) > 4 {
				s = s[:4]
			}
			return s
		}
		//w.FormatMessage = func(i interface{}) string {
		//	return fmt.Sprintf("[%s]", i)
		//}
		//w.FormatFieldName = func(i interface{}) string {
		//	return fmt.Sprintf("\n\t%s:", i)
		//}
		//w.FormatFieldValue = func(i interface{}) string {
		//	return fmt.Sprintf("%s", i)
		//}
		//w.FormatExtra = func(m map[string]interface{}, b *bytes.Buffer) error {
		//	// remove request_id from extra fields
		//	delete(m, LOG_KEY_REQUEST_ID)
		//	for k, v := range m {
		//		b.WriteString(fmt.Sprintf("\n\t%s:%v", k, v))
		//	}
		//	return nil
		//}
	})

	log.Logger = log.Output(consoleWriter).With().Caller().Logger()
}
