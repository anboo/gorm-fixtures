package gorm_fixtures

import (
	"fmt"

	"github.com/schollz/progressbar/v3"
)

func createProgressBar(totalFixtures int) *progressbar.ProgressBar {
	return progressbar.NewOptions(totalFixtures,
		progressbar.OptionSetWriter(ansiWriter{}),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetDescription("[INFO] Loading fixtures"),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionEnableColorCodes(true),
	)
}

type ansiWriter struct{}

// Write записывает данные в терминал с использованием ANSI цветов.
func (w ansiWriter) Write(p []byte) (n int, err error) {
	return fmt.Print(string(p))
}
