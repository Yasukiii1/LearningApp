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
	SW_SHOW             = 5

	WM_DESTROY    = 0x0002
	WM_PAINT      = 0x000F
	WM_ERASEBKGND = 0x0014

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001
	IDC_ARROW  = 32512

	DT_CENTER     = 0x00000001
	DT_VCENTER    = 0x00000004
	DT_SINGLELINE = 0x00000020
	DT_NOPREFIX   = 0x00000800
	TRANSPARENT   = 1
)

type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type PAINTSTRUCT struct {
	Hdc         uintptr
	Erase       int32
	RcPaint     RECT
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procRegisterClassExW = user32.NewProc("RegisterClassExW")
	procCreateWindowExW = user32.NewProc("CreateWindowExW")
	procDefWindowProcW  = user32.NewProc("DefWindowProcW")
	procShowWindow      = user32.NewProc("ShowWindow")
	procUpdateWindow    = user32.NewProc("UpdateWindow")
	procGetMessageW     = user32.NewProc("GetMessageW")
	procTranslateMessage = user32.NewProc("TranslateMessage")
	procDispatchMessageW = user32.NewProc("DispatchMessageW")
	procPostQuitMessage  = user32.NewProc("PostQuitMessage")
	procLoadCursorW      = user32.NewProc("LoadCursorW")

	procBeginPaint      = user32.NewProc("BeginPaint")
	procEndPaint        = user32.NewProc("EndPaint")
	procGetClientRect   = user32.NewProc("GetClientRect")
	procFillRect        = user32.NewProc("FillRect")
	procSetBkMode       = gdi32.NewProc("SetBkMode")
	procSetTextColor    = gdi32.NewProc("SetTextColor")
	procDrawTextW       = user32.NewProc("DrawTextW")
	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
)

func wstr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func rgb(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_ERASEBKGND:
		return 1

	case WM_PAINT:
		var ps PAINTSTRUCT
		hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))

		var rc RECT
		procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))

		// White base.
		whiteBrush, _, _ := procCreateSolidBrush.Call(uintptr(rgb(255, 255, 255)))
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&rc)), whiteBrush)
		procDeleteObject.Call(whiteBrush)

		// Soft yellow header.
		header := RECT{
			Left:   0,
			Top:    0,
			Right:  rc.Right,
			Bottom: 96,
		}
		yellowBrush, _, _ := procCreateSolidBrush.Call(uintptr(rgb(255, 244, 190)))
		procFillRect.Call(hdc, uintptr(unsafe.Pointer(&header)), yellowBrush)
		procDeleteObject.Call(yellowBrush)

		// Dark text on the light background.
		procSetBkMode.Call(hdc, TRANSPARENT)
		procSetTextColor.Call(hdc, uintptr(rgb(55, 55, 55)))

		titleRect := RECT{
			Left:   24,
			Top:    18,
			Right:  rc.Right - 24,
			Bottom: 76,
		}
		title := wstr("LearningApp")
		procDrawTextW.Call(
			hdc,
			uintptr(unsafe.Pointer(title)),
			uintptr(^uint(0)>>1),
			uintptr(unsafe.Pointer(&titleRect)),
			DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
		)

		procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
		return 0

	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func main() {
	className := wstr("LearningAppWindow")
	windowTitle := wstr("LearningApp")

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HCursor:       func() uintptr { r, _, _ := procLoadCursorW.Call(0, IDC_ARROW); return r }(),
		LpszClassName: className,
	}

	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW failed")
	}

	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(windowTitle)),
		WS_OVERLAPPEDWINDOW|WS_VISIBLE,
		CW_USEDEFAULT,
		CW_USEDEFAULT,
		900,
		600,
		0,
		0,
		0,
		0,
	)

	if hwnd == 0 {
		panic("CreateWindowExW failed")
	}

	procShowWindow.Call(hwnd, SW_SHOW)
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
