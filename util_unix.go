// +build !windows

package main

import (
	"github.com/rivo/tview"
)

// do nothing for non windows
// remove once https://github.com/gdamore/tcell/issues/319 is fixed
func (session *viewerSession) enablePaste(field *tview.InputField, form *tview.Form) {
}
