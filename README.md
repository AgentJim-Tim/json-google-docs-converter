# JSON ↔ Google Docs Converter

A compact Windows utility for converting **JSON into a formatted Google Doc** and turning copied **Google Doc content back into JSON**.

JSON → Google Doc is the primary workflow. The app validates your JSON, expands into a document-style preview, formats the data into readable sections/lists/tables, places rich content on the clipboard, and opens a fresh Google Doc in your default browser.

## Features

- **JSON → Google Doc** primary workflow
- **Google Doc → JSON** secondary workflow
- Compact Claude-inspired frameless Windows UI
- Preview-first conversion flow
- Detected document title
- Approximate page count and structure summary
- Natural ordering for common report fields
- Long text → narrative sections
- Scalar arrays → bullet lists
- Compact record arrays → tables
- Nested objects → readable sections
- UTF-8 BOM, Unicode, and emoji support
- Drag-and-drop `.json` files
- High-DPI / Per-Monitor-v2 rendering
- Native Windows hand cursor on interactive controls
- Embedded multi-resolution application icon
- Browser-window safeguard: maximized Chrome/Edge windows are not un-maximized
- Clipboard fallback if the browser blocks the single paste attempt

## JSON → Google Doc

1. Run `JSON-Google-Docs-Converter.exe`.
2. Click **Google Doc** or drop a `.json` file onto the app.
3. Review the expanded preview, title, structure counts, and approximate page count.
4. Click **Open in Google Docs** or press **Enter**.
5. A fresh Google Doc opens in your default browser.

The converter places both rich HTML and plain text on the Windows clipboard and makes one restrained paste attempt after the new Google Docs editor becomes the foreground window.

If your browser blocks the paste, the formatted document remains on the clipboard — press **Ctrl+V** once in the open Google Doc.

## Google Doc → JSON

1. Open the Google Doc.
2. Copy its content.
3. Return to the converter.
4. Click **JSON**.
5. A best-effort JSON export is written to your Downloads folder and opened with the default handler.

Because a rich document does not preserve every original JSON type/relationship, arbitrary Google Doc → JSON conversion is necessarily best-effort.

## Preview

Loading a JSON file expands the compact launcher into a review state showing the detected document title, page-style first-page preview, approximate page count, object/array/nesting counts, and the final **Open in Google Docs** action.

Press **Esc** to return to the compact view.

## Build from source

Requires **Go 1.23+** on Windows.

```powershell
go test ./...
./build.ps1
```

The build script embeds `assets/app-icon.ico` and creates `dist/JSON-Google-Docs-Converter.exe`.

## GitHub Actions

`.github/workflows/build.yml` tests and builds the Windows x64 executable on every push and pull request. The compiled EXE is uploaded as a workflow artifact.

## Privacy

JSON parsing and document formatting happen locally. The converter does not send your source JSON to a separate conversion server.

## Limitations

- Browser/Windows security may block the automatic paste, in which case the formatted content remains on the clipboard for manual `Ctrl+V`.
- Google Doc → JSON is best-effort because rich formatting does not map perfectly back to JSON.
- Current desktop build targets Windows x64.

## License

MIT — see [LICENSE](LICENSE).
