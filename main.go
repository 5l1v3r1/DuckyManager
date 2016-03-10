package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/nsf/termbox-go"
)

func main() {
	// Load lang
	var msg string
	msg, err := checkLangs(os.Args)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(errExitCode)
	}
	if msg != "" {
		fmt.Printf(msg)
		os.Exit(errExitCode)
	}

	if err = parseLang(os.Args[1]); err != nil {
		fmt.Println(err.Error())
		os.Exit(errExitCode)
	}

	// Load config
	cf, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println(translate.ErrOpeningConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	err = json.Unmarshal(cf, &config)
	if err != nil {
		fmt.Println(errStr + translate.ErrParsingConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	// Init log
	f, err := os.OpenFile(config.LogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(errStr + translate.ErrParsingConfig + ": " + err.Error())
		os.Exit(errExitCode)
	}

	l = log.New(f, "", log.Ltime)

	l.SetOutput(f)

	// Load scripts
	l.Println("+------------------------------+")
	l.Println(translate.LoadingLocal)

	scripts, valid, deleted, modified, newOnes, err := CheckLocal(config.LocalDBFile, config.ScriptsPath)

	if err != nil {
		fmt.Println(errStr + translate.ErrCheckingLocal + " : " + err.Error())
		l.Println(errStr + translate.ErrCheckingLocal + " : " + err.Error())
		os.Exit(errExitCode)
	}
	defer func() {
		if err = Save(config.LocalDBFile, scripts); err != nil {
			//TODO handle
		}
	}()

	l.Println("[" + strconv.Itoa(int(valid)) + "] " + translate.Valid + " , " +
		"[" + strconv.Itoa(int(deleted)) + "] " + translate.Deleted + " , " +
		"[" + strconv.Itoa(int(modified)) + "] " + translate.Modified + " , " +
		"[" + strconv.Itoa(int(newOnes)) + "] " + translate.NewScripts)

	// GUI
	err = termbox.Init()
	if err != nil {
		fmt.Println(errStr + translate.ErrTermboxInit + ": " + err.Error())
		l.Println(errStr + translate.ErrTermboxInit + ": " + err.Error())
		os.Exit(errExitCode)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)
	termbox.SetOutputMode(termbox.Output256)
	l.Println(okStr + translate.TermInputMode + ": InputESC || " + translate.TermOutputMode + ": Output256")

	currentState := State{
		Scripts:       scripts,
		Position:      0,
		PositionUpper: 0,
	}

	mainLoop(currentState)
}

func mainLoop(currentState State) {
	var ev termbox.Event
	for ev.Key != termbox.KeyEsc && ev.Key != termbox.KeyCtrlC {

		if err := redrawMain(currentState); err != nil {
			l.Println(errStr + translate.ErrDrawing + ": " + err.Error())
			os.Exit(errExitCode)
		}

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			currentState.SwitchKey(ev)

		case termbox.EventError:
			l.Println(errStr + translate.ErrEvent + ": " + ev.Err.Error())
			os.Exit(errExitCode)
		}

	}
}
