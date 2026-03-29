package dashboard

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%236366f1'/><text y='22' x='5' font-size='18' font-family='monospace' fill='white'>&#9670;</text></svg>">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
<script src="https://cdnjs.cloudflare.com/ajax/libs/Chart.js/4.4.1/chart.umd.min.js"></script>
<style>
:root {
  --bg:#0c0c14;--surface:#111120;--surface2:#181828;--surface3:#1f1f35;
  --border:#232338;--border2:#2a2a45;
  --accent:#6366f1;--accent2:#818cf8;--accent-dim:rgba(99,102,241,.12);--accent-glow:rgba(99,102,241,.25);
  --green:#10b981;--green-dim:rgba(16,185,129,.12);
  --red:#ef4444;--red-dim:rgba(239,68,68,.12);
  --yellow:#f59e0b;--yellow-dim:rgba(245,158,11,.12);
  --blue:#3b82f6;--blue-dim:rgba(59,130,246,.12);
  --text:#e2e2f0;--text2:#9090b8;--muted:#4a4a6a;
  --radius:10px;--radius-sm:6px;
}
*{box-sizing:border-box;margin:0;padding:0}
html,body{height:100%;background:var(--bg);color:var(--text);font-family:'Inter',sans-serif;font-size:14px;overflow:hidden}

/* ── Header ── */
.hdr{height:56px;display:flex;align-items:center;padding:0 20px;gap:16px;background:var(--surface);border-bottom:1px solid var(--border);flex-shrink:0;position:relative;z-index:10}
.logo{font-family:'JetBrains Mono',monospace;font-weight:600;font-size:16px;color:var(--accent);letter-spacing:-1px;display:flex;align-items:center;gap:6px}
.logo-diamond{color:var(--accent2)}
.hdr-spacer{flex:1}
.ws-badge{display:flex;align-items:center;gap:6px;font-size:11px;color:var(--text2);background:var(--surface2);border:1px solid var(--border);border-radius:20px;padding:4px 12px;font-family:'JetBrains Mono',monospace}
.ws-dot{width:7px;height:7px;border-radius:50%;background:var(--muted);transition:background .3s,box-shadow .3s}
.ws-dot.connected{background:var(--green);box-shadow:0 0 8px var(--green)}
.ws-dot.connecting{background:var(--yellow);animation:pulse 1.5s ease-in-out infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
.btn-logout{background:none;border:1px solid var(--border);color:var(--text2);border-radius:var(--radius-sm);padding:5px 12px;cursor:pointer;font-size:12px;font-family:'Inter',sans-serif;transition:all .15s}
.btn-logout:hover{border-color:var(--red);color:var(--red)}

/* ── Layout ── */
.layout{display:flex;height:calc(100vh - 56px)}

/* ── Sidebar ── */
.sidebar{width:212px;background:var(--surface);border-right:1px solid var(--border);padding:12px 8px;display:flex;flex-direction:column;gap:1px;flex-shrink:0;overflow-y:auto}
.nav{display:flex;align-items:center;gap:10px;padding:9px 12px;border-radius:var(--radius-sm);cursor:pointer;color:var(--text2);transition:all .12s;font-size:13px;user-select:none;position:relative}
.nav:hover{background:var(--surface2);color:var(--text)}
.nav.active{background:var(--accent-dim);color:var(--accent);font-weight:500}
.nav.active::before{content:'';position:absolute;left:0;top:50%;transform:translateY(-50%);width:3px;height:60%;background:var(--accent);border-radius:0 3px 3px 0}
.nav-icon{font-size:15px;width:20px;text-align:center;flex-shrink:0}
.sidebar-sep{height:1px;background:var(--border);margin:6px 4px}
.sidebar-label{font-size:10px;font-weight:600;color:var(--muted);text-transform:uppercase;letter-spacing:.1em;padding:8px 12px 4px}

/* ── Content ── */
.content{flex:1;overflow-y:auto;padding:28px;display:flex;flex-direction:column;gap:24px;scroll-behavior:smooth}
.view{display:none;flex-direction:column;gap:24px;animation:fadeIn .2s ease}
.view.active{display:flex}
@keyframes fadeIn{from{opacity:0;transform:translateY(4px)}to{opacity:1;transform:translateY(0)}}

/* ── Page titles ── */
.page-header{display:flex;align-items:center;justify-content:space-between;gap:16px}
.page-title{font-size:20px;font-weight:600;color:var(--text)}
.page-subtitle{font-size:13px;color:var(--text2);margin-top:2px}

/* ── Stat cards ── */
.stats-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}
.stat-card{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:18px 20px;transition:border-color .2s}
.stat-card:hover{border-color:var(--border2)}
.stat-card.accent{border-color:var(--accent-dim);background:linear-gradient(135deg,var(--surface) 0%,rgba(99,102,241,.05) 100%)}
.stat-val{font-family:'JetBrains Mono',monospace;font-size:30px;font-weight:600;line-height:1;color:var(--text);margin-bottom:6px}
.stat-card.accent .stat-val{color:var(--accent)}
.stat-label{font-size:11px;color:var(--text2);text-transform:uppercase;letter-spacing:.07em}
.stat-sub{font-size:11px;color:var(--muted);margin-top:3px}

/* ── Panels ── */
.panel{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);overflow:hidden}
.panel-head{display:flex;align-items:center;gap:8px;padding:14px 18px;border-bottom:1px solid var(--border);font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:.08em;color:var(--text2)}
.panel-head-right{margin-left:auto;display:flex;align-items:center;gap:8px}
.panel-body{padding:18px}

