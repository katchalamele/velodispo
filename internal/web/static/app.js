const SVGNS = "http://www.w3.org/2000/svg";
const parisFmt = new Intl.DateTimeFormat("fr-FR", {
  timeZone: "Europe/Paris", hour: "2-digit", minute: "2-digit", hour12: false,
});

const COLORS = { green: "#16a34a", orange: "#f97316", red: "#dc2626", gray: "#94a3b8" };

let currentMetric = "bikes";
let currentData = null;

const map = L.map("map", { preferCanvas: true, zoomControl: true }).setView([47.5, 1.5], 6);
L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
  maxZoom: 19,
  attribution: "&copy; OpenStreetMap",
}).addTo(map);

function color(bikes, capacity) {
  if (bikes === null || bikes === undefined) return COLORS.gray;
  if (bikes === 0) return COLORS.red;
  const ratio = capacity > 0 ? bikes / capacity : 1;
  if (ratio < 0.2) return COLORS.orange;
  return COLORS.green;
}

async function loadMap() {
  const stations = await fetch("/map").then((r) => r.json());
  const bounds = [];
  for (const s of stations) {
    const c = color(s.bikes_available, s.capacity);
    const m = L.circleMarker([s.lat, s.lon], {
      radius: 5, color: c, fillColor: c, fillOpacity: 0.85, weight: 1,
    }).addTo(map);
    m.bindTooltip(`${s.name} — ${s.city}`);
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
  document.getElementById("empty").hidden = true;
  document.getElementById("detail").hidden = false;
  document.getElementById("d-name").textContent = station.name;
  document.getElementById("d-city").textContent = station.city;

  const st = station.status;
  const live = document.getElementById("d-live");
  if (st) {
    live.innerHTML = `
      <div class="stat"><div class="top">🚲 vélos disponibles</div><div class="n">${st.bikes_available}</div></div>
      <div class="stat"><div class="top">🅿️ bornes libres</div><div class="n">${st.docks_available}</div></div>`;
  } else {
    live.innerHTML = `<div class="stat"><div class="top">Aucune disponibilité relevée pour l'instant.</div></div>`;
  }

  const totalSamples = prediction.profile.reduce((a, s) => a + s.samples, 0);
  document.getElementById("d-samples").textContent =
    `Prédiction : moyenne par tranche de 30 min — ${totalSamples} relevé(s) pour ce jour de semaine.`;

  drawChart();
}

function drawChart() {
  const { history, prediction } = currentData;
  const svg = document.getElementById("chart");
  while (svg.firstChild) svg.removeChild(svg.firstChild);

  const W = 760, H = 380, padL = 44, padR = 18, padT = 18, padB = 34;
  const key = currentMetric === "bikes" ? "bikes_available" : "docks_available";
  const predKey = currentMetric === "bikes" ? "avg_bikes" : "avg_docks";

  const predPts = prediction.profile
    .filter((s) => s.samples > 0)
    .map((s) => ({ x: s.minutes, y: s[predKey] }));
  const realPts = history.data
    .map((d) => ({ x: parisMinutes(d.time), y: d[key] }))
    .sort((a, b) => a.x - b.x);

  let maxY = 1;
  for (const p of predPts) maxY = Math.max(maxY, p.y);
  for (const p of realPts) maxY = Math.max(maxY, p.y);
  maxY = Math.ceil(maxY * 1.15);

  const xScale = (m) => padL + (m / 1440) * (W - padL - padR);
  const yScale = (v) => H - padB - (v / maxY) * (H - padT - padB);

  const el = (name, attrs) => {
    const e = document.createElementNS(SVGNS, name);
    for (const k in attrs) e.setAttribute(k, attrs[k]);
    return e;
  };
  const text = (x, y, s, cls, anchor) => {
    const t = el("text", { x, y, class: cls || "tick" });
    if (anchor) t.setAttribute("text-anchor", anchor);
    t.textContent = s;
    return t;
  };

  // grille horizontale + graduations Y
  for (const frac of [0, 0.5, 1]) {
    const v = Math.round(maxY * frac);
    const y = yScale(v);
    svg.appendChild(el("line", { class: "grid", x1: padL, y1: y, x2: W - padR, y2: y }));
    svg.appendChild(text(padL - 8, y + 4, String(v), "tick", "end"));
  }

  // axe X + graduations toutes les 3h
  svg.appendChild(el("line", { class: "axis", x1: padL, y1: H - padB, x2: W - padR, y2: H - padB }));
  for (let h = 0; h <= 24; h += 3) {
    const x = xScale(h * 60);
    svg.appendChild(text(x, H - padB + 18, `${h}h`, "tick", "middle"));
  }

  // aire + ligne de prédiction
  if (predPts.length) {
    const base = yScale(0);
    const area = predPts.map((p) => `${xScale(p.x)},${yScale(p.y)}`).join(" ");
    svg.appendChild(el("polygon", {
      class: "serie-pred-area",
      points: `${xScale(predPts[0].x)},${base} ${area} ${xScale(predPts[predPts.length - 1].x)},${base}`,
    }));
    svg.appendChild(el("polyline", { class: "serie-pred", points: area }));
  }

  // repère "maintenant"
  const nowM = parisMinutes(new Date().toISOString());
  const nx = xScale(nowM);
  svg.appendChild(el("line", { class: "now-line", x1: nx, y1: padT, x2: nx, y2: H - padB }));
  svg.appendChild(text(Math.min(nx + 4, W - padR - 60), padT + 10, "maintenant", "now-label", "start"));

  // ligne réelle + points
  if (realPts.length) {
    svg.appendChild(el("polyline", {
      class: "serie-real",
      points: realPts.map((p) => `${xScale(p.x)},${yScale(p.y)}`).join(" "),
    }));
    for (const p of realPts) {
      svg.appendChild(el("circle", { class: "serie-real-dot", cx: xScale(p.x), cy: yScale(p.y), r: 3 }));
    }
  }
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
