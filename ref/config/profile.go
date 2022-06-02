package config

import (
	"github.com/jchavannes/jgo/jutil"
	"github.com/pkg/profile"
	"os"
	"os/signal"
	"syscall"
)

func SetProfileSignalListener() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGUSR1)
	var profiler interface{ Stop() }
	for range signalChan {
		if jutil.IsNil(profiler) {
			profiler = profile.Start(profile.NoShutdownHook)
		} else {
			profiler.Stop()
			profiler = nil
		}
	}
}