/* ── Live feed ── */
.feed{font-family:'JetBrains Mono',monospace;font-size:12px;overflow-y:auto;max-height:420px}
.feed-empty{padding:48px 24px;text-align:center;color:var(--muted);font-family:'Inter',sans-serif}
.feed-row{display:grid;grid-template-columns:100px 52px minmax(0,1.4fr) minmax(0,1fr) 50px 68px;gap:8px;padding:7px 10px;align-items:center;border-radius:var(--radius-sm);transition:background .08s}
.feed-row:hover{background:rgba(255,255,255,.03)}
.feed-row+.feed-row{border-top:1px solid rgba(255,255,255,.03)}
.ts{color:var(--muted)}
.meth{font-weight:600;color:var(--blue)}
.hcol{color:var(--text);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.pcol{color:var(--text2);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.sc{display:inline-block;padding:2px 7px;border-radius:4px;font-size:11px;font-weight:700}
.sc-2{background:var(--green-dim);color:var(--green)}
.sc-3{background:var(--blue-dim);color:var(--blue)}
.sc-4{background:var(--yellow-dim);color:var(--yellow)}
.sc-5{background:var(--red-dim);color:var(--red)}
.lat{color:var(--text2);text-align:right}
.lat.slow{color:var(--yellow)}
.lat.vslow{color:var(--red)}

/* ── Filter bar ── */
.filter-bar{display:flex;gap:8px;align-items:center;flex-wrap:wrap}
.filter-btn{background:var(--surface2);border:1px solid var(--border);color:var(--text2);border-radius:var(--radius-sm);padding:5px 12px;cursor:pointer;font-size:12px;transition:all .12s;font-family:'Inter',sans-serif}
.filter-btn:hover,.filter-btn.active{background:var(--accent-dim);border-color:var(--accent);color:var(--accent)}
.filter-input{background:var(--surface2);border:1px solid var(--border);color:var(--text);border-radius:var(--radius-sm);padding:5px 12px;font-size:12px;outline:none;font-family:'JetBrains Mono',monospace;width:180px;transition:border-color .15s}
.filter-input:focus{border-color:var(--accent)}
.filter-input::placeholder{color:var(--muted)}
.spacer{flex:1}
.pause-btn{display:flex;align-items:center;gap:6px;background:var(--surface2);border:1px solid var(--border);color:var(--text2);border-radius:var(--radius-sm);padding:5px 12px;cursor:pointer;font-size:12px;transition:all .12s;font-family:'Inter',sans-serif}
.pause-btn.paused{background:var(--yellow-dim);border-color:var(--yellow);color:var(--yellow)}
.pause-btn:hover{border-color:var(--text2);color:var(--text)}

/* ── Routes table ── */
.rtable-wrap{overflow-x:auto}
.rtable{width:100%;border-collapse:collapse;font-size:13px}
.rtable th{text-align:left;font-size:10px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);padding:9px 14px;border-bottom:1px solid var(--border);font-weight:600}
.rtable td{padding:12px 14px;border-bottom:1px solid rgba(255,255,255,.04)}
.rtable tr:last-child td{border-bottom:none}
.rtable tr:hover td{background:rgba(255,255,255,.02)}
.mono{font-family:'JetBrains Mono',monospace;font-size:12px}
.pill{display:inline-flex;align-items:center;gap:4px;padding:3px 9px;border-radius:20px;font-size:11px;font-weight:600}
.pill-green{background:var(--green-dim);color:var(--green)}
.pill-red{background:var(--red-dim);color:var(--red)}
.pill-blue{background:var(--blue-dim);color:var(--blue)}
.pill-yellow{background:var(--yellow-dim);color:var(--yellow)}
.pill-muted{background:rgba(74,74,106,.2);color:var(--muted)}

/* ── Actions ── */
.icon-btn{background:none;border:none;cursor:pointer;padding:5px 7px;border-radius:var(--radius-sm);color:var(--text2);transition:all .12s;font-size:13px;line-height:1}
.icon-btn:hover{background:var(--surface2);color:var(--text)}
.icon-btn.del:hover{color:var(--red);background:var(--red-dim)}
.actions{display:flex;gap:2px;align-items:center}

/* ── Add route form ── */
.add-form{display:flex;gap:10px;align-items:flex-end;flex-wrap:wrap}
.field{display:flex;flex-direction:column;gap:6px;flex:1;min-width:150px}
.field label{font-size:11px;color:var(--text2);text-transform:uppercase;letter-spacing:.06em;font-weight:500}
.field input{background:var(--bg);border:1px solid var(--border);border-radius:var(--radius-sm);padding:9px 12px;color:var(--text);font-family:'JetBrains Mono',monospace;font-size:12px;outline:none;transition:border-color .15s;width:100%}
.field input:focus{border-color:var(--accent)}
.field input::placeholder{color:var(--muted)}
.checkbox-field{display:flex;align-items:center;gap:8px;cursor:pointer;padding-bottom:2px}
.checkbox-field input[type=checkbox]{accent-color:var(--accent);width:14px;height:14px;cursor:pointer}
.checkbox-field span{font-size:13px;color:var(--text2)}

/* ── Buttons ── */
.btn{padding:9px 18px;border-radius:var(--radius-sm);border:none;font-family:'Inter',sans-serif;font-size:13px;font-weight:500;cursor:pointer;transition:all .15s;display:inline-flex;align-items:center;gap:6px}
.btn-primary{background:var(--accent);color:#fff}
.btn-primary:hover{background:var(--accent2)}
.btn-primary:disabled{opacity:.4;cursor:not-allowed}
.btn-ghost{background:var(--surface2);border:1px solid var(--border);color:var(--text2)}
.btn-ghost:hover{border-color:var(--border2);color:var(--text)}
.btn-danger{background:var(--red-dim);border:1px solid var(--red);color:var(--red)}
.btn-danger:hover{background:var(--red);color:#fff}
.btn-sm{padding:6px 12px;font-size:12px}

/* ── Cert cards ── */
.cert-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(280px,1fr));gap:12px}
.cert-card{background:var(--surface2);border:1px solid var(--border);border-radius:var(--radius);padding:16px 18px;position:relative;overflow:hidden}
.cert-card::before{content:'';position:absolute;top:0;left:0;right:0;height:3px}
.cert-card.valid::before{background:var(--green)}
.cert-card.expiring_soon::before{background:var(--yellow)}
.cert-card.error,.cert-card.pending{border-color:var(--border2)}
.cert-card.error::before{background:var(--red)}
.cert-card.pending::before{background:var(--muted)}
.cert-host{font-family:'JetBrains Mono',monospace;font-size:13px;font-weight:600;margin-bottom:10px;color:var(--text);overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.cert-meta{display:flex;flex-direction:column;gap:5px}
.cert-row{display:flex;justify-content:space-between;align-items:center;font-size:12px}
.cert-row .label{color:var(--text2)}
.cert-row .val{font-family:'JetBrains Mono',monospace;color:var(--text);font-size:11px}

/* ── Stats view ── */
.charts-grid{display:grid;grid-template-columns:1fr 1fr;gap:16px}
.chart-wrap{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:18px}
.chart-title{font-size:12px;font-weight:600;color:var(--text2);text-transform:uppercase;letter-spacing:.07em;margin-bottom:16px}
.chart-canvas{max-height:200px}
.route-stats-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:12px}
.rs-card{background:var(--surface2);border:1px solid var(--border);border-radius:var(--radius);padding:15px 17px}
.rs-host{font-family:'JetBrains Mono',monospace;font-size:12px;font-weight:600;margin-bottom:10px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap}
.rs-metrics{display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px;margin-bottom:10px}
.rs-val{font-family:'JetBrains Mono',monospace;font-size:15px;font-weight:600}
.rs-lbl{font-size:10px;color:var(--muted);margin-top:2px;text-transform:uppercase}
.rs-bar{height:3px;background:var(--border);border-radius:2px}
.rs-fill{height:100%;border-radius:2px;background:var(--accent);max-width:100%;transition:width .5s ease}
.rs-fill.has-errors{background:var(--red)}

/* ── Settings ── */
.settings-section{background:var(--surface);border:1px solid var(--border);border-radius:var(--radius);padding:24px;max-width:480px}
.settings-title{font-size:15px;font-weight:600;margin-bottom:4px}
.settings-desc{font-size:13px;color:var(--text2);margin-bottom:20px}
.form-stack{display:flex;flex-direction:column;gap:14px}

/* ── About ── */
.about-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(200px,1fr));gap:12px}
.about-card{background:var(--surface2);border:1px solid var(--border);border-radius:var(--radius);padding:16px 18px}
.about-card-label{font-size:10px;color:var(--muted);text-transform:uppercase;letter-spacing:.08em;margin-bottom:6px}
.about-card-val{font-family:'JetBrains Mono',monospace;font-size:14px;font-weight:600;color:var(--text)}
.about-links{display:flex;gap:10px;flex-wrap:wrap}
.about-link{color:var(--accent);font-size:13px;text-decoration:none;display:inline-flex;align-items:center;gap:4px}
.about-link:hover{color:var(--accent2);text-decoration:underline}

