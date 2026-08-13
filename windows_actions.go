//go:build windows

package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "syscall"
    "time"
    "unsafe"
)

func chooseJSON() {
	path, ok := openJSONDialog()
	if !ok {
		return
	}
	loadJSON(path)
}

func openJSONDialog() (string, bool) {
	buf := make([]uint16, 32768)
	filter := []uint16{'J', 'S', 'O', 'N', ' ', 'f', 'i', 'l', 'e', 's', 0, '*', '.', 'j', 's', 'o', 'n', 0, 'A', 'l', 'l', ' ', 'f', 'i', 'l', 'e', 's', 0, '*', '.', '*', 0, 0}
	ofn := OPENFILENAME{StructSize: uint32(unsafe.Sizeof(OPENFILENAME{})), Owner: state.hwnd, Filter: &filter[0], FilterIndex: 1, File: &buf[0], MaxFile: uint32(len(buf)), Title: utf16("Choose JSON file"), Flags: OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_EXPLORER, DefExt: utf16("json")}
	r, _, _ := procGetOpenFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", false
	}
	return syscall.UTF16ToString(buf), true
}

func loadJSON(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		setStatus("Could not read JSON")
		return
	}
	p, err := buildPreview(data, filepath.Base(path))
	if err != nil {
		setStatus("Invalid JSON: " + shortErr(err))
		return
	}
	state.preview = &p
	state.sourcePath = path
	state.status = "Preview ready"
	procSetWindowPos.Call(state.hwnd, 0, 0, 0, uintptr(px(760)), uintptr(px(650)), SWP_NOMOVE|SWP_NOZORDER|SWP_NOACTIVATE)
	procInvalidateRect.Call(state.hwnd, 0, 1)
}

func collapsePreview() {
	state.preview = nil
	state.sourcePath = ""
	state.status = "Ready"
	procSetWindowPos.Call(state.hwnd, 0, 0, 0, uintptr(px(456)), uintptr(px(374)), SWP_NOMOVE|SWP_NOZORDER|SWP_NOACTIVATE)
	procInvalidateRect.Call(state.hwnd, 0, 1)
}
func setStatus(s string) {
	state.mu.Lock()
	state.status = s
	state.mu.Unlock()
	if state.hwnd != 0 {
		procInvalidateRect.Call(state.hwnd, 0, 0)
	}
}
func shortErr(err error) string {
	s := err.Error()
	if len(s) > 80 {
		s = s[:77] + "…"
	}
	return s
}

func handleDrop(hdrop uintptr) {
	count, _, _ := procDragQueryFileW.Call(hdrop, 0xFFFFFFFF, 0, 0)
	if count > 0 {
		n, _, _ := procDragQueryFileW.Call(hdrop, 0, 0, 0)
		buf := make([]uint16, n+1)
		procDragQueryFileW.Call(hdrop, 0, uintptr(unsafe.Pointer(&buf[0])), n+1)
		path := syscall.UTF16ToString(buf)
		if strings.EqualFold(filepath.Ext(path), ".json") {
			loadJSON(path)
		} else {
			setStatus("Drop a .json file")
		}
	}
	procDragFinish.Call(hdrop)
}

func openPreviewInDocs() {
	if state.preview == nil {
		return
	}
	p := *state.preview
	if err := setRichClipboard(p.PlainText, p.HTML); err != nil {
		setStatus("Clipboard error: " + shortErr(err))
		return
	}
	setStatus("Opening Google Docs…")
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16("open"))), uintptr(unsafe.Pointer(utf16("https://docs.new"))), 0, 0, SW_SHOW)
	go func() {
		for i := 0; i < 12; i++ {
			time.Sleep(700 * time.Millisecond)
			hwnd, _, _ := procGetForegroundWindow.Call()
			if hwnd == 0 {
				continue
			}
			title := windowTitle(hwnd)
			low := strings.ToLower(title)
			if strings.Contains(low, "untitled document") || strings.Contains(low, "google docs") {
				if r, _, _ := procIsIconic.Call(hwnd); r != 0 {
					procShowWindowAsync.Call(hwnd, SW_RESTORE)
					time.Sleep(250 * time.Millisecond)
				}
				sendCtrlV()
				setStatus("Opened in Google Docs")
				return
			}
		}
		setStatus("Doc opened — press Ctrl+V if blank")
	}()
}

func windowTitle(hwnd uintptr) string {
	buf := make([]uint16, 512)
	n, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf[:n])
}
func sendCtrlV() {
	inputs := []INPUT{{Type: 1, Ki: KEYBDINPUT{Vk: VK_CONTROL}}, {Type: 1, Ki: KEYBDINPUT{Vk: VK_V}}, {Type: 1, Ki: KEYBDINPUT{Vk: VK_V, Flags: KEYEVENTF_KEYUP}}, {Type: 1, Ki: KEYBDINPUT{Vk: VK_CONTROL, Flags: KEYEVENTF_KEYUP}}}
	procSendInput.Call(uintptr(len(inputs)), uintptr(unsafe.Pointer(&inputs[0])), unsafe.Sizeof(INPUT{}))
}

