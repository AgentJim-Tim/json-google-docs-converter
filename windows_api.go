//go:build windows

package main

import (
    "sync"
    "syscall"
)

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	gdi32    = syscall.NewLazyDLL("gdi32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	shell32  = syscall.NewLazyDLL("shell32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW              = user32.NewProc("RegisterClassExW")
	procCreateWindowExW               = user32.NewProc("CreateWindowExW")
	procDefWindowProcW                = user32.NewProc("DefWindowProcW")
	procShowWindow                    = user32.NewProc("ShowWindow")
	procUpdateWindow                  = user32.NewProc("UpdateWindow")
	procGetMessageW                   = user32.NewProc("GetMessageW")
	procTranslateMessage              = user32.NewProc("TranslateMessage")
	procDispatchMessageW              = user32.NewProc("DispatchMessageW")
	procPostQuitMessage               = user32.NewProc("PostQuitMessage")
	procBeginPaint                    = user32.NewProc("BeginPaint")
	procEndPaint                      = user32.NewProc("EndPaint")
	procFillRect                      = user32.NewProc("FillRect")
	procDrawTextW                     = user32.NewProc("DrawTextW")
	procInvalidateRect                = user32.NewProc("InvalidateRect")
	procSetCursor                     = user32.NewProc("SetCursor")
	procLoadCursorW                   = user32.NewProc("LoadCursorW")
	procGetClientRect                 = user32.NewProc("GetClientRect")
	procSetWindowPos                  = user32.NewProc("SetWindowPos")
	procReleaseCapture                = user32.NewProc("ReleaseCapture")
	procSendMessageW                  = user32.NewProc("SendMessageW")
	procGetForegroundWindow           = user32.NewProc("GetForegroundWindow")
	procGetWindowTextW                = user32.NewProc("GetWindowTextW")
	procIsIconic                      = user32.NewProc("IsIconic")
	procShowWindowAsync               = user32.NewProc("ShowWindowAsync")
	procSendInput                     = user32.NewProc("SendInput")
	procSetProcessDpiAwarenessContext = user32.NewProc("SetProcessDpiAwarenessContext")
	procGetDpiForWindow               = user32.NewProc("GetDpiForWindow")
	procOpenClipboard                 = user32.NewProc("OpenClipboard")
	procCloseClipboard                = user32.NewProc("CloseClipboard")
	procEmptyClipboard                = user32.NewProc("EmptyClipboard")
	procSetClipboardData              = user32.NewProc("SetClipboardData")
	procGetClipboardData              = user32.NewProc("GetClipboardData")
	procRegisterClipboardFormatW      = user32.NewProc("RegisterClipboardFormatW")
	procDragAcceptFiles               = shell32.NewProc("DragAcceptFiles")
	procDragQueryFileW                = shell32.NewProc("DragQueryFileW")
	procDragFinish                    = shell32.NewProc("DragFinish")

	procCreateSolidBrush = gdi32.NewProc("CreateSolidBrush")
	procCreatePen        = gdi32.NewProc("CreatePen")
	procSelectObject     = gdi32.NewProc("SelectObject")
	procDeleteObject     = gdi32.NewProc("DeleteObject")
	procRoundRect        = gdi32.NewProc("RoundRect")
	procSetTextColor     = gdi32.NewProc("SetTextColor")
	procSetBkMode        = gdi32.NewProc("SetBkMode")
	procCreateFontW      = gdi32.NewProc("CreateFontW")

	procGetOpenFileNameW = comdlg32.NewProc("GetOpenFileNameW")
	procShellExecuteW    = shell32.NewProc("ShellExecuteW")

	procGlobalAlloc      = kernel32.NewProc("GlobalAlloc")
	procGlobalLock       = kernel32.NewProc("GlobalLock")
	procGlobalUnlock     = kernel32.NewProc("GlobalUnlock")
	procGlobalSize       = kernel32.NewProc("GlobalSize")
	procGetModuleHandleW = kernel32.NewProc("GetModuleHandleW")
)

