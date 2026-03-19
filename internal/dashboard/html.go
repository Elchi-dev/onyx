package dashboard

// dashboardHTML is the full dashboard UI, embedded directly in the binary.
const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx Dashboard</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%237c6af7'/><text y='24' x='4' font-size='22' font-family='monospace' fill='white'>&#9670;</text></svg>">
<style>
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600&family=Inter:wght@400;500;600&display=swap');
:root{
  --bg:#0a0a0f;--surface:#0f0f18;--surface2:#151520;--border:#1c1c2e;
  --accent:#7c6af7;--accent-dim:rgba(124,106,247,.15);
  --green:#3ddc97;--red:#ff5370;--yellow:#ffcb6b;--blue:#82aaff;
  --text:#e4e4ef;--muted:#555570;--muted2:#3a3a52;
}
*{box-sizing:border-box;margin:0;padding:0;}
html,body{height:100%;overflow:hidden;}
body{background:var(--bg);color:var(--text);font-family:'Inter',sans-serif;font-size:14px;display:flex;flex-direction:column;}
.hdr{display:flex;align-items:center;gap:16px;padding:0 20px;height:52px;background:var(--surface);border-bottom:1px solid var(--border);flex-shrink:0;}
.logo{font-family:'JetBrains Mono',monospace;font-weight:600;font-size:15px;color:var(--accent);letter-spacing:-.5px;margin-right:4px;}
.logo-dot{color:var(--green);}
.hdr-spacer{flex:1;}
.badge{display:flex;align-items:center;gap:5px;font-size:11px;color:var(--muted);font-family:'JetBrains Mono',monospace;padding:4px 10px;background:var(--surface2);border:1px solid var(--border);border-radius:20px;}
.dot{width:6px;height:6px;border-radius:50%;background:var(--muted);}
.dot.live{background:var(--green);box-shadow:0 0 6px var(--green);animation:blink 2s ease-in-out infinite;}
.dot.running{background:var(--green);box-shadow:0 0 6px var(--green);}
@keyframes blink{0%,100%{opacity:1}50%{opacity:.3}}
.btn-logout{font-size:11px;color:var(--muted);background:none;border:1px solid var(--border);border-radius:6px;padding:4px 10px;cursor:pointer;font-family:'Inter',sans-serif;transition:all .15s;}
.btn-logout:hover{color:var(--red);border-color:var(--red);}
.layout{display:flex;flex:1;overflow:hidden;}
.sidebar{width:200px;background:var(--surface);border-right:1px solid var(--border);padding:12px 8px;display:flex;flex-direction:column;gap:2px;flex-shrink:0;}
.nav{display:flex;align-items:center;gap:9px;padding:8px 10px;border-radius:7px;cursor:pointer;color:var(--muted);transition:all .12s;font-size:13px;user-select:none;}
.nav:hover{background:var(--surface2);color:var(--text);}
.nav.active{background:var(--accent-dim);color:var(--accent);font-weight:500;}
.nav-icon{font-size:14px;width:18px;text-align:center;}
.sidebar-sep{height:1px;background:var(--border);margin:6px 4px;}
.content{flex:1;overflow-y:auto;padding:24px;display:flex;flex-direction:column;gap:20px;}
.stats-row{display:grid;grid-template-columns:repeat(4,1fr);gap:12px;}
.stat{background:var(--surface);border:1px solid var(--border);border-radius:10px;padding:16px 18px;}
.stat.accent{border-color:rgba(124,106,247,.4);}
.stat-val{font-family:'JetBrains Mono',monospace;font-size:28px;font-weight:600;line-height:1;color:var(--text);}
.stat.accent .stat-val{color:var(--accent);}
.stat-label{font-size:11px;color:var(--muted);margin-top:5px;text-transform:uppercase;letter-spacing:.06em;}
.panel{background:var(--surface);border:1px solid var(--border);border-radius:10px;overflow:hidden;}
.panel-head{display:flex;align-items:center;padding:12px 16px;border-bottom:1px solid var(--border);font-size:11px;font-weight:600;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);gap:8px;}
.panel-head .count{margin-left:auto;font-weight:400;}
.panel-body{padding:16px;}
.feed{font-family:'JetBrains Mono',monospace;font-size:12px;max-height:400px;overflow-y:auto;}
.feed-empty{text-align:center;padding:40px;color:var(--muted);font-family:'Inter',sans-serif;}
.feed-row{display:grid;grid-template-columns:110px 48px minmax(0,1.2fr) minmax(0,1fr) 52px 64px;gap:6px;padding:6px 8px;align-items:center;border-radius:4px;transition:background .08s;}
.feed-row:hover{background:rgba(255,255,255,.03);}
.feed-row+.feed-row{border-top:1px solid rgba(255,255,255,.03);}
.ts{color:var(--muted);}
.meth{font-weight:600;color:var(--blue);}
.hcol{color:var(--text);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.pcol{color:var(--muted);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.sc{display:inline-block;padding:1px 6px;border-radius:3px;font-size:11px;font-weight:700;text-align:center;}
.sc-ok{background:rgba(61,220,151,.1);color:var(--green);}
.sc-redir{background:rgba(255,203,107,.1);color:var(--yellow);}
.sc-err{background:rgba(255,83,112,.1);color:var(--red);}
.lat{color:var(--muted);text-align:right;}
.lat.slow{color:var(--yellow);}
.lat.very-slow{color:var(--red);}
.routes-wrap{overflow-x:auto;}
.rtable{width:100%;border-collapse:collapse;font-size:13px;}
.rtable th{text-align:left;font-size:10px;text-transform:uppercase;letter-spacing:.08em;color:var(--muted);padding:8px 12px;border-bottom:1px solid var(--border);font-weight:600;}
.rtable td{padding:11px 12px;border-bottom:1px solid rgba(255,255,255,.04);}
.rtable tr:last-child td{border-bottom:none;}
.rtable tr:hover td{background:rgba(255,255,255,.02);}
.mono{font-family:'JetBrains Mono',monospace;font-size:12px;}
.pill{display:inline-block;padding:2px 8px;border-radius:20px;font-size:11px;font-weight:600;}
.pill-on{background:rgba(61,220,151,.1);color:var(--green);}
.pill-off{background:rgba(90,90,120,.15);color:var(--muted);}
.icon-btn{background:none;border:none;cursor:pointer;padding:4px 6px;border-radius:4px;font-size:13px;color:var(--muted);transition:all .12s;}
.icon-btn:hover{background:var(--surface2);color:var(--text);}
.icon-btn.del:hover{color:var(--red);}
.actions-cell{display:flex;gap:4px;align-items:center;}
.add-form{display:flex;gap:10px;align-items:flex-end;flex-wrap:wrap;}
.field{display:flex;flex-direction:column;gap:5px;flex:1;min-width:160px;}
.field label{font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.06em;}
.field input{background:var(--bg);border:1px solid var(--border);border-radius:7px;padding:8px 12px;color:var(--text);font-family:'JetBrains Mono',monospace;font-size:12px;outline:none;transition:border-color .15s;width:100%;}
.field input:focus{border-color:var(--accent);}
.btn{padding:8px 18px;border-radius:7px;border:none;font-family:'Inter',sans-serif;font-size:13px;font-weight:500;cursor:pointer;transition:all .15s;}
.btn-primary{background:var(--accent);color:#fff;}
.btn-primary:hover{opacity:.85;}
.btn-primary:disabled{opacity:.4;cursor:not-allowed;}
.route-stats-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(260px,1fr));gap:12px;}
.rs-card{background:var(--surface2);border:1px solid var(--border);border-radius:8px;padding:14px 16px;}
.rs-host{font-family:'JetBrains Mono',monospace;font-size:12px;font-weight:600;color:var(--text);margin-bottom:10px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;}
.rs-metrics{display:grid;grid-template-columns:1fr 1fr 1fr;gap:8px;}
.rs-metric-val{font-family:'JetBrains Mono',monospace;font-size:16px;font-weight:600;color:var(--text);}
.rs-metric-label{font-size:10px;color:var(--muted);margin-top:2px;}
.rs-bar{height:3px;background:var(--border);border-radius:2px;margin-top:10px;}
.rs-bar-fill{height:100%;border-radius:2px;background:var(--accent);max-width:100%;}
.rs-bar-fill.has-errors{background:var(--red);}
.about-grid{display:grid;grid-template-columns:repeat(auto-fill,minmax(180px,1fr));gap:12px;}
.about-card{background:var(--surface2);border:1px solid var(--border);border-radius:8px;padding:14px 16px;}
.about-key{font-size:10px;text-transform:uppercase;letter-spacing:.07em;color:var(--muted);margin-bottom:4px;}
.about-val{font-family:'JetBrains Mono',monospace;font-size:13px;color:var(--text);}
.toast-wrap{position:fixed;bottom:20px;right:20px;display:flex;flex-direction:column;gap:8px;z-index:1000;}
.toast{padding:10px 16px;border-radius:8px;font-size:13px;display:flex;align-items:center;gap:8px;animation:slide-in .2s ease;max-width:320px;}
.toast-ok{background:rgba(61,220,151,.15);border:1px solid rgba(61,220,151,.3);color:var(--green);}
.toast-err{background:rgba(255,83,112,.15);border:1px solid rgba(255,83,112,.3);color:var(--red);}
@keyframes slide-in{from{transform:translateX(20px);opacity:0}to{transform:none;opacity:1}}
.view{display:none;}.view.active{display:block;}
.alert-bar{display:none;align-items:center;gap:10px;padding:8px 16px;background:rgba(255,83,112,.1);border:1px solid rgba(255,83,112,.3);border-radius:8px;font-size:12px;color:var(--red);}
.alert-bar.visible{display:flex;}
@media(max-width:900px){.stats-row{grid-template-columns:repeat(2,1fr);}.feed-row{grid-template-columns:90px 42px minmax(0,1fr) 46px 58px;}.hcol.hide-mobile{display:none;}}
@media(max-width:600px){.sidebar{display:none;}.stats-row{grid-template-columns:repeat(2,1fr);}.content{padding:14px;}}
</style>
</head>
<body>
<header class="hdr">
  <div class="logo">onyx<span class="logo-dot">.</span></div>
  <div class="badge"><div class="dot running"></div>running</div>
  <div class="badge" id="ws-badge"><div class="dot" id="ws-dot"></div><span id="ws-lbl">connecting</span></div>
  <div class="hdr-spacer"></div>
  <button class="btn-logout" onclick="logout()">Sign out</button>
</header>
<div class="layout">
  <nav class="sidebar">
    <div class="nav active" id="nav-live" onclick="showView('live',this)"><span class="nav-icon">&#9889;</span>Live Traffic</div>
    <div class="nav" id="nav-routes" onclick="showView('routes',this)"><span class="nav-icon">&#8644;</span>Routes</div>
    <div class="sidebar-sep"></div>
    <div class="nav" id="nav-stats" onclick="showView('stats',this)"><span class="nav-icon">&#128202;</span>Statistics</div>
    <div class="nav" id="nav-about" onclick="showView('about',this)"><span class="nav-icon">&#9881;</span>About</div>
  </nav>
  <main class="content">
    <div class="stats-row">
      <div class="stat accent"><div class="stat-val" id="s-total">--</div><div class="stat-label">Total Requests</div></div>
      <div class="stat"><div class="stat-val" id="s-rps">0</div><div class="stat-label">Req / sec</div></div>
      <div class="stat"><div class="stat-val" id="s-err">--</div><div class="stat-label">5xx Errors</div></div>
      <div class="stat"><div class="stat-val" id="s-lat">--</div><div class="stat-label">Avg Latency ms</div></div>
    </div>
    <div class="alert-bar" id="error-alert">&#128308; <span id="error-alert-msg"></span></div>
    <div id="view-live" class="view active">
      <div class="panel">
        <div class="panel-head">&#9889; Live Request Feed <span class="count" id="feed-count"></span></div>
        <div class="feed" id="feed"><div class="feed-empty">Waiting for requests...</div></div>
      </div>
    </div>
    <div id="view-routes" class="view">
      <div class="panel" style="margin-bottom:16px">
        <div class="panel-head">&#10133; Add Route</div>
        <div class="panel-body">
          <div class="add-form">
            <div class="field"><label>Hostname</label><input id="new-host" type="text" placeholder="api.example.com"></div>
            <div class="field"><label>Backend Target</label><input id="new-target" type="text" placeholder="http://localhost:3000"></div>
            <button class="btn btn-primary" id="add-btn" onclick="addRoute()">Add Route</button>
          </div>
        </div>
      </div>
      <div class="panel">
        <div class="panel-head">&#8644; Routes <span class="count" id="route-count"></span></div>
        <div id="routes-wrap" class="routes-wrap panel-body"><div class="feed-empty">Loading...</div></div>
      </div>
    </div>
    <div id="view-stats" class="view">
      <div class="panel">
        <div class="panel-head">&#128202; Per-Route Statistics</div>
        <div class="panel-body"><div class="route-stats-grid" id="stats-grid"><div class="feed-empty">Loading...</div></div></div>
      </div>
    </div>
    <div id="view-about" class="view">
      <div class="panel">
        <div class="panel-head">&#9881; About Onyx</div>
        <div class="panel-body">
          <div class="about-grid" id="about-grid"><div class="feed-empty">Loading...</div></div>
          <p style="margin-top:16px;font-size:12px;color:var(--muted)">
            <a href="https://github.com/Elchi-dev/onyx" style="color:var(--accent)" target="_blank">github.com/Elchi-dev/onyx</a>
            &nbsp;&middot;&nbsp; MIT License
          </p>
        </div>
      </div>
    </div>
  </main>
</div>
<div class="toast-wrap" id="toasts"></div>
<script>
var rpsWindow = [];
var prevErrors = 0;

function showView(name, el) {
  document.querySelectorAll('.view').forEach(function(v){v.classList.remove('active');});
  document.querySelectorAll('.nav').forEach(function(n){n.classList.remove('active');});
  document.getElementById('view-' + name).classList.add('active');
  if (el) el.classList.add('active');
  if (name === 'routes') loadRoutes();
  if (name === 'stats')  loadStats();
  if (name === 'about')  loadAbout();
}

function logout() {
  fetch('/logout', {method:'POST'}).finally(function(){location.href='/login';});
}

function toast(msg, ok) {
  if (ok === undefined) ok = true;
  var el = document.createElement('div');
  el.className = 'toast ' + (ok ? 'toast-ok' : 'toast-err');
  el.textContent = (ok ? '\u2713 ' : '\u2717 ') + msg;
  document.getElementById('toasts').appendChild(el);
  setTimeout(function(){el.remove();}, 3000);
}

function esc(s) {
  return String(s).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

async function apiFetch(url, opts) {
  opts = opts || {};
  opts.headers = Object.assign({'Content-Type':'application/json'}, opts.headers || {});
  var r = await fetch(url, opts);
  if (r.status === 401) { location.href = '/login'; throw new Error('unauthorized'); }
  if (!r.ok) { var j = await r.json().catch(function(){return {};}); throw new Error(j.error || r.statusText); }
  return r.json();
}

async function loadRoutes() {
  var wrap = document.getElementById('routes-wrap');
  try {
    var data = await apiFetch('/api/routes');
    var routes = data || [];
    document.getElementById('route-count').textContent = routes.length + ' route' + (routes.length !== 1 ? 's' : '');
    if (!routes.length) {
      wrap.innerHTML = '<div class="feed-empty">No routes yet. Add one above.</div>';
      return;
    }
    var rows = routes.map(function(r) {
      var enabledLabel = r.Enabled ? 'active' : 'disabled';
      var pillClass = r.Enabled ? 'pill-on' : 'pill-off';
      var toggleIcon = r.Enabled ? '&#9646;&#9646;' : '&#9654;';
      var toggleTitle = r.Enabled ? 'Disable' : 'Enable';
      var safeHost = esc(r.Host);
      var safeTarget = esc(r.Target);
      return '<tr>' +
        '<td class="mono">' + safeHost + '</td>' +
        '<td class="mono" style="color:var(--muted)">' + safeTarget + '</td>' +
        '<td><span class="pill ' + pillClass + '">' + enabledLabel + '</span></td>' +
        '<td style="color:var(--muted);font-family:\'JetBrains Mono\',monospace;font-size:12px" id="rt-' + btoa(r.Host).replace(/=/g,'') + '">--</td>' +
        '<td><div class="actions-cell">' +
          '<button class="icon-btn" title="' + toggleTitle + '" onclick="toggleRoute(\'' + safeHost + '\',' + (!r.Enabled) + ')">' + toggleIcon + '</button>' +
          '<button class="icon-btn del" title="Delete" onclick="deleteRoute(\'' + safeHost + '\')">&#128465;</button>' +
        '</div></td>' +
      '</tr>';
    });
    wrap.innerHTML = '<table class="rtable"><thead><tr><th>Host</th><th>Target</th><th>Status</th><th>Requests</th><th>Actions</th></tr></thead><tbody>' + rows.join('') + '</tbody></table>';
    apiFetch('/api/stats').then(function(s) {
      if (s && s.per_route) {
        s.per_route.forEach(function(rs) {
          var id = 'rt-' + btoa(rs.Host).replace(/=/g,'');
          var el = document.getElementById(id);
          if (el) el.textContent = rs.Total.toLocaleString();
        });
      }
    }).catch(function(){});
  } catch(e) {
    wrap.innerHTML = '<div class="feed-empty">Failed to load routes.</div>';
  }
}

async function addRoute() {
  var host   = document.getElementById('new-host').value.trim();
  var target = document.getElementById('new-target').value.trim();
  if (!host || !target) { toast('Host and target are required', false); return; }
  var btn = document.getElementById('add-btn');
  btn.disabled = true;
  try {
    await apiFetch('/api/routes', {method:'POST', body:JSON.stringify({host:host, target:target})});
    document.getElementById('new-host').value = '';
    document.getElementById('new-target').value = '';
    toast('Route added: ' + host);
    loadRoutes();
  } catch(e) {
    toast('Failed to add route: ' + e.message, false);
  } finally {
    btn.disabled = false;
  }
}

async function toggleRoute(host, enabled) {
  try {
    await apiFetch('/api/routes/' + encodeURIComponent(host), {method:'PATCH', body:JSON.stringify({enabled:enabled})});
    toast('Route ' + (enabled ? 'enabled' : 'disabled') + ': ' + host);
    loadRoutes();
  } catch(e) { toast('Failed to update route', false); }
}

async function deleteRoute(host) {
  if (!confirm('Delete route for ' + host + '?')) return;
  try {
    await apiFetch('/api/routes/' + encodeURIComponent(host), {method:'DELETE'});
    toast('Route deleted: ' + host);
    loadRoutes();
  } catch(e) { toast('Failed to delete route', false); }
}

async function loadStats() {
  var grid = document.getElementById('stats-grid');
  try {
    var data = await apiFetch('/api/stats');
    if (data.global) {
      document.getElementById('s-total').textContent = data.global.TotalRequests.toLocaleString();
      document.getElementById('s-err').textContent   = data.global.TotalErrors.toLocaleString();
      document.getElementById('s-lat').textContent   = data.global.AvgLatency.toFixed(0);
    }
    if (!data.per_route || !data.per_route.length) {
      grid.innerHTML = '<div class="feed-empty">No statistics yet.</div>';
      return;
    }
    var maxTotal = Math.max.apply(null, data.per_route.map(function(r){return r.Total;}));
    grid.innerHTML = data.per_route.map(function(r) {
      var pct = maxTotal > 0 ? (r.Total / maxTotal * 100).toFixed(0) : 0;
      var hasErr = r.Errors > 0;
      return '<div class="rs-card">' +
        '<div class="rs-host">' + esc(r.Host) + '</div>' +
        '<div class="rs-metrics">' +
          '<div><div class="rs-metric-val">' + r.Total.toLocaleString() + '</div><div class="rs-metric-label">Requests</div></div>' +
          '<div><div class="rs-metric-val" style="' + (hasErr ? 'color:var(--red)' : '') + '">' + r.Errors + '</div><div class="rs-metric-label">Errors</div></div>' +
          '<div><div class="rs-metric-val">' + r.AvgLatency.toFixed(0) + '</div><div class="rs-metric-label">Avg ms</div></div>' +
        '</div>' +
        '<div class="rs-bar"><div class="rs-bar-fill ' + (hasErr ? 'has-errors' : '') + '" style="width:' + pct + '%"></div></div>' +
      '</div>';
    }).join('');
  } catch(e) {
    grid.innerHTML = '<div class="feed-empty">Failed to load statistics.</div>';
  }
}

async function loadAbout() {
  var grid = document.getElementById('about-grid');
  try {
    var data = await apiFetch('/api/about');
    var items = [
      ['Version', data.version || 'dev'],
      ['Uptime', data.uptime || '--'],
      ['Started', data.start_time ? new Date(data.start_time).toLocaleString() : '--']
    ];
    grid.innerHTML = items.map(function(item) {
      return '<div class="about-card"><div class="about-key">' + item[0] + '</div><div class="about-val">' + item[1] + '</div></div>';
    }).join('');
  } catch(e) {
    grid.innerHTML = '<div class="feed-empty">Failed to load about info.</div>';
  }
}

function fmtTime(ts) {
  var d = new Date(ts);
  return d.toLocaleTimeString('en-GB',{hour12:false}) + '.' + String(d.getMilliseconds()).padStart(3,'0');
}
function scClass(s) { return s>=500 ? 'sc-err' : s>=300 ? 'sc-redir' : 'sc-ok'; }
function latClass(ms) { return ms>1000 ? 'very-slow' : ms>300 ? 'slow' : ''; }

function addFeedRow(e) {
  var feed = document.getElementById('feed');
  var em = feed.querySelector('.feed-empty');
  if (em) em.remove();
  var row = document.createElement('div');
  row.className = 'feed-row';
  row.innerHTML =
    '<span class="ts">' + fmtTime(e.timestamp) + '</span>' +
    '<span class="meth">' + esc(e.method) + '</span>' +
    '<span class="hcol">' + esc(e.host) + '</span>' +
    '<span class="pcol hide-mobile">' + esc(e.path) + '</span>' +
    '<span class="sc ' + scClass(e.status) + '">' + e.status + '</span>' +
    '<span class="lat ' + latClass(e.latency_ms) + '">' + e.latency_ms + 'ms</span>';
  feed.insertBefore(row, feed.firstChild);
  while (feed.children.length > 500) feed.removeChild(feed.lastChild);
  rpsWindow.push(Date.now());
  document.getElementById('feed-count').textContent = feed.children.length + ' events';
}

setInterval(function() {
  apiFetch('/api/stats').then(function(data) {
    if (data && data.global) {
      document.getElementById('s-total').textContent = data.global.TotalRequests.toLocaleString();
      document.getElementById('s-err').textContent   = data.global.TotalErrors.toLocaleString();
      document.getElementById('s-lat').textContent   = data.global.AvgLatency.toFixed(0);
      if (data.global.TotalErrors > prevErrors + 5) {
        var bar = document.getElementById('error-alert');
        bar.classList.add('visible');
        document.getElementById('error-alert-msg').textContent = 'Error spike: ' + data.global.TotalErrors + ' total 5xx errors';
      }
      prevErrors = data.global.TotalErrors;
    }
  }).catch(function(){});
}, 5000);

setInterval(function() {
  var cutoff = Date.now() - 1000;
  rpsWindow = rpsWindow.filter(function(t){return t > cutoff;});
  document.getElementById('s-rps').textContent = rpsWindow.length;
}, 400);

function connect() {
  var proto = location.protocol === 'https:' ? 'wss' : 'ws';
  var ws = new WebSocket(proto + '://' + location.host + '/ws');
  var dot = document.getElementById('ws-dot');
  var lbl = document.getElementById('ws-lbl');
  ws.onopen = function() { dot.classList.add('live'); lbl.textContent = 'live'; };
  ws.onmessage = function(msg) {
    try {
      var ev = JSON.parse(msg.data);
      if (ev.type === 'request') { addFeedRow(ev.payload); rpsWindow.push(Date.now()); }
      if (ev.type === 'routes_changed') {
        if (document.getElementById('view-routes').classList.contains('active')) loadRoutes();
      }
    } catch(_) {}
  };
  ws.onclose = function() { dot.classList.remove('live'); lbl.textContent = 'reconnecting...'; setTimeout(connect, 3000); };
  ws.onerror = function() { ws.close(); };
}

connect();
apiFetch('/api/stats').then(function(data) {
  if (data && data.global) {
    document.getElementById('s-total').textContent = data.global.TotalRequests.toLocaleString();
    document.getElementById('s-err').textContent   = data.global.TotalErrors.toLocaleString();
    document.getElementById('s-lat').textContent   = data.global.AvgLatency.toFixed(0);
  }
}).catch(function(){});
</script>
</body>
</html>`

const loginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx -- Sign in</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%237c6af7'/><text y='24' x='4' font-size='22' font-family='monospace' fill='white'>&#9670;</text></svg>">
<style>
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Inter:wght@400;500;600&display=swap');
:root{--bg:#0a0a0f;--surface:#0f0f18;--border:#1c1c2e;--accent:#7c6af7;--text:#e4e4ef;--muted:#555570;}
*{box-sizing:border-box;margin:0;padding:0;}
body{background:var(--bg);color:var(--text);font-family:'Inter',sans-serif;min-height:100vh;display:grid;place-items:center;}
.card{background:var(--surface);border:1px solid var(--border);border-radius:14px;padding:36px 40px;width:360px;}
.logo{font-family:'JetBrains Mono',monospace;font-size:22px;font-weight:600;color:var(--accent);margin-bottom:4px;}
.logo span{color:#3ddc97;}
.sub{color:var(--muted);font-size:13px;margin-bottom:28px;}
label{display:block;font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.07em;margin-bottom:6px;}
input[type=password]{width:100%;background:var(--bg);border:1px solid var(--border);border-radius:8px;padding:11px 14px;color:var(--text);font-family:'JetBrains Mono',monospace;font-size:13px;outline:none;transition:border-color .15s;}
input[type=password]:focus{border-color:var(--accent);}
.remember{display:flex;align-items:center;gap:8px;margin-top:12px;font-size:12px;color:var(--muted);cursor:pointer;}
.remember input{width:14px;height:14px;accent-color:var(--accent);cursor:pointer;}
button{width:100%;margin-top:20px;background:var(--accent);border:none;border-radius:8px;padding:12px;color:#fff;font-family:'Inter',sans-serif;font-size:14px;font-weight:600;cursor:pointer;transition:opacity .15s;}
button:hover{opacity:.85;}
</style>
</head>
<body>
<div class="card">
  <div class="logo">onyx<span>.</span></div>
  <div class="sub">Dashboard -- sign in</div>
  <form method="POST" action="/login">
    <label for="pw">Password</label>
    <input type="password" id="pw" name="password" placeholder="&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;" autofocus autocomplete="current-password">
    <label class="remember"><input type="checkbox" name="remember_me"> Keep me signed in for 7 days</label>
    <button type="submit">Sign in &#8594;</button>
  </form>
</div>
</body>
</html>`

const loginErrorHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Onyx -- Sign in</title>
<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,<svg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 32 32'><rect width='32' height='32' rx='8' fill='%237c6af7'/><text y='24' x='4' font-size='22' font-family='monospace' fill='white'>&#9670;</text></svg>">
<style>
@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;600&family=Inter:wght@400;500;600&display=swap');
:root{--bg:#0a0a0f;--surface:#0f0f18;--border:#1c1c2e;--accent:#7c6af7;--text:#e4e4ef;--muted:#555570;--red:#ff5370;}
*{box-sizing:border-box;margin:0;padding:0;}
body{background:var(--bg);color:var(--text);font-family:'Inter',sans-serif;min-height:100vh;display:grid;place-items:center;}
.card{background:var(--surface);border:1px solid var(--border);border-radius:14px;padding:36px 40px;width:360px;}
.logo{font-family:'JetBrains Mono',monospace;font-size:22px;font-weight:600;color:var(--accent);margin-bottom:4px;}
.logo span{color:#3ddc97;}
.sub{color:var(--muted);font-size:13px;margin-bottom:16px;}
.err{background:rgba(255,83,112,.1);border:1px solid rgba(255,83,112,.3);border-radius:8px;padding:9px 14px;color:var(--red);font-size:12px;margin-bottom:18px;}
label{display:block;font-size:11px;color:var(--muted);text-transform:uppercase;letter-spacing:.07em;margin-bottom:6px;}
input[type=password]{width:100%;background:var(--bg);border:1px solid rgba(255,83,112,.4);border-radius:8px;padding:11px 14px;color:var(--text);font-family:'JetBrains Mono',monospace;font-size:13px;outline:none;}
.remember{display:flex;align-items:center;gap:8px;margin-top:12px;font-size:12px;color:var(--muted);cursor:pointer;}
.remember input{width:14px;height:14px;accent-color:var(--accent);cursor:pointer;}
button{width:100%;margin-top:20px;background:var(--accent);border:none;border-radius:8px;padding:12px;color:#fff;font-family:'Inter',sans-serif;font-size:14px;font-weight:600;cursor:pointer;}
</style>
</head>
<body>
<div class="card">
  <div class="logo">onyx<span>.</span></div>
  <div class="sub">Dashboard -- sign in</div>
  <div class="err">&#9888; Incorrect password. Please try again.</div>
  <form method="POST" action="/login">
    <label for="pw">Password</label>
    <input type="password" id="pw" name="password" placeholder="&bull;&bull;&bull;&bull;&bull;&bull;&bull;&bull;" autofocus autocomplete="current-password">
    <label class="remember"><input type="checkbox" name="remember_me"> Keep me signed in for 7 days</label>
    <button type="submit">Sign in &#8594;</button>
  </form>
</div>
</body>
</html>`

const loginRateLimitHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Onyx -- Too Many Attempts</title>
<style>
*{box-sizing:border-box;margin:0;padding:0;}
body{background:#0a0a0f;color:#e4e4ef;font-family:Inter,sans-serif;min-height:100vh;display:grid;place-items:center;}
.card{background:#0f0f18;border:1px solid #1c1c2e;border-radius:14px;padding:36px 40px;width:360px;text-align:center;}
.icon{font-size:36px;margin-bottom:12px;}
h2{font-size:16px;margin-bottom:8px;}
p{color:#555570;font-size:13px;line-height:1.6;}
a{color:#7c6af7;text-decoration:none;}
</style>
</head>
<body>
<div class="card">
  <div class="icon">&#128274;</div>
  <h2>Too Many Attempts</h2>
  <p>Too many failed login attempts from your IP.<br>Please wait 60 seconds before trying again.</p>
  <p style="margin-top:16px"><a href="/login">&#8592; Back to login</a></p>
</div>
</body>
</html>`