/* ── Alert banner ── */
.alert-bar{display:none;align-items:center;gap:10px;padding:10px 18px;background:var(--red-dim);border:1px solid var(--red);border-radius:var(--radius-sm);font-size:13px;color:var(--red)}
.alert-bar.show{display:flex}

/* ── Toasts ── */
#toasts{position:fixed;bottom:24px;right:24px;display:flex;flex-direction:column;gap:8px;z-index:1000;pointer-events:none}
.toast{display:flex;align-items:center;gap:10px;background:var(--surface2);border:1px solid var(--border);border-radius:var(--radius-sm);padding:12px 16px;font-size:13px;color:var(--text);box-shadow:0 8px 24px rgba(0,0,0,.4);animation:slideIn .2s ease;pointer-events:all;min-width:240px}
.toast.ok{border-left:3px solid var(--green)}
.toast.err{border-left:3px solid var(--red)}
.toast.info{border-left:3px solid var(--accent)}
@keyframes slideIn{from{opacity:0;transform:translateX(24px)}to{opacity:1;transform:translateX(0)}}
@keyframes slideOut{to{opacity:0;transform:translateX(24px)}}

/* ── Toggle switch ── */
.toggle{position:relative;display:inline-block;width:34px;height:18px;flex-shrink:0}
.toggle input{display:none}
.toggle-slider{position:absolute;inset:0;background:var(--border2);border-radius:20px;cursor:pointer;transition:background .2s}
.toggle-slider::before{content:'';position:absolute;width:12px;height:12px;left:3px;top:3px;background:#fff;border-radius:50%;transition:transform .2s}
.toggle input:checked+.toggle-slider{background:var(--accent)}
.toggle input:checked+.toggle-slider::before{transform:translateX(16px)}

/* ── Responsive ── */
@media(max-width:900px){.stats-grid{grid-template-columns:repeat(2,1fr)}.charts-grid{grid-template-columns:1fr}.sidebar{width:48px}.nav span:not(.nav-icon){display:none}.nav{padding:10px;justify-content:center}}
@media(max-width:600px){.stats-grid{grid-template-columns:1fr}.feed-row{grid-template-columns:80px 44px 1fr 44px}.hcol,.pcol{display:none}}
</style>
</head>
<body>

<!-- Header -->
<header class="hdr">
  <div class="logo"><span class="logo-diamond">&#9670;</span> onyx</div>
  <div class="hdr-spacer"></div>
  <div class="ws-badge"><div class="ws-dot connecting" id="wsDot"></div><span id="wsLabel">Connecting</span></div>
  <button class="btn-logout" onclick="location.href='/logout'">Sign out</button>
</header>

<div class="layout">
  <!-- Sidebar -->
  <nav class="sidebar">
    <div class="nav active" onclick="showView('overview',this)" id="nav-overview">
      <span class="nav-icon">&#9632;</span><span>Overview</span>
    </div>
    <div class="nav" onclick="showView('traffic',this)" id="nav-traffic">
      <span class="nav-icon">&#9889;</span><span>Traffic</span>
    </div>
    <div class="sidebar-sep"></div>
    <div class="nav" onclick="showView('routes',this)" id="nav-routes">
      <span class="nav-icon">&#8644;</span><span>Routes</span>
    </div>
    <div class="nav" onclick="showView('certs',this)" id="nav-certs">
      <span class="nav-icon">&#128274;</span><span>Certificates</span>
    </div>
    <div class="sidebar-sep"></div>
    <div class="nav" onclick="showView('stats',this)" id="nav-stats">
      <span class="nav-icon">&#9641;</span><span>Analytics</span>
    </div>
    <div class="nav" onclick="showView('settings',this)" id="nav-settings">
      <span class="nav-icon">&#9881;</span><span>Settings</span>
    </div>
    <div class="nav" onclick="showView('about',this)" id="nav-about">
      <span class="nav-icon">&#9432;</span><span>About</span>
    </div>
  </nav>

  <!-- Content -->
  <main class="content">

    <!-- ── Overview ── -->
    <div id="view-overview" class="view active">
      <div class="page-header">
        <div><div class="page-title">Overview</div><div class="page-subtitle">System summary and recent activity</div></div>
      </div>
      <div id="overviewAlert" class="alert-bar">&#9888; Error rate spike detected</div>
      <div class="stats-grid">
        <div class="stat-card accent">
          <div class="stat-val" id="ovTotalReq">—</div>
          <div class="stat-label">Total Requests</div>
        </div>
        <div class="stat-card">
          <div class="stat-val" id="ovErrors">—</div>
          <div class="stat-label">5xx Errors</div>
          <div class="stat-sub" id="ovErrorRate">—</div>
        </div>
        <div class="stat-card">
          <div class="stat-val" id="ovAvgLat">—</div>
          <div class="stat-label">Avg Latency</div>
        </div>
        <div class="stat-card">
          <div class="stat-val" id="ovRoutes">—</div>
          <div class="stat-label">Active Routes</div>
          <div class="stat-sub" id="ovUptime">—</div>
        </div>
      </div>
      <div class="panel">
        <div class="panel-head">&#9685; Requests / minute (live)</div>
        <div class="panel-body"><canvas id="sparkChart" height="80"></canvas></div>
      </div>
      <div class="panel">
        <div class="panel-head">&#9889; Recent Traffic <span class="panel-head-right" id="feedCount">0 requests</span></div>
        <div class="feed" id="overviewFeed"><div class="feed-empty">Waiting for traffic...</div></div>
      </div>
    </div>

    <!-- ── Traffic ── -->
    <div id="view-traffic" class="view">
      <div class="page-header">
        <div><div class="page-title">Live Traffic</div><div class="page-subtitle">Real-time request feed</div></div>
      </div>
      <div class="panel">
        <div class="panel-head">
          &#9889; Request Feed
          <div class="panel-head-right">
            <div class="filter-bar">
              <input class="filter-input" id="filterHost" placeholder="filter by host..." oninput="applyFilters()">
              <select class="filter-input" id="filterMethod" onchange="applyFilters()" style="width:90px">
                <option value="">Method</option>
                <option>GET</option><option>POST</option><option>PUT</option>
                <option>DELETE</option><option>PATCH</option>
              </select>
              <select class="filter-input" id="filterStatus" onchange="applyFilters()" style="width:90px">
                <option value="">Status</option>
                <option value="2">2xx</option><option value="3">3xx</option>
                <option value="4">4xx</option><option value="5">5xx</option>
              </select>
              <button class="pause-btn" id="pauseBtn" onclick="togglePause()">&#9646;&#9646; Pause</button>
            </div>
          </div>
        </div>
        <div class="feed" id="trafficFeed" style="max-height:600px"><div class="feed-empty">Waiting for traffic...</div></div>
      </div>
    </div>

    <!-- ── Routes ── -->
    <div id="view-routes" class="view">
      <div class="page-header">
        <div><div class="page-title">Routes</div><div class="page-subtitle">Manage proxy routes</div></div>
      </div>
      <div class="panel">
        <div class="panel-head">&#43; Add Route</div>
        <div class="panel-body">
          <div class="add-form">
            <div class="field">
              <label>Hostname</label>
              <input id="newHost" placeholder="api.example.com" autocomplete="off">
            </div>
            <div class="field">
              <label>Backend Target</label>
              <input id="newTarget" placeholder="http://localhost:3000" autocomplete="off">
            </div>
            <div style="display:flex;flex-direction:column;gap:6px;justify-content:flex-end">
              <label class="checkbox-field">
                <input type="checkbox" id="newHTTPS">
                <span>Enable HTTPS</span>
              </label>
            </div>
            <div style="display:flex;align-items:flex-end">
              <button class="btn btn-primary" onclick="addRoute()">Add Route</button>
            </div>
          </div>
        </div>
      </div>
      <div class="panel">
        <div class="panel-head">&#8644; All Routes <span class="panel-head-right" id="routeCount"></span></div>
        <div class="rtable-wrap">
          <table class="rtable">
            <thead><tr>
              <th>Host</th><th>Target</th><th>HTTPS</th><th>Status</th><th style="width:120px">Actions</th>
            </tr></thead>
            <tbody id="routeTable"></tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ── Certificates ── -->
    <div id="view-certs" class="view">
      <div class="page-header">
        <div><div class="page-title">Certificates</div><div class="page-subtitle">TLS certificate status per route</div></div>
      </div>
      <div id="certsEmpty" class="panel" style="display:none">
        <div class="panel-body" style="text-align:center;padding:48px;color:var(--text2)">
          No HTTPS routes configured. Enable HTTPS on a route to see certificates here.
        </div>
      </div>
      <div class="cert-grid" id="certGrid"></div>
    </div>

    <!-- ── Analytics ── -->
    <div id="view-stats" class="view">
      <div class="page-header">
        <div><div class="page-title">Analytics</div><div class="page-subtitle">Request statistics per route</div></div>
      </div>
      <div class="charts-grid">
        <div class="chart-wrap">
          <div class="chart-title">Requests by Route (total)</div>
          <canvas id="routeBarChart" class="chart-canvas"></canvas>
        </div>
        <div class="chart-wrap">
          <div class="chart-title">Error Rate by Route (%)</div>
          <canvas id="errorChart" class="chart-canvas"></canvas>
        </div>
      </div>
      <div class="panel">
        <div class="panel-head">&#9641; Route Breakdown</div>
        <div class="panel-body">
          <div class="route-stats-grid" id="routeStatsGrid"></div>
        </div>
      </div>
    </div>

    <!-- ── Settings ── -->
    <div id="view-settings" class="view">
      <div class="page-header">
        <div><div class="page-title">Settings</div></div>
      </div>
      <div class="settings-section">
        <div class="settings-title">Change Password</div>
        <div class="settings-desc">Update your dashboard login password.</div>
        <div class="form-stack">
          <div class="field"><label>Current Password</label><input type="password" id="pwCurrent" autocomplete="current-password"></div>
          <div class="field"><label>New Password</label><input type="password" id="pwNew" autocomplete="new-password"></div>
          <div class="field"><label>Confirm New Password</label><input type="password" id="pwConfirm" autocomplete="new-password"></div>
          <div><button class="btn btn-primary" onclick="changePassword()">Update Password</button></div>
        </div>
      </div>
    </div>

    <!-- ── About ── -->
    <div id="view-about" class="view">
      <div class="page-header">
        <div><div class="page-title">About</div></div>
      </div>
      <div class="about-grid">
        <div class="about-card"><div class="about-card-label">Version</div><div class="about-card-val" id="abVersion">—</div></div>
        <div class="about-card"><div class="about-card-label">Uptime</div><div class="about-card-val" id="abUptime">—</div></div>
        <div class="about-card"><div class="about-card-label">Started</div><div class="about-card-val" id="abStart">—</div></div>
      </div>
      <div class="panel">
        <div class="panel-head">&#128279; Links</div>
        <div class="panel-body">
          <div class="about-links">
            <a class="about-link" href="https://github.com/Elchi-dev/onyx" target="_blank">&#128279; GitHub</a>
            <a class="about-link" href="https://github.com/Elchi-dev/onyx/releases" target="_blank">&#128196; Releases</a>
            <a class="about-link" href="https://github.com/Elchi-dev/onyx/issues" target="_blank">&#128030; Issues</a>
            <a class="about-link" href="https://github.com/Elchi-dev/onyx/blob/main/docs/configuration.md" target="_blank">&#128218; Docs</a>
          </div>
        </div>
      </div>
    </div>

  </main>
</div>

<div id="toasts"></div>

<script>
// ── State ──────────────────────────────────────────────────────────────────
var ws = null;
var wsReconnectTimer = null;
var wsReconnectDelay = 1000;
var allEvents = [];
var paused = false;
var filterHostVal = '';
var filterMethodVal = '';
var filterStatusVal = '';
var sparkData = new Array(60).fill(0);
var sparkChart = null;
var routeBarChart = null;
var errorChart = null;
var reqCountWindow = [];

// ── View switching ─────────────────────────────────────────────────────────
function showView(name, el) {
  document.querySelectorAll('.view').forEach(function(v) { v.classList.remove('active'); });
  document.querySelectorAll('.nav').forEach(function(n) { n.classList.remove('active'); });
  document.getElementById('view-' + name).classList.add('active');
  if (el) el.classList.add('active');
  else { var n = document.getElementById('nav-' + name); if (n) n.classList.add('active'); }

  if (name === 'routes') loadRoutes();
  if (name === 'certs') loadCerts();
  if (name === 'stats') loadStats();
  if (name === 'about') loadAbout();
}

// ── Toast notifications ────────────────────────────────────────────────────
function toast(msg, type) {
  var el = document.createElement('div');
  el.className = 'toast ' + (type || 'info');
  el.textContent = msg;
  document.getElementById('toasts').appendChild(el);
  setTimeout(function() {
    el.style.animation = 'slideOut .2s ease forwards';
    setTimeout(function() { el.remove(); }, 200);
  }, 3000);
}

// ── WebSocket ──────────────────────────────────────────────────────────────
function connectWS() {
  var proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
  ws = new WebSocket(proto + '//' + location.host + '/ws');

  ws.onopen = function() {
    wsReconnectDelay = 1000;
    setWSDot('connected', 'Live');
    clearTimeout(wsReconnectTimer);
  };

  ws.onmessage = function(e) {
    try {
      var msg = JSON.parse(e.data);
      if (msg.type === 'request') handleRequest(msg.payload);
      if (msg.type === 'routes_changed') {
        if (document.getElementById('view-routes').classList.contains('active')) loadRoutes();
        if (document.getElementById('view-certs').classList.contains('active')) loadCerts();
        loadOverview();
      }
    } catch(ex) {}
  };

  ws.onerror = function() { setWSDot('connecting', 'Reconnecting'); };
  ws.onclose = function() {
    setWSDot('connecting', 'Disconnected');
    wsReconnectTimer = setTimeout(connectWS, wsReconnectDelay);
    wsReconnectDelay = Math.min(wsReconnectDelay * 2, 30000);
  };
}

function setWSDot(state, label) {
  var dot = document.getElementById('wsDot');
  dot.className = 'ws-dot ' + state;
  document.getElementById('wsLabel').textContent = label;
}

// ── Request handling ───────────────────────────────────────────────────────
function handleRequest(r) {
  // Track for sparkline.
  reqCountWindow.push(Date.now());

  allEvents.unshift(r);
  if (allEvents.length > 500) allEvents.pop();

  // Update overview counters.
  loadOverview();

  if (!paused) {
    var row = buildFeedRow(r);
    appendToFeed('overviewFeed', row, 50);
    if (matchesFilter(r)) appendToFeed('trafficFeed', row.cloneNode(true), 200);
  }

  // Update feed count.
  document.getElementById('feedCount').textContent = allEvents.length + ' requests';
}

function buildFeedRow(r) {
  var now = new Date(r.timestamp);
  var ts = pad(now.getHours()) + ':' + pad(now.getMinutes()) + ':' + pad(now.getSeconds());
  var latMs = r.latency_ms;
  var latClass = latMs > 2000 ? 'lat vslow' : latMs > 500 ? 'lat slow' : 'lat';
  var sc = String(r.status);
  var scClass = sc[0] === '2' ? 'sc sc-2' : sc[0] === '3' ? 'sc sc-3' : sc[0] === '4' ? 'sc sc-4' : 'sc sc-5';

  var row = document.createElement('div');
  row.className = 'feed-row';
  row.innerHTML =
    '<span class="ts">' + esc(ts) + '</span>' +
    '<span class="meth">' + esc(r.method) + '</span>' +
    '<span class="hcol">' + esc(r.host) + '</span>' +
    '<span class="pcol">' + esc(r.path) + '</span>' +
    '<span class="' + scClass + '">' + esc(sc) + '</span>' +
    '<span class="' + latClass + '">' + latMs + 'ms</span>';
  return row;
}

function appendToFeed(id, row, limit) {
  var feed = document.getElementById(id);
  var empty = feed.querySelector('.feed-empty');
  if (empty) empty.remove();
  feed.insertBefore(row, feed.firstChild);
  while (feed.children.length > limit) feed.removeChild(feed.lastChild);
}

function matchesFilter(r) {
  if (filterHostVal && r.host.indexOf(filterHostVal) < 0) return false;
  if (filterMethodVal && r.method !== filterMethodVal) return false;
  if (filterStatusVal && String(r.status)[0] !== filterStatusVal) return false;
  return true;
}

function applyFilters() {
  filterHostVal = document.getElementById('filterHost').value.toLowerCase();
  filterMethodVal = document.getElementById('filterMethod').value;
  filterStatusVal = document.getElementById('filterStatus').value;

  var feed = document.getElementById('trafficFeed');
  feed.innerHTML = '<div class="feed-empty">Waiting for traffic...</div>';
  var filtered = allEvents.filter(matchesFilter);
  filtered.slice(0, 200).forEach(function(r) { appendToFeed('trafficFeed', buildFeedRow(r), 200); });
}

function togglePause() {
  paused = !paused;
  var btn = document.getElementById('pauseBtn');
  btn.textContent = paused ? '&#9654; Resume' : '&#9646;&#9646; Pause';
  btn.className = paused ? 'pause-btn paused' : 'pause-btn';
  if (paused) btn.innerHTML = '&#9654; Resume';
  else btn.innerHTML = '&#9646;&#9646; Pause';
}

// ── Overview ───────────────────────────────────────────────────────────────
function loadOverview() {
  fetch('/api/stats').then(function(r) { return r.json(); }).then(function(d) {
    var g = d.global || {};
    document.getElementById('ovTotalReq').textContent = fmt(g.TotalRequests || 0);
    document.getElementById('ovErrors').textContent = fmt(g.TotalErrors || 0);
    var rate = g.TotalRequests > 0 ? ((g.TotalErrors / g.TotalRequests) * 100).toFixed(1) : '0.0';
    document.getElementById('ovErrorRate').textContent = rate + '% error rate';
    document.getElementById('ovAvgLat').textContent = Math.round(g.AvgLatency || 0) + 'ms';
    document.getElementById('ovRoutes').textContent = g.RouteCount || 0;
    document.getElementById('ovUptime').textContent = d.uptime || '—';

    // Error spike alert.
    var alertBar = document.getElementById('overviewAlert');
    if (parseFloat(rate) > 10 && g.TotalRequests > 10) alertBar.classList.add('show');
    else alertBar.classList.remove('show');
  }).catch(function() {});
}

// ── Sparkline ──────────────────────────────────────────────────────────────
function initSparkChart() {
  var ctx = document.getElementById('sparkChart').getContext('2d');
  sparkChart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: sparkData.map(function() { return ''; }),
      datasets: [{
        data: sparkData,
        borderColor: '#6366f1',
        backgroundColor: 'rgba(99,102,241,.08)',
        borderWidth: 2,
        pointRadius: 0,
        fill: true,
        tension: 0.4
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      plugins: { legend: { display: false }, tooltip: { enabled: false } },
      scales: {
        x: { display: false },
        y: { display: true, min: 0, grid: { color: 'rgba(255,255,255,.05)' }, ticks: { color: '#4a4a6a', font: { size: 10 }, maxTicksLimit: 4 } }
      }
    }
  });
}

