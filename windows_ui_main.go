//go:build windows

package main

import "unsafe"

func main() {
	procSetProcessDpiAwarenessContext.Call(^uintptr(3))
	arrowCursor, _, _ = procLoadCursorW.Call(0, IDC_ARROW)
	handCursor, _, _ = procLoadCursorW.Call(0, IDC_HAND)

	hInst, _, _ := procGetModuleHandleW.Call(0)
	className := utf16("JSONGoogleDocsConverterV32")
	icon, _, _ := user32.NewProc("LoadIconW").Call(hInst, 1)
	wc := WNDCLASSEX{Size: uint32(unsafe.Sizeof(WNDCLASSEX{})), Style: CS_HREDRAW | CS_VREDRAW, WndProc: wndProcCB, Instance: hInst, Icon: icon, Cursor: arrowCursor, ClassName: className, IconSm: icon}
	if r, _, _ := procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc))); r == 0 {
		panic("RegisterClassExW failed")
	}

	state.dpi = 96
	w, h := px(456), px(374)
	hwnd, _, _ := procCreateWindowExW.Call(WS_EX_ACCEPTFILES, uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(utf16("JSON ↔ Google Docs Converter"))), WS_POPUP|WS_VISIBLE|WS_MINIMIZEBOX|WS_SYSMENU, 240, 160, uintptr(w), uintptr(h), 0, 0, hInst, 0)
	if hwnd == 0 {
		panic("CreateWindowExW failed")
	}
	state.hwnd = hwnd
	if r, _, _ := procGetDpiForWindow.Call(hwnd); r != 0 {
		state.dpi = int(r)
	}
	procDragAcceptFiles.Call(hwnd, 1)
	procShowWindow.Call(hwnd, SW_SHOW)
	procUpdateWindow.Call(hwnd)
	var msg MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&msg)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	case WM_CLOSE:
		user32.NewProc("DestroyWindow").Call(hwnd)
		return 0
	case WM_PAINT:
		paint(hwnd)
		return 0
	case WM_DPICHANGED:
		state.dpi = int(wParam & 0xFFFF)
		procInvalidateRect.Call(hwnd, 0, 1)
		return 0
	case WM_DROPFILES:
		handleDrop(wParam)
		return 0
	case WM_MOUSEMOVE:
		x, y := int16(lParam&0xffff), int16((lParam>>16)&0xffff)
		h := hitTest(int32(x), int32(y))
		if h != state.hover {
			state.hover = h
			procInvalidateRect.Call(hwnd, 0, 0)
		}
		if h != 0 {
			procSetCursor.Call(handCursor)
		} else {
			procSetCursor.Call(arrowCursor)
		}
		return 0
	case WM_SETCURSOR:
		if state.hover != 0 {
			procSetCursor.Call(handCursor)
			return 1
		}
	case WM_LBUTTONDOWN:
		x, y := int32(int16(lParam&0xffff)), int32(int16((lParam>>16)&0xffff))
		switch hitTest(x, y) {
		case 1:
			chooseJSON()
		case 2:
			reverseFromClipboard()
		case 3:
			procShowWindow.Call(hwnd, SW_MINIMIZE)
		case 4:
			user32.NewProc("DestroyWindow").Call(hwnd)
		case 5:
			openPreviewInDocs()
		case 6:
			chooseJSON()
		case 7:
			collapsePreview()
		default:
			if y < px(54) {
				procReleaseCapture.Call()
				procSendMessageW.Call(hwnd, 0x00A1, HTCAPTION, 0)
			}
		}
		return 0
	case WM_KEYDOWN:
		if wParam == VK_ESCAPE && state.preview != nil {
			collapsePreview()
			return 0
		}
		if wParam == VK_RETURN && state.preview != nil {
			openPreviewInDocs()
			return 0
		}
	}
	r, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return r
}

func hitTest(x, y int32) int {
	if y >= px(8) && y <= px(50) {
		if x >= px(360) && x < px(404) {
			return 3
		}
		if x >= px(404) && x <= px(452) {
			return 4
		}
	}
	if state.preview != nil {
		if x >= px(500) && x <= px(720) && y >= px(555) && y <= px(607) {
			return 5
		}
		if x >= px(40) && x <= px(138) && y >= px(565) && y <= px(607) {
			return 6
		}
		if x >= px(148) && x <= px(225) && y >= px(565) && y <= px(607) {
			return 7
		}
		return 0
	}
	if x >= px(34) && x <= px(422) && y >= px(144) && y <= px(224) {
		return 1
	}
	if x >= px(34) && x <= px(422) && y >= px(236) && y <= px(312) {
		return 2
	}
	return 0
}

func paint(hwnd uintptr) {
	var ps PAINTSTRUCT
	hdc, _, _ := procBeginPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	defer procEndPaint.Call(hwnd, uintptr(unsafe.Pointer(&ps)))
	var rc RECT
	procGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc)))
	fill(hdc, rc, rgb(239, 237, 232))
	procSetBkMode.Call(hdc, TRANSPARENT)

	drawText(hdc, "Converter", 26, 19, 230, 30, 12, false, rgb(89, 85, 79), false)
	drawText(hdc, "−", 367, 13, 36, 34, 19, false, rgb(91, 87, 81), true)
	drawText(hdc, "×", 411, 12, 36, 34, 22, false, rgb(91, 87, 81), true)

	if state.preview == nil {
		paintCompact(hdc)
	} else {
		paintPreview(hdc)
	}
}

func paintCompact(hdc uintptr) {
	drawText(hdc, "Convert", 34, 80, 388, 50, 27, true, rgb(54, 51, 48), false)
	drawText(hdc, "JSON and Google Docs", 35, 120, 350, 24, 10, false, rgb(132, 126, 118), false)
	drawCard(hdc, 34, 144, 388, 80, state.hover == 1, true)
	drawText(hdc, "Google Doc", 54, 160, 210, 28, 16, true, rgb(51, 49, 46), false)
	drawText(hdc, "from JSON", 54, 188, 220, 20, 10, false, rgb(126, 119, 111), false)
	drawText(hdc, "→", 372, 169, 30, 26, 16, false, rgb(184, 105, 77), true)
	drawCard(hdc, 34, 236, 388, 76, state.hover == 2, false)
	drawText(hdc, "JSON", 54, 251, 210, 28, 15, true, rgb(61, 58, 54), false)
	drawText(hdc, "from Google Doc", 54, 278, 220, 20, 10, false, rgb(135, 129, 122), false)
	drawText(hdc, "→", 372, 258, 30, 26, 15, false, rgb(165, 120, 99), true)
	status := state.status
	if status == "" {
		status = "Ready"
	}
	drawText(hdc, "●  "+status, 35, 336, 385, 20, 9, false, rgb(126, 121, 115), false)
}