func reverseFromClipboard() {
	text, err := getClipboardText()
	if err != nil || strings.TrimSpace(text) == "" {
		setStatus("Copy Google Doc content first")
		return
	}
	obj := clipboardTextToJSON(text)
	data, _ := json.MarshalIndent(obj, "", "  ")
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "Downloads")
	_ = os.MkdirAll(dir, 0755)
	title := "google-doc-export"
	if s, ok := obj["title"].(string); ok {
		title = sanitizeFilename(s)
	}
	path := uniquePath(filepath.Join(dir, title+".json"))
	if err := os.WriteFile(path, data, 0644); err != nil {
		setStatus("Could not save JSON")
		return
	}
	setStatus("JSON saved to Downloads")
	procShellExecuteW.Call(0, uintptr(unsafe.Pointer(utf16("open"))), uintptr(unsafe.Pointer(utf16(path))), 0, 0, SW_SHOW)
}

func sanitizeFilename(s string) string {
	bad := `<>:"/\\|?*`
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(bad, r) || r < 32 {
			return '-'
		}
		return r
	}, strings.TrimSpace(s))
	if len([]rune(s)) > 80 {
		s = string([]rune(s)[:80])
	}
	if s == "" {
		return "google-doc-export"
	}
	return s
}
func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 2; i < 1000; i++ {
		p := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return p
		}
	}
	return path
}

func setRichClipboard(plain, htmlDoc string) error {
	if r, _, _ := procOpenClipboard.Call(state.hwnd); r == 0 {
		return fmt.Errorf("OpenClipboard failed")
	}
	defer procCloseClipboard.Call()
	procEmptyClipboard.Call()
	if err := setClipboardUnicode(plain); err != nil {
		return err
	}
	formatName := utf16("HTML Format")
	cfHTML, _, _ := procRegisterClipboardFormatW.Call(uintptr(unsafe.Pointer(formatName)))
	payload := buildCFHTML(htmlDoc)
	if cfHTML != 0 {
		_ = setClipboardBytes(cfHTML, []byte(payload))
	}
	return nil
}
func setClipboardUnicode(s string) error {
	u, _ := syscall.UTF16FromString(s)
	size := uintptr(len(u) * 2)
	h, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if h == 0 {
		return fmt.Errorf("GlobalAlloc")
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return fmt.Errorf("GlobalLock")
	}
	dst := unsafe.Slice((*uint16)(unsafe.Pointer(p)), len(u))
	copy(dst, u)
	procGlobalUnlock.Call(h)
	if r, _, _ := procSetClipboardData.Call(CF_UNICODETEXT, h); r == 0 {
		return fmt.Errorf("SetClipboardData")
	}
	return nil
}
func setClipboardBytes(format uintptr, b []byte) error {
	b = append(b, 0)
	h, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, uintptr(len(b)))
	if h == 0 {
		return fmt.Errorf("GlobalAlloc")
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return fmt.Errorf("GlobalLock")
	}
	dst := unsafe.Slice((*byte)(unsafe.Pointer(p)), len(b))
	copy(dst, b)
	procGlobalUnlock.Call(h)
	if r, _, _ := procSetClipboardData.Call(format, h); r == 0 {
		return fmt.Errorf("SetClipboardData")
	}
	return nil
}
func buildCFHTML(fragment string) string {
	body := "<html><body><!--StartFragment-->" + fragment + "<!--EndFragment--></body></html>"
	headerTemplate := "Version:0.9\r\nStartHTML:%010d\r\nEndHTML:%010d\r\nStartFragment:%010d\r\nEndFragment:%010d\r\n"
	dummy := fmt.Sprintf(headerTemplate, 0, 0, 0, 0)
	startHTML := len(dummy)
	startFrag := startHTML + strings.Index(body, "<!--StartFragment-->") + len("<!--StartFragment-->")
	endFrag := startHTML + strings.Index(body, "<!--EndFragment-->")
	endHTML := startHTML + len(body)
	return fmt.Sprintf(headerTemplate, startHTML, endHTML, startFrag, endFrag) + body
}

func getClipboardText() (string, error) {
	if r, _, _ := procOpenClipboard.Call(state.hwnd); r == 0 {
		return "", fmt.Errorf("OpenClipboard")
	}
	defer procCloseClipboard.Call()
	h, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if h == 0 {
		return "", fmt.Errorf("clipboard has no text")
	}
	p, _, _ := procGlobalLock.Call(h)
	if p == 0 {
		return "", fmt.Errorf("GlobalLock")
	}
	defer procGlobalUnlock.Call(h)
	sz, _, _ := procGlobalSize.Call(h)
	max := int(sz / 2)
	u := unsafe.Slice((*uint16)(unsafe.Pointer(p)), max)
	n := 0
	for n < len(u) && u[n] != 0 {
		n++
	}
	return syscall.UTF16ToString(u[:n]), nil
}
