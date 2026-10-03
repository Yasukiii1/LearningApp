//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	CW_USEDEFAULT       = 0x80000000
	SW_SHOWMAXIMIZED    = 3

	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_SIZE        = 0x0005
	WM_ERASEBKGND  = 0x0014
	WM_LBUTTONDOWN = 0x0201

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001
	IDC_ARROW  = 32512
)

var mainWindow uintptr

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_SIZE:
		layoutSearchBox()
		invalidate(hwnd)
		return 0

	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		var client RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&client)))

		drawSidebar(hdc, client)

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_LBUTTONDOWN:
		x := lowWord(lParam)
		y := highWord(lParam)

		if sidebarOpen && pointInRect(x, y, toggleX, toggleY, toggleX+toggleW, toggleY+toggleH) {
			sidebarOpen = false
			layoutSearchBox()
			invalidate(hwnd)
			return 0
		}

		if !sidebarOpen && pointInRect(x, y, 17, 58, 56, 98) {
			sidebarOpen = true
			layoutSearchBox()
			invalidate(hwnd)
			return 0
		}

		// Navigation, profile, settings, and notification are placeholders.
		// They intentionally do nothing until their functionality is planned.
		return 0

	case WM_DESTROY:
		releaseUIResources()
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func main() {
	className := wstr("LearningAppWindow")
	windowTitle := wstr("LearningApp")

	initialiseUIResources()

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HCursor: func() uintptr {
			cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
			return cursor
		}(),
		LpszClassName: className,
	}

	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		releaseUIResources()
		panic("RegisterClassExW failed")
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowTitle)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		1200,
		760,
		0, 0, 0, 0,
	)

	if hwnd == 0 {
		releaseUIResources()
		panic("CreateWindowExW failed")
	}

	mainWindow = hwnd
	searchBox = createSearchBox(hwnd)

	procShowWindow.Call(hwnd, SW_SHOWMAXIMIZED)
	procUpdateWindow.Call(hwnd)

	var msg MSG
	for {
		result, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(result) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}
