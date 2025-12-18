<!doctype html>
<html lang="zh-Hant">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>健走追蹤器 — 跨平台健康資料整合</title>
  <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" crossorigin=""/>
  <style>
    :root{--bg:#f7fafc;--card:#ffffff;--accent:#1e88e5}
    body{font-family:system-ui,-apple-system,Segoe UI,Roboto; margin:0; background:var(--bg)}
    header{background:#fff;padding:14px 18px;box-shadow:0 1px 0 rgba(0,0,0,.05)}
    .container{max-width:1200px;margin:16px auto;padding:12px}
    .layout{display:grid;grid-template-columns:1fr 420px;gap:14px}
    .card{background:var(--card);border-radius:10px;padding:12px;box-shadow:0 6px 18px rgba(0,0,0,.06)}
    #map{height:520px;border-radius:8px}
    button{background:var(--accent);color:#fff;border:none;padding:8px 12px;border-radius:8px;cursor:pointer}
    button.alt{background:#4caf50}
    .muted{color:#666;font-size:13px}
    .row{display:flex;gap:8px;align-items:center}
    .metric{background:#f4f8fb;padding:8px;border-radius:8px;flex:1}
    .metric .label{font-size:12px;color:#666}
    .metric .value{font-weight:700;font-size:18px}
    .lifebar{height:20px;background:#eee;border-radius:10px;overflow:hidden}
    .lifebar i{display:block;height:100%;width:0;background:linear-gradient(90deg,#4caf50,#81c784)}
    footer{margin-top:12px;font-size:13px;color:#666}
    pre{background:#f3f3f3;padding:8px;border-radius:6px;font-size:12px;overflow:auto}
  </style>
</head>
<body>
<header>
  <div class="container">
    <h2>健走追蹤器（Health Connect / Apple Health 整合版）</h2>
    <div class="muted">Android 透過 Health Connect、iPhone 透過 Apple Health 匯入「非即時」健康資料</div>
  </div>
</header>

<main class="container">
  <div class="layout">
    <section class="card">
      <div id="map"></div>
      <div class="row" style="justify-content:space-between;margin-top:8px">
        <div class="muted">GPS 精準度：<span id="accuracy">-</span> m</div>
        <div class="row">
          <button id="startBtn">開始</button>
          <button id="stopBtn" class="alt" disabled>結束</button>
        </div>
      </div>
    </section>

    <aside class="card">
      <div class="row">
        <div class="metric">
          <div class="label">本次時間</div>
          <div class="value" id="duration">00:00:00</div>
        </div>
        <div class="metric">
          <div class="label">距離 (km)</div>
          <div class="value" id="distance">0.00</div>
        </div>
      </div>

      <div style="margin-top:10px">
        <div class="label">每日步數目標</div>
        <input id="goalSteps" type="number" value="8000" style="width:100%">
        <div class="lifebar" style="margin-top:6px"><i id="lifeSteps"></i></div>
        <div class="muted" id="lifeStepsText">0%</div>
      </div>

      <!-- 新增：健康資料匯入 -->
      <div style="margin-top:14px">
        <div class="label">同步手機健康資料</div>
        <div class="row">
          <button id="importHealthBtn">模擬同步（Demo）</button>
          <div class="muted" id="healthStatus">尚未同步</div>
        </div>
        <div class="muted" style="margin-top:6px">實際環境：由 Android / iOS App 呼叫 applyHealthData()</div>
      </div>

      <div style="margin-top:12px">
        <div class="label">最近匯入資料（JSON）</div>
        <pre id="healthPreview">尚無</pre>
      </div>
    </aside>
  </div>
</main>

<footer class="container">
  本頁面不直接存取 Health Connect / Apple Health，僅接收已授權後整理完成的資料。
</footer>

<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" crossorigin=""></script>
<script>
// ---------------- 地圖與 GPS（原功能保留） ----------------
const map = L.map('map');
let marker, polyline, watchId=null, startTime=null, timer=null, path=[];
L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png',{maxZoom:19}).addTo(map);
if(navigator.geolocation){navigator.geolocation.getCurrentPosition(p=>map.setView([p.coords.latitude,p.coords.longitude],16),()=>map.setView([25.03,121.56],13));}

function onPos(p){const c=p.coords;document.getElementById('accuracy').innerText=Math.round(c.accuracy);const ll=[c.latitude,c.longitude];if(!marker)marker=L.marker(ll).addTo(map);else marker.setLatLng(ll);if(!polyline)polyline=L.polyline([],{}).addTo(map);polyline.addLatLng(ll);path.push(ll);} 
function start(){path=[];if(polyline){map.removeLayer(polyline);polyline=null;}watchId=navigator.geolocation.watchPosition(onPos);startTime=Date.now();timer=setInterval(updateStats,1000);}
function stop(){navigator.geolocation.clearWatch(watchId);clearInterval(timer);}
function updateStats(){const sec=Math.floor((Date.now()-startTime)/1000);document.getElementById('duration').innerText=new Date(sec*1000).toISOString().substr(11,8);}

// ---------------- 新增：健康資料整合 ----------------

// 統一入口（Android / iOS App 會呼叫）
function applyHealthData(payload){
  // 儲存
  localStorage.setItem('external_health_'+payload.date, JSON.stringify(payload));

  // UI 更新
  document.getElementById('healthStatus').innerText = '已同步：' + payload.platform;
  document.getElementById('healthPreview').innerText = JSON.stringify(payload, null, 2);

  // 更新生命條（以外部步數為主）
  updateLifeBar(payload.steps);
}

function updateLifeBar(steps){
  const goal = parseInt(document.getElementById('goalSteps').value)||8000;
  const pct = Math.min(100, Math.round(steps/goal*100));
  document.getElementById('lifeSteps').style.width = pct+'%';
  document.getElementById('lifeStepsText').innerText = pct+'%';
}

// Demo：模擬 Android / iOS App 傳資料
function demoImport(){
  const demo = {
    platform: Math.random()>0.5 ? 'android' : 'ios',
    date: new Date().toISOString().slice(0,10),
    steps: 8650,
    distanceKm: 6.2,
    calories: 315,
    workouts:[{type:'walking',durationMin:32,distanceKm:2.5,avgHeartRate:110}]
  };
  applyHealthData(demo);
}

// ---------------- 綁定事件 ----------------
document.getElementById('startBtn').onclick=start;
document.getElementById('stopBtn').onclick=stop;
document.getElementById('importHealthBtn').onclick=demoImport;
</script>
</body>
</html>
