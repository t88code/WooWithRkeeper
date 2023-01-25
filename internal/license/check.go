package check

import (
	"os"
	"time"
)

func Check() {
	tm := time.Date(2024, time.February, 2, 0, 0, 0, 0, time.UTC)

	if time.Now().Sub(tm) > 0 {
		os.Exit(1)
	}
}