const (
	WM_DESTROY     = 0x0002
	WM_PAINT       = 0x000F
	WM_CLOSE       = 0x0010
	WM_SETCURSOR   = 0x0020
	WM_NCHITTEST   = 0x0084
	WM_MOUSEMOVE   = 0x0200
	WM_LBUTTONDOWN = 0x0201
	WM_KEYDOWN     = 0x0100
	WM_DROPFILES   = 0x0233
	WM_DPICHANGED  = 0x02E0

	HTCAPTION   = 2
	SW_SHOW     = 5
	SW_MINIMIZE = 6
	SW_RESTORE  = 9

	WS_POPUP          = 0x80000000
	WS_VISIBLE        = 0x10000000
	WS_MINIMIZEBOX    = 0x00020000
	WS_SYSMENU        = 0x00080000
	WS_EX_ACCEPTFILES = 0x00000010

	CS_HREDRAW    = 0x0002
	CS_VREDRAW    = 0x0001
	IDC_ARROW     = 32512
	IDC_HAND      = 32649
	DT_LEFT       = 0x0000
	DT_CENTER     = 0x0001
	DT_VCENTER    = 0x0004
	DT_SINGLELINE = 0x0020
	DT_WORDBREAK  = 0x0010
	TRANSPARENT   = 1
	PS_SOLID      = 0

	OFN_FILEMUSTEXIST = 0x00001000
	OFN_PATHMUSTEXIST = 0x00000800
	OFN_EXPLORER      = 0x00080000

	CF_UNICODETEXT  = 13
	GMEM_MOVEABLE   = 0x0002
	VK_CONTROL      = 0x11
	VK_ESCAPE       = 0x1B
	VK_RETURN       = 0x0D
	VK_LEFT         = 0x25
	VK_RIGHT        = 0x27
	VK_V            = 0x56
	KEYEVENTF_KEYUP = 0x0002

	SWP_NOMOVE     = 0x0002
	SWP_NOZORDER   = 0x0004
	SWP_NOACTIVATE = 0x0010
)

type POINT struct{ X, Y int32 }
type RECT struct{ Left, Top, Right, Bottom int32 }
type MSG struct {
	Hwnd           uintptr
	Message        uint32
	WParam, LParam uintptr
	Time           uint32
	Pt             POINT
}
type PAINTSTRUCT struct {
	Hdc                uintptr
	Erase              int32
	RcPaint            RECT
	Restore, IncUpdate int32
	Reserved           [32]byte
}
type WNDCLASSEX struct {
	Size                               uint32
	Style                              uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background uintptr
	MenuName, ClassName                *uint16
	IconSm                             uintptr
}
type OPENFILENAME struct {
	StructSize                 uint32
	Owner, Instance            uintptr
	Filter, CustomFilter       *uint16
	MaxCustFilter, FilterIndex uint32
	File                       *uint16
	MaxFile                    uint32
	FileTitle                  *uint16
	MaxFileTitle               uint32
	InitialDir, Title          *uint16
	Flags                      uint32
	FileOffset, FileExtension  uint16
	DefExt                     *uint16
	CustData                   uintptr
	Hook                       uintptr
	TemplateName               *uint16
	Reserved                   uintptr
	Reserved2                  uint32
	FlagsEx                    uint32
}
type INPUT struct {
	Type    uint32
	Ki      KEYBDINPUT
	Padding [8]byte
}
type KEYBDINPUT struct {
	Vk, Scan    uint16
	Flags, Time uint32
	ExtraInfo   uintptr
}

type appState struct {
	hwnd       uintptr
	dpi        int
	preview    *DocumentPreview
	sourcePath string
	status     string
	hover      int
	mu         sync.Mutex
}

var state appState
var wndProcCB = syscall.NewCallback(wndProc)
var handCursor uintptr
var arrowCursor uintptr

func rgb(r, g, b byte) uintptr { return uintptr(r) | uintptr(g)<<8 | uintptr(b)<<16 }
func px(v int) int32 {
	d := state.dpi
	if d <= 0 {
		d = 96
	}
	return int32(v * d / 96)
}

func utf16(s string) *uint16 { p, _ := syscall.UTF16PtrFromString(s); return p }