function updateSparkline() {
  var now = Date.now();
  var oneSecAgo = now - 1000;
  reqCountWindow = reqCountWindow.filter(function(t) { return t > now - 60000; });
  var lastSecCount = reqCountWindow.filter(function(t) { return t > oneSecAgo; }).length;
  sparkData.push(lastSecCount);
  sparkData.shift();
  if (sparkChart) {
    sparkChart.data.datasets[0].data = sparkData.slice();
    sparkChart.update('none');
  }
}

// ── Routes ─────────────────────────────────────────────────────────────────
function loadRoutes() {
  fetch('/api/routes').then(function(r) { return r.json(); }).then(function(routes) {
    var tbody = document.getElementById('routeTable');
    tbody.innerHTML = '';
    document.getElementById('routeCount').textContent = (routes ? routes.length : 0) + ' routes';
    if (!routes || routes.length === 0) {
      tbody.innerHTML = '<tr><td colspan="5" style="padding:32px;text-align:center;color:var(--muted)">No routes yet. Add one above.</td></tr>';
      return;
    }
    routes.forEach(function(r) {
      var tr = document.createElement('tr');
      var httpsPill = r.HTTPS
        ? '<span class="pill pill-blue">&#128274; HTTPS</span>'
        : '<span class="pill pill-muted">HTTP</span>';
      var statusPill = r.Enabled
        ? '<span class="pill pill-green">&#9679; Active</span>'
        : '<span class="pill pill-muted">&#9679; Disabled</span>';
      tr.innerHTML =
        '<td class="mono">' + esc(r.Host) + '</td>' +
        '<td class="mono" style="color:var(--text2)">' + esc(r.Target) + '</td>' +
        '<td>' + httpsPill + '</td>' +
        '<td>' + statusPill + '</td>' +
        '<td><div class="actions">' +
          '<label class="toggle" title="' + (r.Enabled ? 'Disable' : 'Enable') + '">' +
            '<input type="checkbox" ' + (r.Enabled ? 'checked' : '') + ' onchange="toggleRoute(\'' + esc(r.Host) + '\',this.checked)">' +
            '<span class="toggle-slider"></span>' +
          '</label>' +
          '<button class="icon-btn" title="Toggle HTTPS" onclick="toggleHTTPS(\'' + esc(r.Host) + '\',' + !r.HTTPS + ')">&#128274;</button>' +
          '<button class="icon-btn del" title="Delete" onclick="deleteRoute(\'' + esc(r.Host) + '\')">&#128465;</button>' +
        '</div></td>';
      tbody.appendChild(tr);
    });
  }).catch(function() { toast('Failed to load routes', 'err'); });
}

