package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"bricspayir/internal/ledger"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write([]byte(landingPageHTML))
	}
}

func (s *Server) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "bricspay-core-ledger",
	})
}

func (s *Server) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		accounts, err := ledger.ListAccounts(r.Context(), s.DB)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(accounts)
	case http.MethodPost:
		var req struct {
			Code     string `json:"code"`
			Type     string `json:"type"`
			Currency string `json:"currency"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		acc, err := ledger.CreateAccount(r.Context(), s.DB, req.Code, req.Type, req.Currency)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(acc)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp, err := ledger.RecordTransaction(r.Context(), s.DB, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

const landingPageHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>BRICS Pay | Decentralized Settlement Infrastructure</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #06090e;
            --surface: #0b1118;
            --surface-card: rgba(16, 24, 34, 0.7);
            --border: rgba(255, 255, 255, 0.08);
            --border-glow: rgba(0, 230, 153, 0.25);
            --text-main: #f0f4f8;
            --text-muted: #8a99a8;
            --emerald: #00e699;
            --emerald-dark: #008f5d;
            --emerald-glow: rgba(0, 230, 153, 0.12);
            --blue-glow: rgba(0, 153, 255, 0.12);
            --font: 'Plus Jakarta Sans', -apple-system, sans-serif;
            --font-mono: 'JetBrains Mono', monospace;
        }

        * { box-sizing: border-box; margin: 0; padding: 0; }
        body {
            background-color: var(--bg);
            color: var(--text-main);
            font-family: var(--font);
            line-height: 1.6;
            overflow-x: hidden;
            position: relative;
        }

        .ambient-glow-1 {
            position: absolute;
            top: -150px;
            left: 50%;
            transform: translateX(-50%);
            width: 800px;
            height: 500px;
            background: radial-gradient(circle, var(--emerald-glow) 0%, rgba(0,0,0,0) 70%);
            pointer-events: none;
            z-index: 0;
        }

        .ambient-glow-2 {
            position: absolute;
            top: 700px;
            right: -100px;
            width: 600px;
            height: 600px;
            background: radial-gradient(circle, var(--blue-glow) 0%, rgba(0,0,0,0) 70%);
            pointer-events: none;
            z-index: 0;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 0 24px;
            position: relative;
            z-index: 1;
        }

        header {
            border-bottom: 1px solid var(--border);
            backdrop-filter: blur(16px);
            position: sticky;
            top: 0;
            z-index: 100;
            background: rgba(6, 9, 14, 0.85);
        }
        .nav-wrap {
            display: flex;
            align-items: center;
            justify-content: space-between;
            height: 72px;
        }
        .brand {
            display: flex;
            align-items: center;
            gap: 12px;
            font-weight: 800;
            font-size: 1.2rem;
            letter-spacing: -0.02em;
            color: #fff;
            text-decoration: none;
        }
        .brand-badge {
            font-size: 0.7rem;
            font-weight: 600;
            background: rgba(0, 230, 153, 0.15);
            color: var(--emerald);
            border: 1px solid rgba(0, 230, 153, 0.3);
            padding: 3px 8px;
            border-radius: 20px;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
        .nav-links {
            display: flex;
            gap: 32px;
            align-items: center;
        }
        .nav-links a {
            color: var(--text-muted);
            text-decoration: none;
            font-size: 0.9rem;
            font-weight: 500;
            transition: color 0.2s ease;
        }
        .nav-links a:hover { color: #fff; }
        .btn-status {
            display: inline-flex;
            align-items: center;
            gap: 8px;
            background: rgba(255, 255, 255, 0.04);
            border: 1px solid var(--border);
            padding: 6px 14px;
            border-radius: 24px;
            font-size: 0.8rem;
            color: var(--text-muted);
        }
        .pulse-dot {
            width: 8px;
            height: 8px;
            background: var(--emerald);
            border-radius: 50%;
            box-shadow: 0 0 10px var(--emerald);
            animation: pulse 2s infinite;
        }
        @keyframes pulse {
            0% { transform: scale(0.95); opacity: 0.7; }
            50% { transform: scale(1.2); opacity: 1; }
            100% { transform: scale(0.95); opacity: 0.7; }
        }

        .hero-section {
            padding: 90px 0 60px;
            text-align: center;
        }
        .pill-badge {
            display: inline-flex;
            align-items: center;
            gap: 10px;
            background: rgba(0, 230, 153, 0.06);
            border: 1px solid rgba(0, 230, 153, 0.2);
            padding: 8px 18px;
            border-radius: 30px;
            font-size: 0.85rem;
            color: var(--emerald);
            margin-bottom: 28px;
            font-weight: 500;
        }
        h1.hero-title {
            font-size: clamp(2.4rem, 5vw, 4rem);
            font-weight: 800;
            line-height: 1.15;
            letter-spacing: -0.03em;
            margin-bottom: 20px;
            color: #fff;
        }
        .gradient-text {
            background: linear-gradient(135deg, #ffffff 40%, #00e699 100%);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        p.hero-desc {
            font-size: 1.15rem;
            color: var(--text-muted);
            max-width: 680px;
            margin: 0 auto 36px;
            font-weight: 400;
        }
        .hero-actions {
            display: flex;
            gap: 16px;
            justify-content: center;
            align-items: center;
            flex-wrap: wrap;
        }
        .btn-primary {
            background: linear-gradient(135deg, var(--emerald) 0%, var(--emerald-dark) 100%);
            color: #04100b;
            font-weight: 700;
            padding: 13px 26px;
            border-radius: 10px;
            text-decoration: none;
            box-shadow: 0 4px 20px rgba(0, 230, 153, 0.3);
            transition: transform 0.2s ease, box-shadow 0.2s ease;
        }
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 6px 28px rgba(0, 230, 153, 0.45);
        }
        .btn-secondary {
            background: rgba(255, 255, 255, 0.04);
            border: 1px solid var(--border);
            color: #fff;
            font-weight: 600;
            padding: 13px 26px;
            border-radius: 10px;
            text-decoration: none;
            transition: background 0.2s ease;
        }
        .btn-secondary:hover {
            background: rgba(255, 255, 255, 0.08);
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 20px;
            margin: 60px 0;
            padding: 24px;
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            backdrop-filter: blur(12px);
        }
        .stat-item { text-align: center; }
        .stat-value {
            font-size: 1.8rem;
            font-weight: 800;
            color: #fff;
            margin-bottom: 4px;
            font-family: var(--font-mono);
        }
        .stat-label { font-size: 0.8rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }

        .section-header {
            text-align: center;
            margin-bottom: 48px;
        }
        .section-title {
            font-size: 2.1rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 12px;
            color: #fff;
        }
        .section-subtitle {
            color: var(--text-muted);
            font-size: 1rem;
            max-width: 600px;
            margin: 0 auto;
        }

        .features-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 24px;
            margin-bottom: 90px;
        }
        .feature-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 30px;
            backdrop-filter: blur(12px);
            transition: all 0.3s ease;
        }
        .feature-card:hover {
            border-color: var(--border-glow);
            transform: translateY(-4px);
        }
        .feature-icon {
            width: 44px;
            height: 44px;
            background: rgba(0, 230, 153, 0.08);
            border: 1px solid rgba(0, 230, 153, 0.2);
            border-radius: 10px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-bottom: 18px;
            color: var(--emerald);
        }
        .feature-card h3 {
            font-size: 1.2rem;
            font-weight: 700;
            margin-bottom: 10px;
            color: #fff;
        }
        .feature-card p {
            color: var(--text-muted);
            font-size: 0.92rem;
            line-height: 1.6;
        }

        .workflow-section {
            background: var(--surface);
            border-top: 1px solid var(--border);
            border-bottom: 1px solid var(--border);
            padding: 80px 0;
            margin: 70px 0;
        }
        .steps-container {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
            gap: 18px;
            margin-top: 36px;
        }
        .step-card {
            background: rgba(255, 255, 255, 0.02);
            border: 1px solid var(--border);
            border-radius: 12px;
            padding: 22px;
        }
        .step-num {
            display: inline-block;
            font-family: var(--font-mono);
            font-size: 0.75rem;
            font-weight: 700;
            color: var(--emerald);
            background: rgba(0, 230, 153, 0.1);
            padding: 2px 8px;
            border-radius: 6px;
            margin-bottom: 10px;
        }
        .step-card h4 {
            font-size: 1rem;
            font-weight: 600;
            color: #fff;
            margin-bottom: 6px;
        }
        .step-card p {
            font-size: 0.85rem;
            color: var(--text-muted);
            line-height: 1.5;
        }

        .api-preview-card {
            background: #020508;
            border: 1px solid var(--border);
            border-radius: 16px;
            padding: 28px;
            margin: 50px 0 80px;
        }
        .api-header {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 16px;
            padding-bottom: 12px;
            border-bottom: 1px solid var(--border);
        }
        .terminal-dots { display: flex; gap: 6px; }
        .dot { width: 10px; height: 10px; border-radius: 50%; }
        .dot-red { background: #ff5f56; }
        .dot-yellow { background: #ffbd2e; }
        .dot-green { background: #27c93f; }
        pre.code-block {
            font-family: var(--font-mono);
            font-size: 0.85rem;
            color: #a0aec0;
            overflow-x: auto;
            line-height: 1.7;
        }
        .code-keyword { color: #f6789e; font-weight: 600; }
        .code-string { color: #85e89d; }

        footer {
            border-top: 1px solid var(--border);
            padding: 40px 0 30px;
            text-align: center;
        }
        .footer-text {
            color: var(--text-muted);
            font-size: 0.85rem;
            margin-bottom: 12px;
        }
        .footer-disclaimer {
            color: #556270;
            font-size: 0.75rem;
            max-width: 800px;
            margin: 0 auto;
            line-height: 1.5;
        }

        @media (max-width: 768px) {
            .stats-grid { grid-template-columns: 1fr 1fr; }
            .nav-links { display: none; }
        }
    </style>
</head>
<body>
    <div class="ambient-glow-1"></div>
    <div class="ambient-glow-2"></div>

    <header>
        <div class="container nav-wrap">
            <a href="/" class="brand">
                <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color:var(--emerald)"><polygon points="12 2 2 7 12 12 22 7 12 2"></polygon><polyline points="2 17 12 22 22 17"></polyline><polyline points="2 12 12 17 22 12"></polyline></svg>
                BRICS PAY <span class="brand-badge">CORE NETWORK</span>
            </a>
            <div class="nav-links">
                <a href="#services">Architecture</a>
                <a href="#workflow">Settlement Protocol</a>
                <a href="#api">API Reference</a>
                <div class="btn-status">
                    <div class="pulse-dot"></div>
                    <span>Ledger Online</span>
                </div>
            </div>
        </div>
    </header>

    <main>
        <section class="hero-section container">
            <div class="pill-badge">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                Next-Generation Cross-Border Settlement Infrastructure
            </div>
            <h1 class="hero-title">
                Move business forward,<br><span class="gradient-text">together across borders.</span>
            </h1>
            <p class="hero-desc">
                Decentralized financial messaging and multilateral atomic settlement framework designed to empower frictionless global commerce.
            </p>
            <div class="hero-actions">
                <a href="#api" class="btn-primary">Explore API Endpoints</a>
                <a href="#workflow" class="btn-secondary">Settlement Lifecycle</a>
            </div>

            <div class="stats-grid">
                <div class="stat-item">
                    <div class="stat-value">24 / 7</div>
                    <div class="stat-label">Atomic Settlement</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">ISO 20022</div>
                    <div class="stat-label">Messaging Standard</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">&lt; 3 sec</div>
                    <div class="stat-label">Execution Latency</div>
                </div>
                <div class="stat-item">
                    <div class="stat-value">Multi-CCY</div>
                    <div class="stat-label">National Currencies</div>
                </div>
            </div>
        </section>

        <section id="services" class="container">
            <div class="section-header">
                <h2 class="section-title">Institutional Core Modules</h2>
                <p class="section-subtitle">Engineered to complement existing multilateral payment infrastructures with verifiable compliance and real-time execution.</p>
            </div>

            <div class="features-grid">
                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="5" width="20" height="14" rx="2"/><line x1="2" y1="10" x2="22" y2="10"/></svg>
                    </div>
                    <h3>B2B Trade Settlement</h3>
                    <p>End-to-end clearing for international trade contracts with integrated digital tracking for multilateral trade instruments.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
                    </div>
                    <h3>DCMS Interbank Protocol</h3>
                    <p>Decentralized Cross-Border Messaging System providing sovereign, secure transaction instructions between member institutions.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
                    </div>
                    <h3>BRICS Settlement Unit</h3>
                    <p>Standardized accounting mechanism enabling direct national-currency pairs without reliance on third-party reserve intermediaries.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
                    </div>
                    <h3>Retail & Transfer Rails</h3>
                    <p>Unified QR and POS gateway standard facilitating cross-border business and commercial transactions seamlessly.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
                    </div>
                    <h3>Risk & Compliance Screening</h3>
                    <p>Automated, distributed multi-tier compliance layers adhering strictly to international multilateral AML/CFT frameworks.</p>
                </div>

                <div class="feature-card">
                    <div class="feature-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>
                    </div>
                    <h3>Immutable Core Ledger</h3>
                    <p>High-throughput double-entry atomic posting ledger ensuring non-repudiation and zero financial discrepancy across balances.</p>
                </div>
            </div>
        </section>

        <section id="workflow" class="workflow-section">
            <div class="container">
                <div class="section-header">
                    <h2 class="section-title">Structured Settlement Flow</h2>
                    <p class="section-subtitle">A multi-party verifiable lifecycle ensuring strict compliance checks prior to cross-border settlement.</p>
                </div>

                <div class="steps-container">
                    <div class="step-card">
                        <div class="step-num">STAGE 01</div>
                        <h4>Deal Initiation</h4>
                        <p>Buyer and seller agree on commercial trade terms and contract specifications.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 02</div>
                        <h4>Application Filing</h4>
                        <p>Buyer submits trade parameters and collateral for institutional credit review.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 03</div>
                        <h4>Instrument Issuance</h4>
                        <p>Issuing Bank issues Bank Guarantee (BG) or Bill of Exchange (BoE).</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 04</div>
                        <h4>Advisory Verification</h4>
                        <p>Corresponding banking nodes review and verify trade instruments under compliance.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 05</div>
                        <h4>Shipment & Notice</h4>
                        <p>Seller dispatches cargo and issues electronic shipping advice notice.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 06</div>
                        <h4>Document Presentation</h4>
                        <p>Shipping and trade documentation presented according to instrument terms.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 07</div>
                        <h4>Document Examination</h4>
                        <p>Automated discrepancy checks and multilateral trade compliance verification.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 08</div>
                        <h4>Ledger Settlement</h4>
                        <p>Atomic debit and multilateral clearing executed through the BRICS Pay network.</p>
                    </div>
                    <div class="step-card">
                        <div class="step-num">STAGE 09</div>
                        <h4>Document Release</h4>
                        <p>Final payment confirmation received and titles of ownership released.</p>
                    </div>
                </div>
            </div>
        </section>

        <section id="api" class="container">
            <div class="section-header">
                <h2 class="section-title">Operational Ledger Endpoints</h2>
                <p class="section-subtitle">Real-time developer integration ready for financial institutions.</p>
            </div>

            <div class="api-preview-card">
                <div class="api-header">
                    <div class="terminal-dots">
                        <div class="dot dot-red"></div>
                        <div class="dot dot-yellow"></div>
                        <div class="dot dot-green"></div>
                    </div>
                    <span style="font-family: var(--font-mono); font-size: 0.8rem; color: #777;">REST API Interface</span>
                </div>
                <pre class="code-block">
<span class="code-keyword">GET</span>  /health                         <span class="code-string">// Service health & DB ping</span>
<span class="code-keyword">GET</span>  /api/v1/accounts                <span class="code-string">// Query all participant clearing accounts</span>
<span class="code-keyword">POST</span> /api/v1/accounts                <span class="code-string">// Open institutional currency account</span>
<span class="code-keyword">POST</span> /api/v1/transactions            <span class="code-string">// Record atomic multi-leg posting</span>
                </pre>
            </div>
        </section>
    </main>

    <footer>
        <div class="container">
            <p class="footer-text">© 2026 BRICS Pay Consortium. All rights reserved.</p>
            <p class="footer-disclaimer">
                BRICS Pay operates as a complementary financial technology infrastructure adhering to international compliance frameworks. Service availability is subject to partner institution licensing and regulatory requirements.
            </p>
        </div>
    </footer>
</body>
</html>`
