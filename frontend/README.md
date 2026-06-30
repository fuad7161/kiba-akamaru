# JobCirculer Frontend (HTMX)

> Modular HTMX-based frontend for the BD Govt Job Circular API.

## Architecture

```
frontend/
├── index.html                    # Shell (loads navbar + first page)
├── pages/                        # Full-page views (loaded via HTMX into #main-content)
│   ├── home.html                 # Hero + featured circulars + features + footer
│   ├── circulars.html            # Search, category pills, circular list, pagination
│   ├── circular-detail.html      # Single circular detail view
│   └── dashboard.html            # User dashboard (bookmarks + alerts)
├── components/                   # Reusable partials
│   ├── navbar.html               # Top navigation bar
│   └── auth-panel.html           # Login/Register slide panel
├── js/
│   ├── app.js                    # Auth state, render helpers, bookmark logic
│   └── data.js                   # Mock data (categories, demo circulars)
└── css/
    └── style.css                 # Complete stylesheet
```

## How it works

- **HTMX** loads partial HTML into `#main-content` on nav clicks — no SPA framework needed
- **Hyperscript** (`_="on click ..."`) handles inline interactions (bookmarks, pagination, auth panel)
- **Auth state** managed in `localStorage` (access_token + user object)
- **Mock data** in `data.js` until the backend API for circulars is ready

## Running

Serve with any HTTP server:

```bash
# Option A: Python
cd frontend && python3 -m http.server 4040

# Option B: Node (npx)
cd frontend && npx serve .

# Option C: Go file server
cd frontend && go run server.go
```

Then open `http://localhost:4040` in your browser.

> **Note:** The auth forms POST to `http://localhost:8080/api/v1/auth/...` — make sure the Go API backend is running.

## Key HTMX attributes used

| Attribute | Purpose |
|---|---|
| `hx-get` | Fetch a URL (GET request) |
| `hx-post` | Submit a form (POST request) |
| `hx-target` | Which element to replace with the response |
| `hx-swap` | How to swap (innerHTML, outerHTML) |
| `hx-push-url` | Update browser URL after request |
| `hx-trigger` | What triggers the request (load, click) |
| `hx-vals` | Extra JSON values sent with request |
| `hx-ext` | HTMX extensions (response-targets) |
