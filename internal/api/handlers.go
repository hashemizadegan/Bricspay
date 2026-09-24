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
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #06090e;
            --surface: #0b1118;
            --surface-card: rgba(16, 24, 34, 0.75);
            --border: rgba(255, 255, 255, 0.08);
            --border-glow: rgba(0, 230, 153, 0.3);
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
            gap: 28px;
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
            padding: 80px 0 50px;
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
            margin-bottom: 24px;
            font-weight: 500;
        }
        h1.hero-title {
            font-size: clamp(2.4rem, 5vw, 3.8rem);
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
            margin: 0 auto 32px;
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
            padding: 12px 24px;
            border-radius: 10px;
            text-decoration: none;
            box-shadow: 0 4px 20px rgba(0, 230, 153, 0.3);
            cursor: pointer;
            border: none;
            display: inline-flex;
            align-items: center;
            gap: 8px;
            transition: all 0.2s ease;
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
            padding: 12px 24px;
            border-radius: 10px;
            text-decoration: none;
            cursor: pointer;
            transition: background 0.2s ease;
        }
        .btn-secondary:hover { background: rgba(255, 255, 255, 0.08); }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 20px;
            margin: 50px 0;
            padding: 24px;
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 16px;
            backdrop-filter: blur(12px);
        }
        .stat-item { text-align: center; }
        .stat-value {
            font-size: 1.7rem;
            font-weight: 800;
            color: #fff;
            margin-bottom: 4px;
            font-family: var(--font-mono);
        }
        .stat-label { font-size: 0.75rem; color: var(--text-muted); text-transform: uppercase; letter-spacing: 0.05em; }

        /* Dashboard Sandbox */
        .sandbox-section {
            background: var(--surface);
            border-top: 1px solid var(--border);
            border-bottom: 1px solid var(--border);
            padding: 60px 0;
            margin: 60px 0;
        }
        .sandbox-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 24px;
            margin-top: 30px;
        }
        .console-card {
            background: #090e15;
            border: 1px solid var(--border);
            border-radius: 14px;
            padding: 24px;
            display: flex;
            flex-direction: column;
            gap: 18px;
        }
        .console-title {
            font-size: 1.1rem;
            font-weight: 700;
            color: #fff;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .form-group {
            display: flex;
            flex-direction: column;
            gap: 6px;
        }
        .form-group label {
            font-size: 0.8rem;
            color: var(--text-muted);
            font-weight: 500;
        }
        .form-control {
            background: rgba(255, 255, 255, 0.03);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 10px 14px;
            color: #fff;
            font-family: var(--font-mono);
            font-size: 0.9rem;
            outline: none;
            transition: border-color 0.2s ease;
        }
        .form-control:focus {
            border-color: var(--emerald);
        }
        .form-row {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 12px;
        }
        .accounts-list-wrap {
            margin-top: 30px;
            background: #090e15;
            border: 1px solid var(--border);
            border-radius: 14px;
            padding: 24px;
        }
        .accounts-table {
            width: 100%;
            border-collapse: collapse;
            font-size: 0.85rem;
            text-align: left;
            margin-top: 14px;
        }
        .accounts-table th {
            padding: 10px;
            color: var(--text-muted);
            border-bottom: 1px solid var(--border);
            font-weight: 600;
            text-transform: uppercase;
            font-size: 0.75rem;
        }
        .accounts-table td {
            padding: 12px 10px;
            border-bottom: 1px solid rgba(255,255,255,0.04);
            font-family: var(--font-mono);
        }
        .badge-type {
            background: rgba(255, 255, 255, 0.08);
            padding: 3px 8px;
            border-radius: 4px;
            font-size: 0.75rem;
        }
        .result-box {
            font-family: var(--font-mono);
            font-size: 0.8rem;
            background: rgba(0, 0, 0, 0.5);
            border: 1px solid var(--border);
            border-radius: 8px;
            padding: 12px;
            max-height: 120px;
            overflow-y: auto;
            color: #a0aec0;
            white-space: pre-wrap;
            display: none;
        }

        .section-header {
            text-align: center;
            margin-bottom: 40px;
        }
        .section-title {
            font-size: 2rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 10px;
            color: #fff;
        }
        .section-subtitle {
            color: var(--text-muted);
            font-size: 0.95rem;
            max-width: 600px;
            margin: 0 auto;
        }

        .features-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 20px;
            margin-bottom: 60px;
        }
        .feature-card {
            background: var(--surface-card);
            border: 1px solid var(--border);
            border-radius: 14px;
            padding: 26px;
            backdrop-filter: blur(12px);
            transition: all 0.3s ease;
        }
        .feature-card:hover {
            border-color: var(--border-glow);
            transform: translateY(-3px);
        }
        .feature-icon {
            width: 40px;
            height: 40px;
            background: rgba(0, 230, 153, 0.08);
            border: 1px solid rgba(0, 230, 153, 0.2);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            margin-bottom: 16px;
            color: var(--emerald);
        }
        .feature-card h3 {
            font-size: 1.15rem;
            font-weight: 700;
            margin-bottom: 8px;
            color: #fff;
        }
        .feature-card p {
            color: var(--text-muted);
            font-size: 0.9rem;
        }

        footer {
            border-top: 1px solid var(--border);
            padding: 40px 0 30px;
            text-align: center;
        }
        .footer-text {
            color: var(--text-muted);
            font-size: 0.85rem;
            margin-bottom: 10px;
        }
        .footer-disclaimer {
            color: #556270;
            font-size: 0.75rem;
            max-width: 800px;
            margin: 0 auto;
        }

        @media (max-width: 768px) {
            .stats-grid, .sandbox-grid { grid-template-columns: 1fr; }
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
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="color:var(--emerald)"><polygon points="12 2 2 7 12 12 22 7 12 2"></polygon><polyline points="2 17 12 22 22 17"></polyline><polyline points="2 12 12 17 22 12"></polyline></svg>
                BRICS PAY <span class="brand-badge">CORE NETWORK</span>
            </a>
            <div class="nav-links">
                <a href="#sandbox">Pilot Sandbox</a>
                <a href="#services">Architecture</a>
                <div class="btn-status">
                    <div class="pulse-dot"></div>
                    <span>Ledger Engine Active</span>
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
                <a href="#sandbox" class="btn-primary">Launch Pilot Sandbox</a>
                <a href="#services" class="btn-secondary">Core Architecture</a>
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

        <!-- Interactive Sandbox Section -->
        <section id="sandbox" class="sandbox-section">
            <div class="container">
                <div class="section-header">
                    <h2 class="section-title">Institutional Pilot Sandbox</h2>
                    <p class="section-subtitle">Simulate real-time account creation and double-entry ledger postings on the live backend database.</p>
                </div>

                <div class="sandbox-grid">
                    <!-- Create Account Form -->
                    <div class="console-card">
                        <div class="console-title">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color:var(--emerald)"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="8.5" cy="7" r="4"></circle><line x1="20" y1="8" x2="20" y2="14"></line><line x1="23" y1="11" x2="17" y2="11"></line></svg>
                            Open Ledger Account
                        </div>
                        <form id="createAccountForm" onsubmit="handleCreateAccount(event)">
                            <div class="form-group" style="margin-bottom: 12px;">
                                <label>Account Code / Identifier</label>
                                <input type="text" id="accCode" class="form-control" placeholder="e.g. VTB-NOSTRO-RUB" required>
                            </div>
                            <div class="form-row" style="margin-bottom: 16px;">
                                <div class="form-group">
                                    <label>Account Type</label>
                                    <select id="accType" class="form-control">
                                        <option value="nostro">Nostro</option>
                                        <option value="vostro">Vostro</option>
                                        <option value="clearing">Clearing / Settlement</option>
                                    </select>
                                </div>
                                <div class="form-group">
                                    <label>Currency</label>
                                    <select id="accCcy" class="form-control">
                                        <option value="RUB">RUB (Russian Ruble)</option>
                                        <option value="CNY">CNY (Chinese Yuan)</option>
                                        <option value="AED">AED (Emirati Dirham)</option>
                                        <option value="IRR">IRR (Iranian Rial)</option>
                                        <option value="INR">INR (Indian Rupee)</option>
                                    </select>
                                </div>
                            </div>
                            <button type="submit" class="btn-primary" style="width:100%; justify-content:center;">Create Account</button>
                        </form>
                        <div id="createAccResult" class="result-box"></div>
                    </div>

                    <!-- Post Transaction Form -->
                    <div class="console-card">
                        <div class="console-title">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="color:var(--emerald)"><polyline points="17 1 21 5 17 9"></polyline><path d="M3 11V9a4 4 0 0 1 4-4h14"></path><polyline points="7 23 3 19 7 15"></polyline><path d="M21 13v2a4 4 0 0 1-4 4H3"></path></svg>
                            Execute Atomic Double-Entry Settlement
                        </div>
                        <form id="postTxForm" onsubmit="handlePostTx(event)">
                            <div class="form-group" style="margin-bottom: 12px;">
                                <label>Settlement Reference / Purpose</label>
                                <input type="text" id="txRef" class="form-control" placeholder="e.g. TR-2026-AGRI-SETTLE" required>
                            </div>
                            <div class="form-row" style="margin-bottom: 12px;">
                                <div class="form-group">
                                    <label>Debit Account Code</label>
                                    <input type="text" id="debitAcc" class="form-control" placeholder="e.g. VTB-NOSTRO-RUB" required>
                                </div>
                                <div class="form-group">
                                    <label>Credit Account Code</label>
                                    <input type="text" id="creditAcc" class="form-control" placeholder="e.g. BIM-VOSTRO-RUB" required>
                                </div>
                            </div>
                            <div class="form-row" style="margin-bottom: 16px;">
                                <div class="form-group">
                                    <label>Amount</label>
                                    <input type="number" id="txAmount" class="form-control" placeholder="100000.00" step="any" required>
                                </div>
                                <div class="form-group">
                                    <label>Currency</label>
                                    <select id="txCcy" class="form-control">
                                        <option value="RUB">RUB</option>
                                        <option value="CNY">CNY</option>
                                        <option value="AED">AED</option>
                                        <option value="IRR">IRR</option>
                                    </select>
                                </div>
                            </div>
                            <button type="submit" class="btn-primary" style="width:100%; justify-content:center;">Execute Settlement</button>
                        </form>
                        <div id="txResult" class="result-box"></div>
                    </div>
                </div>

                <!-- Live Accounts List -->
                <div class="accounts-list-wrap">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                        <h3 style="font-size:1.1rem; color:#fff;">Live Participant Accounts on Ledger</h3>
                        <button onclick="refreshAccounts()" class="btn-secondary" style="padding: 6px 14px; font-size:0.8rem;">↻ Refresh Accounts</button>
                    </div>
                    <table class="accounts-table">
                        <thead>
                            <tr>
                                <th>Account Code</th>
                                <th>Type</th>
                                <th>Currency</th>
                                <th>Status</th>
                            </tr>
                        </thead>
                        <tbody id="accountsBody">
                            <tr><td colspan="4" style="color:var(--text-muted); text-align:center;">Loading accounts from ledger...</td></tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </section>

        <!-- Services Section -->
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

    <script>
        async function refreshAccounts() {
            const body = document.getElementById('accountsBody');
            try {
                const res = await fetch('/api/v1/accounts');
                const data = await res.json();
                if (!data || data.length === 0) {
                    body.innerHTML = '<tr><td colspan="4" style="color:var(--text-muted); text-align:center; padding:18px;">No accounts found on ledger yet. Create one above.</td></tr>';
                    return;
                }
                body.innerHTML = data.map(acc => `
                    <tr>
                        <td style="color:#fff; font-weight:600;">${acc.code || acc.id}</td>
                        <td><span class="badge-type">${acc.type || 'Nostro'}</span></td>
                        <td style="color:var(--emerald); font-weight:600;">${acc.currency || 'RUB'}</td>
                        <td style="color:#27c93f;">● ACTIVE</td>
                    </tr>
                `).join('');
            } catch (err) {
                body.innerHTML = '<tr><td colspan="4" style="color:#ff5f56; text-align:center;">Failed to load accounts from server.</td></tr>';
            }
        }

        async function handleCreateAccount(e) {
            e.preventDefault();
            const resBox = document.getElementById('createAccResult');
            resBox.style.display = 'block';
            resBox.textContent = 'Sending request to ledger...';

            const payload = {
                code: document.getElementById('accCode').value,
                type: document.getElementById('accType').value,
                currency: document.getElementById('accCcy').value
            };

            try {
                const res = await fetch('/api/v1/accounts', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                const result = await res.json();
                resBox.textContent = 'HTTP ' + res.status + ':\n' + JSON.stringify(result, null, 2);
                if (res.ok) {
                    document.getElementById('accCode').value = '';
                    refreshAccounts();
                }
            } catch (err) {
                resBox.textContent = 'Error: ' + err.message;
            }
        }

        async function handlePostTx(e) {
            e.preventDefault();
            const resBox = document.getElementById('txResult');
            resBox.style.display = 'block';
            resBox.textContent = 'Executing atomic posting...';

            const amount = parseFloat(document.getElementById('txAmount').value);
            const ccy = document.getElementById('txCcy').value;
            const ref = document.getElementById('txRef').value;
            const debit = document.getElementById('debitAcc').value;
            const credit = document.getElementById('creditAcc').value;

            const payload = {
                reference: ref,
                postings: [
                    { account_code: debit, direction: "debit", amount: amount, currency: ccy },
                    { account_code: credit, direction: "credit", amount: amount, currency: ccy }
                ]
            };

            try {
                const res = await fetch('/api/v1/transactions', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                const result = await res.json();
                resBox.textContent = 'HTTP ' + res.status + ':\n' + JSON.stringify(result, null, 2);
                if (res.ok) {
                    refreshAccounts();
                }
            } catch (err) {
                resBox.textContent = 'Error: ' + err.message;
            }
        }

        // Initial fetch on page load
        document.addEventListener('DOMContentLoaded', refreshAccounts);
    </script>
</body>
</html>`
