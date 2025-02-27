package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

func (session *viewerSession) checkCommands(commands ...string) {
	var found int
	for _, cmd := range commands {
		_, err := exec.LookPath(cmd)
		session.commands[cmd] = err == nil
		if err == nil {
			found++
		} else {
			session.logInfo("could not find ", cmd)
		}
	}
	if found == 0 {
		session.logError("Both MPV and VLC are unavailable!")
	}
}

func (session *viewerSession) commandAvailable(command string) bool {
	available, ok := session.commands[command]
	return ok && available
}

// takes year/race ID and returns full year and race nuber as strings
func getYearAndRace(input string) (string, string, error) {
	var fullYear string
	var raceNumber string
	if len(input) < 4 {
		return fullYear, raceNumber, errors.New("not long enough")
	}
	_, err := strconv.Atoi(input[:4])
	if err != nil {
		return fullYear, raceNumber, errors.New("not a valid YearRaceID")
	}
	// TODO fix before 2020
	if input[:4] == "2018" || input[:4] == "2019" {
		return input[:4], "0", nil
	}
	year := input[:2]
	intYear, _ := strconv.Atoi(year)
	// TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + year
	} else {
		fullYear = "19" + year
	}
	raceNumber = input[2:4]
	return fullYear, raceNumber, nil
}

func (session *viewerSession) logError(v ...interface{}) {
	if session.textWindow != nil {
		fmt.Fprintln(session.textWindow, fmt.Sprintf("[%s::b]ERROR:[-::-]", colortoHexString(activeTheme.ErrorColor)), fmt.Sprint(v...))
	}
	log.Println("[ERROR]", fmt.Sprint(v...))
}

func (session *viewerSession) logInfo(v ...interface{}) {
	if session.textWindow != nil {
		fmt.Fprintln(session.textWindow, fmt.Sprintf("[%s::b]INFO:[-::-]", colortoHexString(activeTheme.InfoColor)), fmt.Sprint(v...))
	}
	log.Println("[INFO]", fmt.Sprint(v...))
}

func (session *viewerSession) withBlink(node *tview.TreeNode, fn func(), after func()) func() {
	return func() {
		done := make(chan struct{})
		go func() {
			fn()
			done <- struct{}{}
		}()
		go func() {
			session.blinkNode(node, done)
			if after != nil {
				after()
			}
		}()
	}
}

func (session *viewerSession) blinkNode(node *tview.TreeNode, done chan struct{}) {
	originalText := node.GetText()
	originalColor := node.GetColor()
	color1 := originalColor
	color2 := activeTheme.LoadingColor
	node.SetText("loading...")

	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-done:
			node.SetText(originalText)
			node.SetColor(originalColor)
			session.app.Draw()
			return
		case <-ticker.C:
			node.SetColor(color2)
			session.app.Draw()
			c := color1
			color1 = color2
			color2 = c
		}
	}
}

func hexStringToColor(hex string) tcell.Color {
	hex = strings.ReplaceAll(hex, "#", "")
	//TODO: check err?
	color, _ := strconv.ParseInt(hex, 16, 32)
	return tcell.NewHexColor(int32(color))
}

func colortoHexString(color tcell.Color) string {
	return fmt.Sprintf("#%06x", color.Hex())
}

func (t theme) apply() {
	if t.TerminalTextColor != "" {
		tview.Styles.PrimaryTextColor = hexStringToColor(t.TerminalTextColor)
	}
	if t.CategoryNodeColor != "" {
		activeTheme.CategoryNodeColor = hexStringToColor(t.CategoryNodeColor)
	}
	if t.FolderNodeColor != "" {
		activeTheme.FolderNodeColor = hexStringToColor(t.FolderNodeColor)
	}
	if t.ItemNodeColor != "" {
		activeTheme.ItemNodeColor = hexStringToColor(t.ItemNodeColor)
	}
	if t.ActionNodeColor != "" {
		activeTheme.ActionNodeColor = hexStringToColor(t.ActionNodeColor)
	}
	if t.BackgroundColor != "" {
		tview.Styles.PrimitiveBackgroundColor = hexStringToColor(t.BackgroundColor)
	}
	if t.BorderColor != "" {
		tview.Styles.BorderColor = hexStringToColor(t.BorderColor)
	}
	if t.NoContentColor != "" {
		activeTheme.NoContentColor = hexStringToColor(t.NoContentColor)
	}
	if t.LoadingColor != "" {
		activeTheme.LoadingColor = hexStringToColor(t.LoadingColor)
	}
	if t.LiveColor != "" {
		activeTheme.LiveColor = hexStringToColor(t.LiveColor)
	}
	if t.UpdateColor != "" {
		activeTheme.UpdateColor = hexStringToColor(t.UpdateColor)
	}
	if t.TerminalAccentColor != "" {
		activeTheme.TerminalAccentColor = hexStringToColor(t.TerminalAccentColor)
	}
	if t.TerminalTextColor != "" {
		activeTheme.TerminalTextColor = hexStringToColor(t.TerminalTextColor)
	}
	if t.InfoColor != "" {
		activeTheme.InfoColor = hexStringToColor(t.InfoColor)
	}
	if t.ErrorColor != "" {
		activeTheme.ErrorColor = hexStringToColor(t.ErrorColor)
	}
	if t.MultiCommandColor != "" {
		activeTheme.MultiCommandColor = hexStringToColor(t.MultiCommandColor)
	}
}

func sanitizeFileName(s string) string {
	whitespace := regexp.MustCompile(`\s+`)
	var illegal *regexp.Regexp
	if runtime.GOOS == "windows" {
		illegal = regexp.MustCompile(`[<>:"/\\|?*]`)
	} else {
		illegal = regexp.MustCompile(`/`)
	}
	s = illegal.ReplaceAllString(s, " ")
	s = whitespace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}