function addRoute() {
  var host = document.getElementById('newHost').value.trim();
  var target = document.getElementById('newTarget').value.trim();
  var https = document.getElementById('newHTTPS').checked;
  if (!host || !target) { toast('Host and target are required', 'err'); return; }
  fetch('/api/routes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ host: host, target: target, https: https })
  }).then(function(r) { return r.json(); }).then(function(d) {
    if (d.error) { toast(d.error, 'err'); return; }
    document.getElementById('newHost').value = '';
    document.getElementById('newTarget').value = '';
    document.getElementById('newHTTPS').checked = false;
    toast('Route added: ' + host, 'ok');
    loadRoutes();
  }).catch(function() { toast('Failed to add route', 'err'); });
}

function deleteRoute(host) {
  if (!confirm('Delete route for ' + host + '?')) return;
  fetch('/api/routes/' + encodeURIComponent(host), { method: 'DELETE' })
    .then(function(r) { return r.json(); })
    .then(function(d) {
      if (d.error) { toast(d.error, 'err'); return; }
      toast('Route deleted: ' + host, 'ok');
      loadRoutes();
    }).catch(function() { toast('Failed to delete route', 'err'); });
}

function toggleRoute(host, enabled) {
  fetch('/api/routes/' + encodeURIComponent(host), {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ enabled: enabled })
  }).then(function(r) { return r.json(); }).then(function(d) {
    if (d.error) { toast(d.error, 'err'); return; }
    toast((enabled ? 'Enabled' : 'Disabled') + ': ' + host, 'ok');
    loadRoutes();
  }).catch(function() { toast('Failed to update route', 'err'); });
}

