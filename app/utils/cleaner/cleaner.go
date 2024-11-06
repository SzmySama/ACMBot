package cleaner

import (
	"github.com/YourSuzumiya/ACMBot/app/model/render"
	"os"
	"os/signal"
	"syscall"
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
