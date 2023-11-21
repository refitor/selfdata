package main

import (
	"fmt"
	rice "github.com/GeertJohan/go.rice"
)

func initRice() *rice.Box {
	conf := rice.Config{
		LocateOrder: []rice.LocateMethod{rice.LocateEmbedded, rice.LocateAppended, rice.LocateFS},
	}
	box, err := conf.FindBox("datas")
	if err != nil {
		fmt.Printf("failed to initialize memory file system, detail: %s\n", err.Error())
		return nil
	}
	return box
}