function toggleHTTPS(host, https) {
  fetch('/api/routes/' + encodeURIComponent(host), {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ https: https })
  }).then(function(r) { return r.json(); }).then(function(d) {
    if (d.error) { toast(d.error, 'err'); return; }
    toast((https ? 'HTTPS enabled' : 'HTTPS disabled') + ': ' + host, 'ok');
    loadRoutes();
    loadCerts();
  }).catch(function() { toast('Failed to update route', 'err'); });
}

// ── Certificates ───────────────────────────────────────────────────────────
function loadCerts() {
  fetch('/api/certs').then(function(r) { return r.json(); }).then(function(certs) {
    var grid = document.getElementById('certGrid');
    var empty = document.getElementById('certsEmpty');
    grid.innerHTML = '';
    if (!certs || certs.length === 0) {
      empty.style.display = '';
      return;
    }
    empty.style.display = 'none';
    certs.forEach(function(c) {
      var card = document.createElement('div');
      card.className = 'cert-card ' + c.status;
      var statusPill = {
        valid: '<span class="pill pill-green">&#10003; Valid</span>',
        expiring_soon: '<span class="pill pill-yellow">&#9888; Expiring Soon</span>',
        error: '<span class="pill pill-red">&#10005; Error</span>',
        pending: '<span class="pill pill-muted">&#8635; Pending</span>'
      }[c.status] || '<span class="pill pill-muted">' + c.status + '</span>';
      var modePill = c.mode === 'acme'
        ? '<span class="pill pill-blue">Let\'s Encrypt</span>'
        : '<span class="pill pill-muted">Self-Signed</span>';
      var expiry = c.expires_at
        ? new Date(c.expires_at).toLocaleDateString()
        : '—';
      card.innerHTML =
        '<div class="cert-host">' + esc(c.host) + '</div>' +
        '<div class="cert-meta">' +
          '<div class="cert-row"><span class="label">Status</span>' + statusPill + '</div>' +
          '<div class="cert-row"><span class="label">Mode</span>' + modePill + '</div>' +
          '<div class="cert-row"><span class="label">Expires</span><span class="val">' + expiry + '</span></div>' +
        '</div>';
      grid.appendChild(card);
    });
  }).catch(function() {});
}

