//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

const (
	WS_CHILD       = 0x40000000
	WS_VISIBLE     = 0x10000000
	WS_TABSTOP     = 0x00010000
	ES_AUTOHSCROLL = 0x00000080
	ES_LEFT        = 0x00000000

	SW_SHOW = 5

	WM_ERASEBKGND = 0x0014
	WM_PAINT      = 0x000F
	WM_SIZE       = 0x0005
	WM_SETFONT    = 0x0030
	WM_LBUTTONDOWN = 0x0201
	WM_DESTROY    = 0x0002

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
	EC_LEFTMARGIN   = 0x0001

	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010

	PS_SOLID = 0

	FW_NORMAL = 400
	FW_BOLD   = 700
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
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
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
	procSendMessageW       = user32.NewProc("SendMessageW")
	procShowWindowChild    = user32.NewProc("ShowWindow")
	procSetWindowPos       = user32.NewProc("SetWindowPos")
	procInvalidateRect     = user32.NewProc("InvalidateRect")
)

var (
	searchBox   uintptr
	sidebarOpen = true

	uiFont     uintptr
	uiBoldFont uintptr

	colorWindow      = rgb(255, 255, 255)
	colorSidebar     = rgb(248, 248, 248)
	colorBorder      = rgb(221, 221, 221)
	colorIconBox     = rgb(255, 255, 255)
	colorLogo        = rgb(35, 36, 39)
	colorText        = rgb(55, 55, 58)
	colorMuted       = rgb(112, 112, 116)
	colorSoftAccent  = rgb(244, 244, 245)

	brushWindow  uintptr
	brushSidebar uintptr
	brushWhite   uintptr
	brushAccent  uintptr
	brushLogo    uintptr

	penBorder uintptr
)

const (
	sidebarWidthOpen   = 304
	sidebarWidthClosed = 72

	logoX    = 18
	logoY    = 14
	logoSize = 36

	searchX = 66
	searchY = 15
	searchW = 188
	searchH = 34

	toggleX = 264
	toggleY = 15
	toggleW = 28
	toggleH = 34

	iconSize  = 42
	iconStart = 82
	iconGap   = 58

	settingsBottom = 52
	profileBottom  = 104
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
	h, _, _ := procCreatePen.Call(PS_SOLID, 1, uintptr(color))
	return h
}

func createFont(height, weight int32) uintptr {
	h, _, _ := procCreateFontW.Call(
		uintptr(height),
		0, 0, 0,
		uintptr(weight),
		0, 0, 0,
		0, 0, 0, 0, 0,
		uintptr(unsafe.Pointer(wstr("Segoe UI"))),
	)
	return h
}

func initialiseUIResources() {
	brushWindow = makeBrush(colorWindow)
	brushSidebar = makeBrush(colorSidebar)
	brushWhite = makeBrush(colorIconBox)
	brushAccent = makeBrush(colorSoftAccent)
	brushLogo = makeBrush(colorLogo)
	penBorder = makePen(colorBorder)

	uiFont = createFont(-16, FW_NORMAL)
	uiBoldFont = createFont(-19, FW_BOLD)
}

func releaseUIResources() {
	for _, object := range []uintptr{
		brushWindow, brushSidebar, brushWhite, brushAccent, brushLogo,
		penBorder, uiFont, uiBoldFont,
	} {
		if object != 0 {
			procDeleteObject.Call(object)
		}
	}
}

func setFont(hwnd, font uintptr) {
	if hwnd != 0 && font != 0 {
		procSendMessageW.Call(hwnd, WM_SETFONT, font, 1)
	}
}

