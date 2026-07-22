const SVGNS = "http://www.w3.org/2000/svg";
const parisFmt = new Intl.DateTimeFormat("fr-FR", {
  timeZone: "Europe/Paris", hour: "2-digit", minute: "2-digit", hour12: false,
});

let currentMetric = "bikes";
let currentData = null;

const map = L.map("map", { preferCanvas: true }).setView([47.0, 0.5], 6);
L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
  maxZoom: 19,
  attribution: "&copy; OpenStreetMap",
}).addTo(map);

function color(bikes, capacity) {
  if (bikes === null || bikes === undefined) return "#9ca3af";
  const ratio = capacity > 0 ? bikes / capacity : (bikes > 0 ? 1 : 0);
  if (bikes === 0) return "#dc2626";
  if (ratio < 0.2) return "#f97316";
  return "#16a34a";
}

async function loadMap() {
  const res = await fetch("/map");
  const stations = await res.json();
  const bounds = [];
  for (const s of stations) {
    const m = L.circleMarker([s.lat, s.lon], {
      radius: 5,
      color: color(s.bikes_available, s.capacity),
      fillColor: color(s.bikes_available, s.capacity),
      fillOpacity: 0.85,
      weight: 1,
    }).addTo(map);
    m.bindTooltip(`${s.name} (${s.city})`);
    m.on("click", () => selectStation(s.id));
    bounds.push([s.lat, s.lon]);
  }
  if (bounds.length) map.fitBounds(bounds, { padding: [30, 30] });
}

function parisMinutes(iso) {
  const parts = parisFmt.formatToParts(new Date(iso));
  let h = 0, m = 0;
  for (const p of parts) {
    if (p.type === "hour") h = parseInt(p.value, 10);
    if (p.type === "minute") m = parseInt(p.value, 10);
  }
  return h * 60 + m;
}

async function selectStation(id) {
  const [station, history, prediction] = await Promise.all([
    fetch(`/stations/${id}`).then((r) => r.json()),
    fetch(`/stations/${id}/history`).then((r) => r.json()),
    fetch(`/stations/${id}/prediction`).then((r) => r.json()),
  ]);
  currentData = { station, history, prediction };
  renderDetail();
}

function renderDetail() {
  const { station, prediction } = currentData;
  document.getElementById("detail").hidden = false;
  document.getElementById("d-name").textContent = station.name;
  document.getElementById("d-city").textContent = station.city;

  const st = station.status;
  const live = document.getElementById("d-live");
  if (st) {
    live.innerHTML = `
      <div class="stat"><div class="n">${st.bikes_available}</div><div class="l">vélos dispo</div></div>
      <div class="stat"><div class="n">${st.docks_available}</div><div class="l">bornes libres</div></div>`;
  } else {
    live.innerHTML = `<div class="stat"><div class="l">Aucune dispo relevée pour l'instant.</div></div>`;
  }

  const totalSamples = prediction.profile.reduce((a, s) => a + s.samples, 0);
  document.getElementById("d-samples").textContent =
    `Prédiction : moyenne par tranche de 30 min (${totalSamples} relevés ce jour de semaine).`;

  drawChart();
}

function drawChart() {
  const { history, prediction } = currentData;
  const svg = document.getElementById("chart");
  while (svg.firstChild) svg.removeChild(svg.firstChild);

  const W = 640, H = 260, padL = 34, padR = 12, padT = 12, padB = 24;
  const key = currentMetric === "bikes" ? "bikes_available" : "docks_available";
  const predKey = currentMetric === "bikes" ? "avg_bikes" : "avg_docks";

  const predPts = prediction.profile.map((s) => ({ x: s.minutes, y: s[predKey] }));
  const realPts = history.data
    .map((d) => ({ x: parisMinutes(d.time), y: d[key] }))
    .sort((a, b) => a.x - b.x);

  let maxY = 1;
  for (const p of predPts) maxY = Math.max(maxY, p.y);
  for (const p of realPts) maxY = Math.max(maxY, p.y);
  maxY = Math.ceil(maxY * 1.1);

  const xScale = (m) => padL + (m / 1440) * (W - padL - padR);
  const yScale = (v) => H - padB - (v / maxY) * (H - padT - padB);

  const line = (cls) => { const el = document.createElementNS(SVGNS, "line"); el.setAttribute("class", cls); return el; };
  const text = (x, y, s, anchor) => {
    const el = document.createElementNS(SVGNS, "text");
    el.setAttribute("x", x); el.setAttribute("y", y); el.setAttribute("class", "tick");
    if (anchor) el.setAttribute("text-anchor", anchor);
    el.textContent = s; return el;
  };

  const axisX = line("axis");
  axisX.setAttribute("x1", padL); axisX.setAttribute("y1", H - padB);
  axisX.setAttribute("x2", W - padR); axisX.setAttribute("y2", H - padB);
  svg.appendChild(axisX);

  for (let h = 0; h <= 24; h += 6) {
    const x = xScale(h * 60);
    const g = line("grid");
    g.setAttribute("x1", x); g.setAttribute("y1", padT);
    g.setAttribute("x2", x); g.setAttribute("y2", H - padB);
    svg.appendChild(g);
    svg.appendChild(text(x, H - padB + 14, `${h}h`, "middle"));
  }
  svg.appendChild(text(padL - 6, yScale(maxY) + 4, String(maxY), "end"));
  svg.appendChild(text(padL - 6, yScale(0) + 4, "0", "end"));

  const polyline = (pts, cls) => {
    if (!pts.length) return;
    const el = document.createElementNS(SVGNS, "polyline");
    el.setAttribute("class", cls);
    el.setAttribute("points", pts.map((p) => `${xScale(p.x)},${yScale(p.y)}`).join(" "));
    svg.appendChild(el);
  };
  polyline(predPts.filter((p) => p.y > 0 || prediction.profile[p.x / 30].samples > 0), "serie-pred");
  polyline(realPts, "serie-real");

  svg.appendChild(text(padL, padT + 8, "— réel", "start")).setAttribute("fill", "var(--real)");
  const legPred = text(padL + 60, padT + 8, "-- typique", "start");
  legPred.setAttribute("fill", "var(--pred)");
  svg.appendChild(legPred);
}

document.querySelectorAll(".toggle button").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".toggle button").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    currentMetric = btn.dataset.metric;
    if (currentData) drawChart();
  });
});

loadMap();