// ── Analytics ──────────────────────────────────────────────────────────────
function loadStats() {
  fetch('/api/stats').then(function(r) { return r.json(); }).then(function(d) {
    var routes = d.per_route || [];
    var grid = document.getElementById('routeStatsGrid');
    grid.innerHTML = '';

    var maxTotal = Math.max.apply(null, routes.map(function(r) { return r.Total || 0; }));

    routes.forEach(function(r) {
      var card = document.createElement('div');
      card.className = 'rs-card';
      var errRate = r.Total > 0 ? ((r.Errors / r.Total) * 100).toFixed(1) : '0.0';
      var fill = maxTotal > 0 ? Math.round((r.Total / maxTotal) * 100) : 0;
      var hasErrors = r.Errors > 0;
      card.innerHTML =
        '<div class="rs-host">' + esc(r.Host) + '</div>' +
        '<div class="rs-metrics">' +
          '<div><div class="rs-val">' + fmt(r.Total || 0) + '</div><div class="rs-lbl">Requests</div></div>' +
          '<div><div class="rs-val" style="color:' + (hasErrors ? 'var(--red)' : 'var(--green)') + '">' + (r.Errors || 0) + '</div><div class="rs-lbl">Errors</div></div>' +
          '<div><div class="rs-val">' + Math.round(r.AvgLatency || 0) + 'ms</div><div class="rs-lbl">Avg Lat</div></div>' +
        '</div>' +
        '<div class="rs-bar"><div class="rs-fill ' + (hasErrors ? 'has-errors' : '') + '" style="width:' + fill + '%"></div></div>';
      grid.appendChild(card);
    });

    // Charts.
    var labels = routes.map(function(r) { return r.Host; });
    var totals = routes.map(function(r) { return r.Total || 0; });
    var errRates = routes.map(function(r) { return r.Total > 0 ? ((r.Errors / r.Total) * 100).toFixed(1) : 0; });

    if (routeBarChart) routeBarChart.destroy();
    var ctx1 = document.getElementById('routeBarChart').getContext('2d');
    routeBarChart = new Chart(ctx1, {
      type: 'bar',
      data: {
        labels: labels,
        datasets: [{ label: 'Requests', data: totals, backgroundColor: 'rgba(99,102,241,.7)', borderRadius: 4 }]
      },
      options: chartOpts('Requests')
    });

    if (errorChart) errorChart.destroy();
    var ctx2 = document.getElementById('errorChart').getContext('2d');
    errorChart = new Chart(ctx2, {
      type: 'bar',
      data: {
        labels: labels,
        datasets: [{ label: 'Error %', data: errRates, backgroundColor: 'rgba(239,68,68,.7)', borderRadius: 4 }]
      },
      options: chartOpts('%')
    });
  }).catch(function() {});
}

function chartOpts(unit) {
  return {
    responsive: true,
    maintainAspectRatio: false,
    plugins: { legend: { display: false } },
    scales: {
      x: { grid: { color: 'rgba(255,255,255,.04)' }, ticks: { color: '#9090b8', font: { size: 11 } } },
      y: { grid: { color: 'rgba(255,255,255,.04)' }, ticks: { color: '#9090b8', font: { size: 11 } }, beginAtZero: true }
    }
  };
}

// ── About ──────────────────────────────────────────────────────────────────
function loadAbout() {
  fetch('/api/about').then(function(r) { return r.json(); }).then(function(d) {
    document.getElementById('abVersion').textContent = d.version || '—';
    document.getElementById('abUptime').textContent = d.uptime || '—';
    var start = d.start_time ? new Date(d.start_time).toLocaleString() : '—';
    document.getElementById('abStart').textContent = start;
  }).catch(function() {});
}

// ── Settings ───────────────────────────────────────────────────────────────
function changePassword() {
  var cur = document.getElementById('pwCurrent').value;
  var nw = document.getElementById('pwNew').value;
  var cf = document.getElementById('pwConfirm').value;
  if (!cur || !nw) { toast('Fill in all fields', 'err'); return; }
  if (nw !== cf) { toast('Passwords do not match', 'err'); return; }
  if (nw.length < 8) { toast('Password must be at least 8 characters', 'err'); return; }
  fetch('/api/settings/password', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ current: cur, 'new': nw })
  }).then(function(r) { return r.json(); }).then(function(d) {
    if (d.error) { toast(d.error, 'err'); return; }
    toast('Password updated successfully', 'ok');
    document.getElementById('pwCurrent').value = '';
    document.getElementById('pwNew').value = '';
    document.getElementById('pwConfirm').value = '';
  }).catch(function() { toast('Failed to update password', 'err'); });
}

