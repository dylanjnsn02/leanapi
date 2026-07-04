# leanapi

A terminal-based HTTP API client (like Postman, but in your terminal) — built in Go with [Bubble Tea](https://github.com/charmbracelet/bubbletea). Mouse-clickable UI for building requests with query params, headers, cookies, auth (Basic/Bearer/API Key), and a body editor, plus pretty-printed/syntax-highlighted JSON responses and persistent request history.

![App image](https://raw.githubusercontent.com/dylanjnsn02/leanapi/refs/heads/main/images/Screenshot.png)

A leaner, no-TUI version here [leanapi_lite](https://github.com/dylanjnsn02/leanapi_lite)

## Features

- **Mouse-driven UI** — click the method pill, URL bar, Send button, tabs, form rows, and response toggles instead of memorizing keybindings.
- **Full request builder** — GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS, with dedicated tabs for:
  - **Params** — query params merged onto the URL at send time
  - **Body** — free-form request body editor (defaults to `application/json` when a body is present and no explicit `Content-Type` is set)
  - **Headers** — custom headers with enable/disable toggles; auth-derived headers show up as read-only rows so they never silently collide with your own
  - **Auth** — No Auth, Basic, Bearer Token, or API Key (sent via header or query param)
  - **Cookies** — name/value pairs sent as a single `Cookie` header
- **Response viewer** — status code, timing, and size at a glance; toggle between pretty-printed/highlighted JSON, response headers, and parsed cookies; scroll with the mouse wheel.
- **Copy to clipboard** — copy whichever response view (Body/Headers/Cookies) is active with one click or a keypress.
- **Persistent history** — every request/response is logged locally and can be browsed and reloaded back into the builder.

![App demo](https://raw.githubusercontent.com/dylanjnsn02/leanapi/refs/heads/main/images/gif.gif)

## Installation

### Quick install (macOS/Linux)

```bash
curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/dylanjnsn02/leanapi/main/install.sh | sh
```

This downloads the right binary for your OS/architecture from the [latest release](https://github.com/dylanjnsn02/leanapi/releases/latest) and installs it as `leanapi` in `$HOME/.local/bin` (no `sudo` required). If that directory isn't already on your `PATH`, the script will tell you what to add.

To install elsewhere, e.g. system-wide:

```bash
curl --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/dylanjnsn02/leanapi/main/install.sh | INSTALL_DIR=/usr/local/bin sh
```

Windows: grab `leanapi-windows-amd64.exe` directly from the [releases page](https://github.com/dylanjnsn02/leanapi/releases/latest).

### Build from source

Requires Go 1.25+.

```bash
git clone https://github.com/dylanjnsn02/leanapi.git
cd leanapi
go build -o leanapi ./cmd/leanapi
./leanapi
```

Or run it directly without building a binary:

```bash
go run ./cmd/leanapi
```

## Usage

### Mouse

- Click the **method pill** to cycle through HTTP methods.
- Click the **URL bar** to type a URL.
- Click **Send** to fire the request.
- Click a **tab** (Params / Body / Headers / Auth / Cookies) to switch it.
- Click a row's checkbox/key/value to toggle or edit it.
- Click **Body / Headers / Cookies** in the response bar to switch views, or **[Copy]** to copy the active view.
- Scroll the mouse wheel over the response viewer or body editor to scroll.

### Keyboard

| Key | Action |
| --- | --- |
| `Tab` / `Shift+Tab` | Move focus between fields |
| `Enter` | Send request (when URL or Send is focused) |
| `←` / `→` | Cycle method (when method pill focused) or switch tabs (when tab strip focused) |
| `Esc` | Return focus to the tab strip from inside a form pane |
| `y` | Copy the active response view (when response pane is focused) |
| `Ctrl+N` / `Ctrl+D` | Add / delete a row (Params, Headers, Cookies) |
| `Space` | Toggle a row's enabled checkbox |
| `Ctrl+H` | Open/close request history |
| `Ctrl+C` | Quit |

Inside the history view: `↑`/`↓` or `j`/`k` to select, `Enter` to reload a past request into the builder, `Esc` to close.

![History demo](https://raw.githubusercontent.com/dylanjnsn02/leanapi/refs/heads/main/images/gif2.gif)

## Project layout

```
cmd/leanapi/        entrypoint
internal/model/      Request/Header/Auth data types
internal/httpclient/  building and sending requests
internal/jsonview/    JSON pretty-printing and syntax highlighting
internal/history/     local JSONL request/response history
internal/ui/          Bubble Tea model, panes, and mouse hit-testing
```

## License

MIT — see [LICENSE](LICENSE).
