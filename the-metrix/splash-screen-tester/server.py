#!/usr/bin/env python3
"""
APOGEE Local Server
Serves the dashboard HTML files and persists survey/session data to JSON files.

Usage:
  python3 server.py

Then open: http://localhost:8765

Data is saved to:
  responses.json       — survey responses
  sessions.json        — OpenCode session data
"""

import json
import os
from http.server import HTTPServer, BaseHTTPRequestHandler
from pathlib import Path

PORT = 8765
DIR = Path(__file__).parent
RESPONSES_FILE = DIR / "responses.json"
SESSIONS_FILE  = DIR / "sessions.json"


def load_json(path):
    try:
        return json.loads(path.read_text()) if path.exists() else []
    except Exception:
        return []


def save_json(path, data):
    path.write_text(json.dumps(data, indent=2))


MIME = {
    ".html": "text/html",
    ".js":   "application/javascript",
    ".css":  "text/css",
    ".json": "application/json",
    ".ico":  "image/x-icon",
}


class Handler(BaseHTTPRequestHandler):

    def log_message(self, fmt, *args):
        # Custom compact logging
        print(f"  {self.command} {self.path}  →  {args[1] if len(args)>1 else ''}")

    def send_json(self, code, data):
        body = json.dumps(data).encode()
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", len(body))
        self.send_header("Access-Control-Allow-Origin", "*")
        self.end_headers()
        self.wfile.write(body)

    def send_cors(self):
        self.send_response(200)
        self.send_header("Access-Control-Allow-Origin", "*")
        self.send_header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        self.send_header("Access-Control-Allow-Headers", "Content-Type")
        self.end_headers()

    def do_OPTIONS(self):
        self.send_cors()

    def do_GET(self):
        path = self.path.split("?")[0]

        # API: load responses
        if path == "/responses":
            self.send_json(200, load_json(RESPONSES_FILE))
            return

        # API: load sessions
        if path == "/sessions":
            self.send_json(200, load_json(SESSIONS_FILE))
            return

        # Serve files
        if path == "/" or path == "":
            path = "/dora-dashboard.html"

        file_path = DIR / path.lstrip("/")
        if file_path.exists() and file_path.is_file():
            ext = file_path.suffix.lower()
            mime = MIME.get(ext, "text/plain")
            body = file_path.read_bytes()
            self.send_response(200)
            self.send_header("Content-Type", mime)
            self.send_header("Content-Length", len(body))
            self.send_header("Access-Control-Allow-Origin", "*")
            self.end_headers()
            self.wfile.write(body)
        else:
            self.send_json(404, {"error": f"Not found: {path}"})

    def do_POST(self):
        length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(length)

        try:
            data = json.loads(body)
        except Exception:
            self.send_json(400, {"error": "Invalid JSON"})
            return

        if self.path == "/save":
            save_json(RESPONSES_FILE, data)
            print(f"  ✓ Saved {len(data)} survey responses → {RESPONSES_FILE.name}")
            self.send_json(200, {"saved": len(data)})

        elif self.path == "/save-sessions":
            save_json(SESSIONS_FILE, data)
            print(f"  ✓ Saved {len(data)} OpenCode sessions → {SESSIONS_FILE.name}")
            self.send_json(200, {"saved": len(data)})

        elif self.path == "/feedback":
            # Append single feedback entry to feedback.json
            feedback_file = DIR / "feedback.json"
            all_fb = load_json(feedback_file)
            all_fb.append(data)
            save_json(feedback_file, all_fb)
            kind = data.get('kind', '?')
            comp = data.get('component', data.get('type', '?'))
            print(f"  ✓ Feedback [{kind}] {comp} → {feedback_file.name}")
            self.send_json(200, {"saved": True, "total": len(all_fb)})

        else:
            self.send_json(404, {"error": "Unknown endpoint"})


def main():
    print(f"""
╔══════════════════════════════════════════╗
║        APOGEE Local Server          ║
╠══════════════════════════════════════════╣
║  http://localhost:{PORT}                   ║
║                                          ║
║  Pages:                                  ║
║    /dora-dashboard.html  →  DORA         ║
║    /survey.html          →  Survey       ║
║    /opencode.html        →  AI Stats
║    /insights.html        →  Insights         ║
║                                          ║
║  API endpoints:                          ║
║    POST /save            →  responses    ║
║    POST /save-sessions   →  sessions     ║
║    GET  /responses       →  load survey  ║
║    GET  /sessions        →  load OC      ║
║                                          ║
║  Data files:                             ║
║    responses.json                        ║
║    sessions.json                         ║
║                                          ║
║  Ctrl+C to stop                          ║
╚══════════════════════════════════════════╝
""")
    server = HTTPServer(("", PORT), Handler)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\n  Server stopped.")


if __name__ == "__main__":
    main()