// ── Helpers ────────────────────────────────────────────────────────────────
function fmt(n) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  return String(n);
}
function pad(n) { return n < 10 ? '0' + n : String(n); }
function esc(s) {
  return String(s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ── Init ───────────────────────────────────────────────────────────────────
initSparkChart();
connectWS();
loadOverview();
setInterval(loadOverview, 10000);
setInterval(updateSparkline, 1000);
</script>
</body>
</html>`

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx — Sign in</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%236366f1'/><text y='22' x='5' font-size='18' font-family='monospace' fill='white'>&#9670;</text></svg>">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@600&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0c0c14;color:#e2e2f0;font-family:'Inter',sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center}
.card{background:#111120;border:1px solid #232338;border-radius:14px;padding:40px;width:100%;max-width:380px;box-shadow:0 24px 64px rgba(0,0,0,.5)}
.logo{font-family:'JetBrains Mono',monospace;font-size:22px;font-weight:600;color:#6366f1;text-align:center;margin-bottom:6px;letter-spacing:-1px}
.logo-sub{text-align:center;font-size:13px;color:#4a4a6a;margin-bottom:32px}
.field{margin-bottom:16px}
.field label{display:block;font-size:11px;color:#9090b8;text-transform:uppercase;letter-spacing:.07em;margin-bottom:6px;font-weight:500}
.field input{width:100%;background:#0c0c14;border:1px solid #232338;border-radius:8px;padding:11px 14px;color:#e2e2f0;font-size:14px;font-family:'Inter',sans-serif;outline:none;transition:border-color .15s}
.field input:focus{border-color:#6366f1}
.remember{display:flex;align-items:center;gap:8px;margin-bottom:22px;cursor:pointer}
.remember input{accent-color:#6366f1;width:14px;height:14px}
.remember span{font-size:13px;color:#9090b8}
.btn{width:100%;background:#6366f1;color:#fff;border:none;border-radius:8px;padding:12px;font-size:14px;font-weight:500;font-family:'Inter',sans-serif;cursor:pointer;transition:background .15s}
.btn:hover{background:#818cf8}
.error{background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:8px;padding:10px 14px;font-size:13px;color:#ef4444;margin-bottom:16px}
</style>
</head>
<body>
<div class="card">
  <div class="logo">&#9670; onyx</div>
  <div class="logo-sub">Dashboard</div>
  <form method="POST" action="/login">
    <div class="field">
      <label>Password</label>
      <input type="password" name="password" autofocus autocomplete="current-password" required>
    </div>
    <label class="remember">
      <input type="checkbox" name="remember_me">
      <span>Keep me signed in for 7 days</span>
    </label>
    <button type="submit" class="btn">Sign in</button>
  </form>
</div>
</body>
</html>`

const loginErrorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx — Sign in</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%236366f1'/><text y='22' x='5' font-size='18' font-family='monospace' fill='white'>&#9670;</text></svg>">
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&family=JetBrains+Mono:wght@600&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0c0c14;color:#e2e2f0;font-family:'Inter',sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center}
.card{background:#111120;border:1px solid #232338;border-radius:14px;padding:40px;width:100%;max-width:380px}
.logo{font-family:'JetBrains Mono',monospace;font-size:22px;font-weight:600;color:#6366f1;text-align:center;margin-bottom:6px;letter-spacing:-1px}
.logo-sub{text-align:center;font-size:13px;color:#4a4a6a;margin-bottom:32px}
.field{margin-bottom:16px}
.field label{display:block;font-size:11px;color:#9090b8;text-transform:uppercase;letter-spacing:.07em;margin-bottom:6px;font-weight:500}
.field input{width:100%;background:#0c0c14;border:1px solid #232338;border-radius:8px;padding:11px 14px;color:#e2e2f0;font-size:14px;outline:none;transition:border-color .15s}
.field input:focus{border-color:#6366f1}
.remember{display:flex;align-items:center;gap:8px;margin-bottom:22px;cursor:pointer}
.remember input{accent-color:#6366f1;width:14px;height:14px}
.remember span{font-size:13px;color:#9090b8}
.btn{width:100%;background:#6366f1;color:#fff;border:none;border-radius:8px;padding:12px;font-size:14px;font-weight:500;font-family:'Inter',sans-serif;cursor:pointer;transition:background .15s}
.btn:hover{background:#818cf8}
.error{background:rgba(239,68,68,.1);border:1px solid rgba(239,68,68,.3);border-radius:8px;padding:10px 14px;font-size:13px;color:#ef4444;margin-bottom:16px}
</style>
</head>
<body>
<div class="card">
  <div class="logo">&#9670; onyx</div>
  <div class="logo-sub">Dashboard</div>
  <div class="error">Incorrect password. Please try again.</div>
  <form method="POST" action="/login">
    <div class="field">
      <label>Password</label>
      <input type="password" name="password" autofocus autocomplete="current-password" required>
    </div>
    <label class="remember">
      <input type="checkbox" name="remember_me">
      <span>Keep me signed in for 7 days</span>
    </label>
    <button type="submit" class="btn">Sign in</button>
  </form>
</div>
</body>
</html>`

const loginRateLimitHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Onyx — Too Many Attempts</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600&display=swap" rel="stylesheet">
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{background:#0c0c14;color:#e2e2f0;font-family:'Inter',sans-serif;min-height:100vh;display:flex;align-items:center;justify-content:center}
.card{background:#111120;border:1px solid #232338;border-radius:14px;padding:40px;width:100%;max-width:380px;text-align:center}
.icon{font-size:40px;margin-bottom:16px}
h1{font-size:20px;font-weight:600;margin-bottom:8px;color:#ef4444}
p{font-size:13px;color:#9090b8;line-height:1.6}
.back{display:inline-block;margin-top:24px;color:#6366f1;font-size:13px;text-decoration:none}
.back:hover{text-decoration:underline}
</style>
</head>
<body>
<div class="card">
  <div class="icon">&#128274;</div>
  <h1>Too Many Attempts</h1>
  <p>Login attempts exceeded. Please wait 60 seconds before trying again.</p>
  <a class="back" href="/login">&#8592; Back to login</a>
</div>
</body>
</html>`
