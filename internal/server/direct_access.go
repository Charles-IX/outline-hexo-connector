package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"outline-hexo-connector/internal/config"
	"strings"
)

const defaultDirectAccessHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Outline Hexo Connector</title>
  <style>
    :root { color-scheme: light; }
    body {
      margin: 0;
      font-family: "Helvetica Neue", Helvetica, Arial, sans-serif;
      color: #1f2937;
      background: radial-gradient(circle at top left, #f3f4f6, #e5e7eb);
      min-height: 100vh;
      display: flex;
      align-items: center;
      justify-content: center;
      padding: 24px;
      box-sizing: border-box;
    }
    .card {
      width: min(760px, 100%);
      background: rgba(255, 255, 255, 0.9);
      border: 1px solid #d1d5db;
      border-radius: 16px;
      box-shadow: 0 10px 30px rgba(31, 41, 55, 0.1);
      padding: 28px;
    }
    h1 { margin-top: 0; margin-bottom: 12px; font-size: 28px; }
    p { margin: 8px 0; line-height: 1.6; }
    code {
      background: #f3f4f6;
      border: 1px solid #e5e7eb;
      border-radius: 6px;
      padding: 1px 6px;
      font-size: 0.95em;
    }
  </style>
</head>
<body>
  <article class="card">
    <h1>Outline Hexo Connector is running</h1>
    <p>This endpoint is for Outline webhook delivery.</p>
    <p>Please send signed requests to <code>POST /webhook</code>.</p>
    <p>Direct browser access is enabled to avoid webhook verification errors.</p>
  </article>
</body>
</html>
`

type DirectAccessHandler struct {
	mode  string
	html  []byte
	proxy *httputil.ReverseProxy
}

func NewDirectAccessHandler(cfg *config.Config) (*DirectAccessHandler, error) {
	handler := &DirectAccessHandler{
		mode: "html",
		html: []byte(defaultDirectAccessHTML),
	}

	if cfg == nil {
		return handler, nil
	}

	mode := strings.TrimSpace(cfg.DirectAccessMode)
	if mode != "" {
		handler.mode = strings.ToLower(mode)
	}

	switch handler.mode {
	case "disabled":
		return handler, nil
	case "html":
		if strings.TrimSpace(cfg.DirectAccessHTMLFile) == "" {
			return handler, nil
		}
		content, err := os.ReadFile(cfg.DirectAccessHTMLFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read Direct_Access_HTML_File: %w", err)
		}
		handler.html = content
		return handler, nil
	case "proxy":
		targetRaw := strings.TrimSpace(cfg.DirectAccessProxyTarget)
		if targetRaw == "" {
			return nil, fmt.Errorf("Direct_Access_Proxy_Target is required when Direct_Access_Mode is proxy")
		}
		target, err := url.Parse(targetRaw)
		if err != nil {
			return nil, fmt.Errorf("invalid Direct_Access_Proxy_Target: %w", err)
		}
		if target.Scheme != "http" && target.Scheme != "https" {
			return nil, fmt.Errorf("Direct_Access_Proxy_Target must start with http:// or https://")
		}
		handler.proxy = httputil.NewSingleHostReverseProxy(target)
		return handler, nil
	default:
		return nil, fmt.Errorf("unsupported Direct_Access_Mode: %s", handler.mode)
	}
}

func (h *DirectAccessHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch h.mode {
	case "disabled":
		http.NotFound(w, r)
		return
	case "proxy":
		if h.proxy == nil {
			http.Error(w, "proxy not configured", http.StatusInternalServerError)
			return
		}
		h.proxy.ServeHTTP(w, r)
		return
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if r.Method != http.MethodHead {
			_, _ = w.Write(h.html)
		}
	}
}
