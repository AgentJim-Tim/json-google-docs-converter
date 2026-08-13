package main

import (
	"strings"
	"testing"
)

func TestBuildPreviewNorthstar(t *testing.T) {
	data := []byte(`{"title":"Project Northstar","status":"In Progress","objectives":["Improve reliability","Reduce overhead"],"milestones":[{"phase":"Discovery","status":"Complete"},{"phase":"Migration","status":"In Progress"}]}`)
	p, err := buildPreview(data, "fallback.json")
	if err != nil {
		t.Fatal(err)
	}
	if p.Title != "Project Northstar" {
		t.Fatalf("title=%q", p.Title)
	}
	if !strings.Contains(p.HTML, "<table") {
		t.Fatalf("expected table HTML")
	}
	if !strings.Contains(p.PlainText, "Objectives") {
		t.Fatalf("missing objectives")
	}
	if p.ApproxPages < 1 {
		t.Fatalf("bad page count")
	}
}

func TestBOM(t *testing.T) {
	_, err := parseJSON(append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"ok":true}`)...))
	if err != nil {
		t.Fatal(err)
	}
}

func TestClipboardReverse(t *testing.T) {
	got := clipboardTextToJSON("Demo\nStatus: Complete\n• One\n• Two")
	if got["title"] != "Demo" {
		t.Fatalf("bad title")
	}
	if got["status"] != "Complete" {
		t.Fatalf("bad status")
	}
}