func drawText(hdc uintptr, value string, rect RECT, color uint32, font uintptr, flags uintptr) {
	oldFont := uintptr(0)
	if font != 0 {
		oldFont, _, _ = procSelectObject.Call(hdc, font)
	}

	procSetBkMode.Call(hdc, TRANSPARENT)
	procSetTextColor.Call(hdc, uintptr(color))

	ptr := wstr(value)
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

func drawRoundRect(hdc uintptr, rect RECT, brush uintptr, radius int) {
	oldBrush, _, _ := procSelectObject.Call(hdc, brush)
	oldPen, _, _ := procSelectObject.Call(hdc, penBorder)

	procRoundRect.Call(
		hdc,
		uintptr(rect.Left), uintptr(rect.Top),
		uintptr(rect.Right), uintptr(rect.Bottom),
		uintptr(radius), uintptr(radius),
	)

	procSelectObject.Call(hdc, oldBrush)
	procSelectObject.Call(hdc, oldPen)
}

func invalidate(hwnd uintptr) {
	procInvalidateRect.Call(hwnd, 0, 1)
}

func layoutSearchBox() {
	if searchBox == 0 {
		return
	}

	if !sidebarOpen {
		procShowWindowChild.Call(searchBox, 0)
		return
	}

	procShowWindowChild.Call(searchBox, SW_SHOW)
	procSetWindowPos.Call(
		searchBox,
		0,
		uintptr(searchX+2),
		uintptr(searchY+2),
		uintptr(searchW-4),
		uintptr(searchH-4),
		SWP_NOZORDER|SWP_NOACTIVATE,
	)
}

func drawSidebar(hdc uintptr, client RECT) {
	sidebarWidth := sidebarWidthOpen
	if !sidebarOpen {
		sidebarWidth = sidebarWidthClosed
	}

	procSelectObject.Call(hdc, brushWindow)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&client)), brushWindow)

	sidebarRect := RECT{
		Left: 0, Top: 0,
		Right: int32(sidebarWidth), Bottom: client.Bottom,
	}
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&sidebarRect)), brushSidebar)

	if !sidebarOpen {
		drawRoundRect(
			hdc,
			RECT{18, 14, 54, 50},
			brushLogo,
			10,
		)
		drawText(
			hdc,
			"L",
			RECT{18, 14, 54, 50},
			colorWindow,
			uiBoldFont,
			DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
		)

		drawRoundRect(
			hdc,
			RECT{18, 60, 54, 94},
			brushWhite,
			10,
		)
		drawText(
			hdc,
			"≡",
			RECT{18, 60, 54, 94},
			colorText,
			uiBoldFont,
			DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
		)
		return
	}

	// LearningApp mark.
	drawRoundRect(hdc, RECT{logoX, logoY, logoX + logoSize, logoY + logoSize}, brushLogo, 11)
	drawText(
		hdc,
		"L",
		RECT{logoX, logoY, logoX + logoSize, logoY + logoSize},
		colorWindow,
		uiBoldFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Search backing.
	drawRoundRect(
		hdc,
		RECT{searchX - 1, searchY - 1, searchX + searchW + 1, searchY + searchH + 1},
		brushWhite,
		10,
	)

	// Sidebar toggle.
	drawRoundRect(
		hdc,
		RECT{toggleX, toggleY, toggleX + toggleW, toggleY + toggleH},
		brushWhite,
		10,
	)
	drawText(
		hdc,
		"≡",
		RECT{toggleX, toggleY, toggleX + toggleW, toggleY + toggleH},
		colorText,
		uiBoldFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Blank smooth-square navigation placeholders.
	for i := 0; i < 5; i++ {
		y := iconStart + i*iconGap
		drawRoundRect(
			hdc,
			RECT{24, int32(y), 24 + iconSize, int32(y + iconSize)},
			brushWhite,
			11,
		)
	}

	// Compact settings utility at the very bottom.
	settingsY := client.Bottom - settingsBottom
	drawRoundRect(
		hdc,
		RECT{24, settingsY, 56, settingsY + 32},
		brushAccent,
		10,
	)
	drawText(
		hdc,
		"⚙",
		RECT{24, settingsY, 56, settingsY + 32},
		colorMuted,
		uiFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Profile sits immediately above the bottom utility area.
	profileY := client.Bottom - profileBottom
	avatar := RECT{18, profileY + 4, 54, profileY + 40}
	drawRoundRect(hdc, avatar, brushAccent, 18)
	drawText(
		hdc,
		"D",
		avatar,
		colorText,
		uiBoldFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)
	drawText(
		hdc,
		"Darius",
		RECT{66, profileY + 3, int32(sidebarWidth - 18), profileY + 41},
		colorText,
		uiFont,
		DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Small notification utility in the main content header.
	notification := RECT{client.Right - 72, 16, client.Right - 38, 50}
	drawRoundRect(hdc, notification, brushWhite, 11)
	drawText(
		hdc,
		"●",
		notification,
		colorMuted,
		uiFont,
		DT_CENTER|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)

	// Main content placeholder heading.
	drawText(
		hdc,
		"LearningApp",
		RECT{int32(sidebarWidth) + 36, 20, client.Right - 92, 56},
		colorText,
		uiBoldFont,
		DT_LEFT|DT_VCENTER|DT_SINGLELINE|DT_NOPREFIX,
	)
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
		0, 0, 0,
	)
	if edit == 0 {
		return 0
	}

	setFont(edit, uiFont)
	margin := uintptr(8) | (uintptr(8) << 16)
	procSendMessageW.Call(edit, EM_SETMARGINS, EC_LEFTMARGIN, margin)
	procSendMessageW.Call(
		edit,
		EM_SETCUEBANNER,
		0,
		uintptr(unsafe.Pointer(wstr("Search notes..."))),
	)

	return edit
}
