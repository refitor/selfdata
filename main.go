package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"rshub/product/rsapp/selfdata/prowork"

	"github.com/jpillora/overseer"
)

var BuildID string
var DebugCode string
var signalChannel = make(chan os.Signal)

func main() {
	// ready()
	overseer.Run(overseer.Config{
		Required:            false,
		Program:             prog,
		Debug:               false, //display log of overseer actions
		Address:             ":5001",
		NoWarn:              true,
		NoRestart:           true,
		NoRestartAfterFetch: true,
		Fetcher:             prowork.GetFetcher(),
		PreUpgrade:          prowork.GetPreUpgrade(),
	})
}

func ready() {
	box := initRice()
	prowork.Init(box, BuildID, DebugCode, func() {
		//signalChannel <- syscall.SIGQUIT
		os.Exit(0)
	})
}

func prog(state overseer.State) {
	ctx, cancel := context.WithCancel(context.Background())

	//box := initRice()
	//prowork.Init(box, BuildID, DebugCode, func() {
	//	//signalChannel <- syscall.SIGQUIT
	//	os.Exit(0)
	//})
	ready()
	go prowork.Run(ctx)

	runForSignal(func(o os.Signal) {
		cancel()
	})
}

func runForSignal(fExit func(os.Signal)) os.Signal {
	signal.Notify(signalChannel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //, syscall.SIGUSR1, syscall.SIGUSR2)
	sig := <-signalChannel

	fExit(sig)
	return sig
}
