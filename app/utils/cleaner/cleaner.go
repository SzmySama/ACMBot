package cleaner

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/SzmySama/ACMBot/app/render"
)

func CleanWhileExit() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	clean := func() {
		<-sigs
		_ = render.ShutdownBowers()
		os.Exit(0)
	}

	go clean()
}
