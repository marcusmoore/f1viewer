// +build windows

package main

import (
	"reflect"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// Hacky solution for a working copy paste on Windows
// Breaks encapsulation by accessing a private member using reflect
// remove once https://github.com/gdamore/tcell/issues/319 is fixed
func (session *viewerSession) enablePaste(field *tview.InputField, form *tview.Form) {
	field.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyCtrlV || event.Key() == tcell.KeyCtrlY {
			session.handlePaste(field)
		}
		return event
	})

	// paste on right click
	field.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		item, _ := form.GetFocusedItemIndex()

		if item != -1 && event.Buttons()&tcell.Button2 != 0 && form.GetFormItem(item) == field {
			session.handlePaste(field)
		}

		return action, event
	})
}

func (session *viewerSession) handlePaste(field *tview.InputField) {
	clipContent, err := clipboard.ReadAll()
	if err != nil {
		session.logError("could not paste: ", err)
	}
	val := reflect.ValueOf(*field)
	cursorPos := val.FieldByName("cursorPos").Int()
	text := field.GetText()
	field.SetText(text[0:cursorPos] + clipContent + text[cursorPos:])
	// TODO: why does Draw() deadlock?
	session.app.ForceDraw()
}
