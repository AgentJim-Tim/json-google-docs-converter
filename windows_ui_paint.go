//go:build windows

package main

import (
    "fmt"
    "strings"
    "unsafe"
)

func paintPreview(hdc uintptr) {
	p := state.preview
	drawText(hdc, "Preview", 34, 69, 300, 42, 25, true, rgb(54, 51, 48), false)
	drawText(hdc, p.Title, 35, 111, 430, 28, 13, true, rgb(68, 64, 60), false)
	meta := fmt.Sprintf("%d page%s  ·  %d objects  ·  %d arrays  ·  depth %d", p.ApproxPages, plural(p.ApproxPages), p.Stats.Objects, p.Stats.Arrays, p.Stats.Depth)
	drawText(hdc, meta, 35, 141, 500, 22, 9, false, rgb(134, 128, 120), false)

	shadowRect := RECT{px(48), px(182), px(472), px(530)}
	fill(hdc, shadowRect, rgb(222, 219, 213))
	page := RECT{px(44), px(178), px(468), px(526)}
	fill(hdc, page, rgb(252, 251, 248))
	drawText(hdc, p.Title, 70, 201, 360, 35, 18, true, rgb(54, 51, 48), false)
	lines := previewLines(p.PlainText, 12)
	y := 247
	for i, line := range lines {
		size := 10
		bold := false
		if i > 0 && !strings.Contains(line, ":") && !strings.HasPrefix(line, "•") && len([]rune(line)) < 48 {
			size = 11
			bold = true
		}
		drawText(hdc, line, 70, y, 360, 30, size, bold, rgb(83, 79, 73), false)
		y += 22
	}

	drawSmallButton(hdc, 40, 565, 98, 42, "Change", state.hover == 6)
	drawSmallButton(hdc, 148, 565, 77, 42, "Back", state.hover == 7)
	rect := RECT{px(500), px(555), px(720), px(607)}
	fillRound(hdc, rect, 12, rgb(59, 57, 53), rgb(59, 57, 53))
	drawText(hdc, "Open in Google Docs", 511, 569, 198, 26, 11, true, rgb(250, 249, 246), true)
	drawText(hdc, "Enter to open", 500, 618, 220, 18, 9, false, rgb(132, 126, 119), true)
}

func previewLines(s string, max int) []string {
	raw := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	out := []string{}
	for _, l := range raw {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		if len([]rune(l)) > 60 {
			r := []rune(l)
			l = string(r[:57]) + "…"
		}
		out = append(out, l)
		if len(out) >= max {
			break
		}
	}
	return out
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func drawCard(hdc uintptr, x, y, w, h int, hover, primary bool) {
	bg := rgb(246, 244, 239)
	border := rgb(211, 206, 198)
	if hover {
		bg = rgb(249, 247, 243)
		border = rgb(188, 154, 137)
	}
	if primary && !hover {
		border = rgb(199, 184, 173)
	}
	fillRound(hdc, RECT{px(x), px(y), px(x + w), px(y + h)}, 14, bg, border)
}
func drawSmallButton(hdc uintptr, x, y, w, h int, label string, hover bool) {
	bg := rgb(241, 239, 234)
	border := rgb(207, 202, 194)
	if hover {
		bg = rgb(247, 245, 241)
		border = rgb(188, 154, 137)
	}
	fillRound(hdc, RECT{px(x), px(y), px(x + w), px(y + h)}, 10, bg, border)
	drawText(hdc, label, x, y+9, w, h-12, 10, true, rgb(81, 77, 72), true)
}
func fillRound(hdc uintptr, r RECT, radius int, bg, border uintptr) {
	brush, _, _ := procCreateSolidBrush.Call(bg)
	pen, _, _ := procCreatePen.Call(PS_SOLID, 1, border)
	oldB, _, _ := procSelectObject.Call(hdc, brush)
	oldP, _, _ := procSelectObject.Call(hdc, pen)
	procRoundRect.Call(hdc, uintptr(r.Left), uintptr(r.Top), uintptr(r.Right), uintptr(r.Bottom), uintptr(px(radius)), uintptr(px(radius)))
	procSelectObject.Call(hdc, oldB)
	procSelectObject.Call(hdc, oldP)
	procDeleteObject.Call(brush)
	procDeleteObject.Call(pen)
}
func fill(hdc uintptr, r RECT, color uintptr) {
	b, _, _ := procCreateSolidBrush.Call(color)
	procFillRect.Call(hdc, uintptr(unsafe.Pointer(&r)), b)
	procDeleteObject.Call(b)
}

func drawText(hdc uintptr, s string, x, y, w, h, size int, bold bool, color uintptr, center bool) {
	weight := int32(400)
	if bold {
		weight = 600
	}
	face := "Segoe UI"
	if size >= 16 || (bold && size >= 11) {
		face = "Georgia"
	}
	font, _, _ := procCreateFontW.Call(uintptr(-px(size)), 0, 0, 0, uintptr(weight), 0, 0, 0, 1, 0, 0, 5, 0, uintptr(unsafe.Pointer(utf16(face))))
	old, _, _ := procSelectObject.Call(hdc, font)
	procSetTextColor.Call(hdc, color)
	r := RECT{px(x), px(y), px(x + w), px(y + h)}
	flags := uintptr(DT_VCENTER | DT_SINGLELINE)
	if center {
		flags |= DT_CENTER
	} else {
		flags |= DT_LEFT
	}
	procDrawTextW.Call(hdc, uintptr(unsafe.Pointer(utf16(s))), uintptr(^uint32(0)), uintptr(unsafe.Pointer(&r)), flags)
	procSelectObject.Call(hdc, old)
	procDeleteObject.Call(font)
}
