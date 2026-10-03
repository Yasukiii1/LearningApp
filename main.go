//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00CF0000
	WS_VISIBLE          = 0x10000000
	WS_CHILD            = 0x40000000
	WS_BORDER            = 0x00800000
	WS_TABSTOP           = 0x00010000
	ES_AUTOHSCROLL      = 0x0080
	ES_LEFT             = 0x0000

	CW_USEDEFAULT = 0x80000000
	SW_SHOW        = 5

	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_SIZE        = 0x0005
	WM_ERASEBKGND  = 0x0014
	WM_LBUTTONDOWN = 0x0201
	WM_SETFONT     = 0x0030
	WM_SETFOCUS    = 0x0007
	WM_KILLFOCUS   = 0x0008

	CS_HREDRAW = 0x0002
	CS_VREDRAW = 0x0001
	IDC_ARROW  = 32512

	TRANSPARENT = 1

	DT_LEFT       = 0x00000000
	DT_CENTER     = 0x00000001
	DT_VCENTER    = 0x00000004
	DT_SINGLELINE = 0x00000020
	DT_NOPREFIX   = 0x00000800

	EM_SETCUEBANNER = 0x1501
	EM_SETMARGINS   = 0x00D3

	EC_LEFTMARGIN = 0x0001
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

type RECT struct {
	Left, Top, Right, Bottom int32
}

type PAINTSTRUCT struct {
	Hdc         uintptr
	Erase       int32
	RcPaint     RECT
	Restore     int32
	IncUpdate   int32
	RgbReserved [32]byte
}

type WNDCLASSEX struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra     int32
	CbWndExtra     int32
	HInstance      uintptr
	HIcon          uintptr
	HCursor        uintptr
	HbrBackground uintptr
	LpszMenuName   *uint16
	LpszClassName  *uint16
	HIconSm        uintptr
}

var (
	user32 = syscall.NewLazyDLL("user32.dll")
	gdi32  = syscall.NewLazyDLL("gdi32.dll")

	procRegisterClassExW   = user32.NewProc("RegisterClassExW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procDefWindowProcW     = user32.NewProc("DefWindowProcW")
	procShowWindow         = user32.NewProc("ShowWindow")
	procUpdateWindow       = user32.NewProc("UpdateWindow")
	procGetMessageW        = user32.NewProc("GetMessageW")
	procTranslateMessage  = user32.NewProc("TranslateMessage")
	procDispatchMessageW   = user32.NewProc("DispatchMessageW")
	procPostQuitMessage    = user32.NewProc("PostQuitMessage")
	procLoadCursorW        = user32.NewProc("LoadCursorW")
	procSetCursor          = user32.NewProc("SetCursor")
	procGetClientRect      = user32.NewProc("GetClientRect")
	procBeginPaint         = user32.NewProc("BeginPaint")
	procEndPaint           = user32.NewProc("EndPaint")
	procFillRect           = user32.NewProc("FillRect")
	procCreateSolidBrush   = gdi32.NewProc("CreateSolidBrush")
	procCreatePen          = gdi32.NewProc("CreatePen")
	procSelectObject       = gdi32.NewProc("SelectObject")
	procDeleteObject       = gdi32.NewProc("DeleteObject")
	procRoundRect          = gdi32.NewProc("RoundRect")
	procSetBkMode          = gdi32.NewProc("SetBkMode")
	procSetTextColor       = gdi32.NewProc("SetTextColor")
	procDrawTextW          = user32.NewProc("DrawTextW")
	procCreateFontW        = gdi32.NewProc("CreateFontW")
	procCreateWindowExW    = user32.NewProc("CreateWindowExW")
	procSendMessageW       = user32.NewProc("SendMessageW")
	procShowWindowChild    = user32.NewProc("ShowWindow")
	procSetWindowPos       = user32.NewProc("SetWindowPos")
	procGetDC              = user32.NewProc("GetDC")
	procReleaseDC          = user32.NewProc("ReleaseDC")
)

var (
	mainWindow uintptr
	searchBox  uintptr
	sidebarOpen = true

	uiFont     uintptr
	uiBoldFont uintptr
)

const (
	sidebarWidthOpen   = 290
	sidebarWidthClosed = 76

	topBarHeight = 64

	searchX = 18
	searchY = 15
	searchW = 210
	searchH = 34

	toggleX = 238
	toggleY = 15
	toggleW = 34
	toggleH = 34

	iconSize  = 42
	iconStart = 92
	iconGap   = 58
)

func wstr(s string) *uint16 {
	p, _ := syscall.UTF16PtrFromString(s)
	return p
}

func rgb(r, g, b byte) uint32 {
	return uint32(r) | uint32(g)<<8 | uint32(b)<<16
}

func lowWord(v uintptr) int32 {
	return int32(int16(v & 0xFFFF))
}

func highWord(v uintptr) int32 {
	return int32(int16((v >> 16) & 0xFFFF))
}

func pointInRect(x, y int32, left, top, right, bottom int32) bool {
	return x >= left && x < right && y >= top && y < bottom
}

func makeBrush(color uint32) uintptr {
	h, _, _ := procCreateSolidBrush.Call(uintptr(color))
	return h
}

func makePen(color uint32) uintptr {
	h, _, _ := procCreatePen.Call(0, 1, uintptr(color))
	return h
}

func createFont(height, weight int32) uintptr {
	font, _, _ := procCreateFontW.Call(
		uintptr(height),
		0,
		0,
		0,
		uintptr(weight),
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		0,
		uintptr(unsafe.Pointer(wstr("Segoe UI"))),
	)
	return font
}

func setFont(hwnd, font uintptr) {
	if hwnd == 0 || font == 0 {
		return
	}
	procSendMessageW.Call(hwnd, WM_SETFONT, font, 1)
}

func drawText(hdc uintptr, text string, rect RECT, color uint32, font uintptr, flags uintptr) {
	var oldFont uintptr
	if font != 0 {
		oldFont, _, _ = procSelectObject.Call(hdc, font)
	}

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(color))

	ptr := wstr(text)
	procDrawTextW.Call(
		hdc,
		uintptr(unsafe.Pointer(ptr)),
		uintptr(^uint(0)>>1),
		uintptr(unsafe.Pointer(&rect)),
		flags,
	)

	if oldFont != 0 {
		procSelectObject.Call(hdc, oldFont)
	}
}

