const SPEED_COLORS = {
  50: "#22c55e",
  80: "#eab308",
  100: "#f97316",
  110: "#ef4444",
};

function getSpeedColor(speedKmh) {
  if (speedKmh <= 50) return SPEED_COLORS[50];
  if (speedKmh <= 80) return SPEED_COLORS[80];
  if (speedKmh <= 100) return SPEED_COLORS[100];
  return SPEED_COLORS[110];
}

let overlayCanvas = null;
let lastBbox = null;
let segmentData = null;

function init() {
  waitForMap().then(() => {
    createOverlay();
    observeMapChanges();
    fetchAndRender();
  });
}

function waitForMap() {
  return new Promise((resolve) => {
    const check = () => {
      const canvas = document.querySelector("canvas.widget-scene-canvas");
      if (canvas) {
        resolve(canvas);
      } else {
        setTimeout(check, 1000);
      }
    };
    check();
  });
}

function createOverlay() {
  overlayCanvas = document.createElement("div");
  overlayCanvas.id = "ke-speed-overlay";
  overlayCanvas.style.cssText =
    "position:absolute;top:0;left:0;width:100%;height:100%;pointer-events:none;z-index:1000;";

  const mapContainer = document.querySelector("#scene");
  if (mapContainer) {
    mapContainer.appendChild(overlayCanvas);
  }
}

function getCurrentBBox() {
  const url = new URL(window.location.href);
  const match = url.pathname.match(/@(-?\d+\.?\d*),(-?\d+\.?\d*),(\d+\.?\d*)z/);

  if (!match) return null;

  const lat = parseFloat(match[1]);
  const lng = parseFloat(match[2]);
  const zoom = parseFloat(match[3]);

  const spread = 180 / Math.pow(2, zoom);

  return {
    minLat: lat - spread,
    minLng: lng - spread * 1.5,
    maxLat: lat + spread,
    maxLng: lng + spread * 1.5,
  };
}

function observeMapChanges() {
  let debounceTimer;

  const observer = new MutationObserver(() => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(fetchAndRender, 500);
  });

  const target = document.querySelector("#scene");
  if (target) {
    observer.observe(target, { subtree: true, childList: true, attributes: true });
  }

  window.addEventListener("popstate", () => {
    clearTimeout(debounceTimer);
    debounceTimer = setTimeout(fetchAndRender, 300);
  });
}

function fetchAndRender() {
  const bbox = getCurrentBBox();
  if (!bbox) return;

  if (lastBbox && JSON.stringify(bbox) === JSON.stringify(lastBbox)) return;
  lastBbox = bbox;

  chrome.runtime.sendMessage({ type: "FETCH_SPEEDS", bbox }, (response) => {
    if (response && response.success) {
      segmentData = response.data;
      renderSpeedInfo(response.data);
    }
  });
}

function renderSpeedInfo(data) {
  if (!overlayCanvas) return;

  overlayCanvas.innerHTML = "";

  if (!data.features || data.features.length === 0) return;

  const legend = document.createElement("div");
  legend.className = "ke-speed-legend";
  legend.innerHTML = `
    <div class="ke-speed-legend-title">Speed Limits (km/h)</div>
    <div class="ke-speed-legend-item"><span style="background:${SPEED_COLORS[50]}"></span> ≤ 50</div>
    <div class="ke-speed-legend-item"><span style="background:${SPEED_COLORS[80]}"></span> 51–80</div>
    <div class="ke-speed-legend-item"><span style="background:${SPEED_COLORS[100]}"></span> 81–100</div>
    <div class="ke-speed-legend-item"><span style="background:${SPEED_COLORS[110]}"></span> 101–110</div>
    <div class="ke-speed-count">${data.features.length} segment(s)</div>
  `;
  legend.style.pointerEvents = "auto";
  overlayCanvas.appendChild(legend);

  const list = document.createElement("div");
  list.className = "ke-speed-list";
  list.style.pointerEvents = "auto";

  data.features.forEach((feature) => {
    const props = feature.properties;
    const color = getSpeedColor(props.speed_limit_kmh);

    const item = document.createElement("div");
    item.className = "ke-speed-item";
    item.innerHTML = `
      <span class="ke-speed-dot" style="background:${color}"></span>
      <span class="ke-speed-name">${props.road_name}</span>
      <span class="ke-speed-value">${props.speed_limit_kmh} km/h</span>
    `;
    list.appendChild(item);
  });

  overlayCanvas.appendChild(list);
}

init();
