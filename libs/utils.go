package libs

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func RunForever() error {
	var err error
	signals := []os.Signal{os.Interrupt, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT,
		syscall.SIGUSR2, syscall.SIGKILL}
	sCtx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(sCtx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case sig := <-signalChan:
			fmt.Println("receive sig: ", sig)
			switch sig {
			case syscall.SIGUSR2:
				// TODO
				return nil
			default:
				return nil
			}
		}
	})
	if err = eg.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		cancel()
		return err
	}
	cancel()
	return nil
}