func fillRoundRect(hdc uintptr, rect RECT, fill uint32, border uint32, radius int) {
	brush := makeBrush(fill)
	pen := makePen(border)

	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, pen)

	procRoundRect.Call(
		hdc,
		uintptr(rect.Left),
		uintptr(rect.Top),
		uintptr(rect.Right),
		uintptr(rect.Bottom),
		uintptr(radius),
		uintptr(radius),
	)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}

func layoutSearchBox() {
	if searchBox == 0 || mainWindow == 0 {
		return
	}

	width := sidebarWidthOpen
	if !sidebarOpen {
		procShowWindowChild.Call(searchBox, 0)
		return
	}

	procShowWindowChild.Call(searchBox, SW_SHOW)
	procSetWindowPos.Call(
		searchBox,
		0,
		searchX,
		searchY,
		searchW,
		searchH,
		0x0010|0x0020,
	)

	_ = width
}

func drawSidebar(hdc uintptr, client RECT) {
	sidebarWidth := sidebarWidthOpen
	if !sidebarOpen {
		sidebarWidth = sidebarWidthClosed
	}

	// Main white canvas.
	base := makeBrush(rgb(255, 255, 255))
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&client)), base)
	procDeleteObject.Call(base)

	// Sidebar surface.
	sidebarRect := RECT{
		Left:   0,
		Top:    0,
		Right:  int32(sidebarWidth),
		Bottom: client.Bottom,
	}
	sidebarBrush := makeBrush(rgb(255, 249, 224))
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&sidebarRect)), sidebarBrush)
	procDeleteObject.Call(sidebarBrush)

	if !sidebarOpen {
		drawText(
			hdc,
			"☰",
			RECT{18, 15, 58, 55},
			rgb(75, 70, 55),
			uiBoldFont,
			DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
		)
		return
	}

	// Search field background/border.
	fillRoundRect(
		hdc,
		RECT{searchX - 1, searchY - 1, searchX + searchW + 1, searchY + searchH + 1},
		rgb(238, 229, 200),
		rgb(238, 229, 200),
		10,
	)

	// Toggle.
	fillRoundRect(
		hdc,
		RECT{toggleX, toggleY, toggleX + toggleW, toggleY + toggleH},
		rgb(255, 239, 170),
		rgb(232, 211, 122),
		10,
	)
	drawText(
		hdc,
		"☰",
		RECT{toggleX, toggleY, toggleX + toggleW, toggleY + toggleH},
		rgb(75, 70, 55),
		uiBoldFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Five blank smooth square placeholders.
	for i := 0; i < 5; i++ {
		y := iconStart + i*iconGap
		fillRoundRect(
			hdc,
			RECT{24, int32(y), 24 + iconSize, int32(y + iconSize)},
			rgb(255, 244, 190),
			rgb(232, 211, 122),
			11,
		)
	}

	// Profile row above settings.
	profileY := client.Bottom - 94
	avatar := RECT{18, profileY + 5, 54, profileY + 41}
	fillRoundRect(hdc, avatar, rgb(255, 220, 110), rgb(238, 201, 89), 18)
	drawText(
		hdc,
		"D",
		avatar,
		rgb(85, 70, 25),
		uiBoldFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)
	drawText(
		hdc,
		"Darius",
		RECT{66, profileY + 4, sidebarWidth - 18, profileY + 42},
		rgb(55, 55, 55),
		uiFont,
		DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Settings placeholder at bottom.
	settingsY := client.Bottom - 48
	fillRoundRect(
		hdc,
		RECT{18, settingsY, sidebarWidth - 18, settingsY + 34},
		rgb(255, 244, 190),
		rgb(232, 211, 122),
		10,
	)
	drawText(
		hdc,
		"Settings",
		RECT{34, settingsY, sidebarWidth - 24, settingsY + 34},
		rgb(75, 70, 55),
		uiFont,
		DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// App content title.
	drawText(
		hdc,
		"LearningApp",
		RECT{sidebarWidth + 36, 22, client.Right - 36, 58},
		rgb(55, 55, 55),
		uiBoldFont,
		DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_SIZE:
		layoutSearchBox()
		Invalidate(hwnd)
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
			if searchBox != 0 {
				procShowWindowChild.Call(searchBox, 0)
			}
			Invalidate(hwnd)
			return 0
		}

		if !sidebarOpen && pointInRect(x, y, 16, 14, 60, 58) {
			sidebarOpen = true
			layoutSearchBox()
			Invalidate(hwnd)
			return 0
		}

		return 0

	case WM_DESTROY:
		if uiFont != 0 {
			procDeleteObject.Call(uiFont)
		}
		if uiBoldFont != 0 {
			procDeleteObject.Call(uiBoldFont)
		}
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

func Invalidate(hwnd uintptr) {
	user32.NewProc("InvalidateRect").Call(hwnd, 0, 1)
}

func createSearchBox(parent uintptr) uintptr {
	edit, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(wstr("EDIT"))),
		uintptr(unsafe.Pointer(wstr(""))),
		WS_CHILD|WS_VISIBLE|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL,
		uintptr(searchX+3),
		uintptr(searchY+3),
		uintptr(searchW-6),
		uintptr(searchH-6),
		parent,
		0,
		0,
		0,
	)

	if edit != 0 {
		setFont(edit, uiFont)
		margin := uintptr(6) | (uintptr(6) << 16)
		procSendMessageW.Call(edit, EM_SETMARGINS, EC_LEFTMARGIN, margin)
		procSendMessageW.Call(edit, EM_SETCUEBANNER, 0, uintptr(unsafe.Pointer(wstr("Search notes..."))))
	}

	return edit
}

func main() {
	className := wstr("LearningAppWindow")
	windowTitle := wstr("LearningApp")

	uiFont = createFont(-16, 400)
	uiBoldFont = createFont(-19, 700)

	wc := WNDCLASSEX{
		CbSize:        uint32(unsafe.Sizeof(WNDCLASSEX{})),
		Style:         CS_HREDRAW | CS_VREDRAW,
		LpfnWndProc:   syscall.NewCallback(wndProc),
		HCursor:       func() uintptr {
			cursor, _, _ := procLoadCursorW.Call(0, IDC_ARROW)
			return cursor
		}(),
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
		1200,
		760,
		0,
		0,
		0,
		0,
	)

	if hwnd == 0 {
		panic("CreateWindowExW failed")
	}

	mainWindow = hwnd

	searchBox = createSearchBox(hwnd)

	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	layoutSearchBox()

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
