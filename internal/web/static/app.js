const SVGNS = "http://www.w3.org/2000/svg";
const parisFmt = new Intl.DateTimeFormat("fr-FR", {
  timeZone: "Europe/Paris", hour: "2-digit", minute: "2-digit", hour12: false,
});
const parisDayFmt = new Intl.DateTimeFormat("en-CA", {
  timeZone: "Europe/Paris", year: "numeric", month: "2-digit", day: "2-digit",
});

const COLORS = { green: "#16a34a", orange: "#f97316", red: "#dc2626", gray: "#94a3b8" };

let currentMetric = "bikes";
let currentData = null;

const map = new maplibregl.Map({
  container: "map",
  style: "https://tiles.openfreemap.org/styles/liberty",
  center: [1.7, 47.3],
  zoom: 5,
});
map.addControl(new maplibregl.NavigationControl({ showCompass: false }), "top-left");

function availabilityState(count, capacity) {
  if (count === null || count === undefined) return "gray";
  if (count === 0) return "red";
  const ratio = capacity > 0 ? count / capacity : 1;
  return ratio < 0.2 ? "orange" : "green";
}

function stateClass(count, capacity) {
  return { green: "g", orange: "o", red: "r", gray: "x" }[availabilityState(count, capacity)];
}

function popupHTML(p) {
  const bikes = p.bikes === "" ? "?" : p.bikes;
  const docks = p.docks === "" ? "?" : p.docks;
  return `<div class="popup-title">${p.name}</div>
    <div class="popup-city">${p.city}</div>
    <div class="popup-badges">
      <span class="badge ${p.bikesClass}"><span class="badge-ico">🚲</span> ${bikes}</span>
      <span class="badge ${p.docksClass}"><span class="badge-ico">🅿️</span> ${docks}</span>
    </div>`;
}

async function loadMap() {
  const stations = await fetch("/map").then((r) => r.json());

  const features = stations.map((s) => ({
    type: "Feature",
    geometry: { type: "Point", coordinates: [s.lon, s.lat] },
    properties: {
      id: s.id,
      name: s.name,
      city: s.city,
      bikes: s.bikes_available ?? "",
      docks: s.docks_available ?? "",
      bikesClass: stateClass(s.bikes_available, s.capacity),
      docksClass: stateClass(s.docks_available, s.capacity),
      state: availabilityState(s.bikes_available, s.capacity),
    },
  }));

  map.addSource("stations", { type: "geojson", data: { type: "FeatureCollection", features } });
  map.addLayer({
    id: "stations",
    type: "circle",
    source: "stations",
    paint: {
      "circle-color": ["match", ["get", "state"],
        "green", COLORS.green, "orange", COLORS.orange, "red", COLORS.red, COLORS.gray],
      "circle-radius": ["interpolate", ["linear"], ["zoom"], 5, 3, 11, 5, 15, 9],
      "circle-stroke-width": 1,
      "circle-stroke-color": "rgba(255,255,255,0.65)",
      "circle-opacity": 0.9,
    },
  });

  const popup = new maplibregl.Popup({ closeButton: false, closeOnClick: false, offset: 10 });
  map.on("mouseenter", "stations", (e) => {
    map.getCanvas().style.cursor = "pointer";
    const f = e.features[0];
    popup.setLngLat(f.geometry.coordinates).setHTML(popupHTML(f.properties)).addTo(map);
  });
  map.on("mouseleave", "stations", () => {
    map.getCanvas().style.cursor = "";
    popup.remove();
  });
  map.on("click", "stations", (e) => selectStation(e.features[0].properties.id));

  const bounds = new maplibregl.LngLatBounds();
  for (const f of features) bounds.extend(f.geometry.coordinates);
  if (!bounds.isEmpty()) map.fitBounds(bounds, { padding: 40, animate: false });
}

map.on("load", loadMap);

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
      <div class="stat"><div class="top">🚲 Vélos disponibles</div><div class="n">${st.bikes_available}</div></div>
      <div class="stat"><div class="top">🅿️ Bornes libres</div><div class="n">${st.docks_available}</div></div>`;
  } else {
    live.innerHTML = `<div class="stat"><div class="top">Aucune disponibilité relevée pour l'instant.</div></div>`;
  }

  const totalSamples = prediction.profile.reduce((a, s) => a + s.samples, 0);
  document.getElementById("d-samples").textContent =
    `Prédiction : moyenne par tranche de 30 min, ${totalSamples} relevé(s) pour ce jour de semaine.`;

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
  const today = parisDayFmt.format(new Date());
  const realPts = history.data
    .filter((d) => parisDayFmt.format(new Date(d.time)) === today)
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
  svg.appendChild(text(Math.min(nx + 4, W - padR - 60), padT + 10, "Maintenant", "now-label", "start"));

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

  attachHover({ svg, el, realPts, predPts, xScale, yScale, W, H, padL, padR, padT, padB });
}

function nearestByX(pts, mx) {
  let best = null, bd = Infinity;
  for (const p of pts) {
    const d = Math.abs(p.x - mx);
    if (d < bd) { bd = d; best = p; }
  }
  return best;
}

function attachHover(ctx) {
  const { svg, el, realPts, predPts, xScale, yScale, W, H, padL, padR, padT, padB } = ctx;
  if (!realPts.length && !predPts.length) return;

  const hover = el("g", { class: "chart-hover", visibility: "hidden" });
  const vline = el("line", { class: "hover-line", y1: padT, y2: H - padB });
  const dotReal = el("circle", { class: "hover-dot-real", r: 4.5, visibility: "hidden" });
  const dotPred = el("circle", { class: "hover-dot-pred", r: 4.5, visibility: "hidden" });
  const box = el("rect", { class: "chart-tip-box", rx: 8 });
  const t1 = el("text", { class: "chart-tip-time" });
  const t2 = el("text", { class: "chart-tip-val" });
  const t3 = el("text", { class: "chart-tip-val" });
  hover.append(vline, dotReal, dotPred, box, t1, t2, t3);
  svg.appendChild(hover);

  const capture = el("rect", {
    class: "hover-capture",
    x: padL, y: padT, width: W - padL - padR, height: H - padT - padB, fill: "transparent",
  });
  svg.appendChild(capture);

  const cursorMinutes = (evt) => {
    const pt = svg.createSVGPoint();
    pt.x = evt.clientX;
    pt.y = evt.clientY;
    const loc = pt.matrixTransform(svg.getScreenCTM().inverse());
    return ((loc.x - padL) / (W - padL - padR)) * 1440;
  };

  capture.addEventListener("mousemove", (evt) => {
    const m = cursorMinutes(evt);
    const anchor = realPts.length ? nearestByX(realPts, m) : nearestByX(predPts, m);
    if (!anchor) return;

    const ax = anchor.x;
    const appx = xScale(ax);
    const rp = realPts.length ? nearestByX(realPts, ax) : null;
    const pp = predPts.length ? nearestByX(predPts, ax) : null;
    const showReal = rp && Math.abs(rp.x - ax) <= 30;
    const showPred = pp && Math.abs(pp.x - ax) <= 30;

    vline.setAttribute("x1", appx);
    vline.setAttribute("x2", appx);

    if (showReal) {
      dotReal.setAttribute("cx", appx);
      dotReal.setAttribute("cy", yScale(rp.y));
      dotReal.setAttribute("visibility", "visible");
    } else {
      dotReal.setAttribute("visibility", "hidden");
    }
    if (showPred) {
      dotPred.setAttribute("cx", appx);
      dotPred.setAttribute("cy", yScale(pp.y));
      dotPred.setAttribute("visibility", "visible");
    } else {
      dotPred.setAttribute("visibility", "hidden");
    }

    const unit = currentMetric === "bikes" ? "vélos" : "bornes";
    const hh = String(Math.floor(ax / 60)).padStart(2, "0");
    const mm = String(Math.round(ax) % 60).padStart(2, "0");
    t1.textContent = `${hh}:${mm}`;
    t2.textContent = showReal ? `Réel : ${rp.y} ${unit}` : "";
    t3.textContent = showPred ? `Typique : ${Math.round(pp.y)} ${unit}` : "";

    const lines = [t2, t3].filter((t) => t.textContent);
    const boxW = 148, lineH = 16, padIn = 8;
    const boxH = padIn * 2 + 15 + lines.length * lineH;
    let bx = appx + 12;
    if (bx + boxW > W - padR) bx = appx - 12 - boxW;
    const anchorY = showReal ? yScale(rp.y) : yScale(pp.y);
    let by = anchorY - boxH - 10;
    if (by < padT) by = anchorY + 12;
    by = Math.max(padT, Math.min(by, H - padB - boxH));

    box.setAttribute("x", bx);
    box.setAttribute("y", by);
    box.setAttribute("width", boxW);
    box.setAttribute("height", boxH);
    t1.setAttribute("x", bx + padIn);
    t1.setAttribute("y", by + padIn + 12);
    lines.forEach((t, i) => {
      t.setAttribute("x", bx + padIn);
      t.setAttribute("y", by + padIn + 15 + lineH * (i + 1));
    });

    hover.setAttribute("visibility", "visible");
  });

  capture.addEventListener("mouseleave", () => {
    hover.setAttribute("visibility", "hidden");
  });
}

document.querySelectorAll(".toggle button").forEach((btn) => {
  btn.addEventListener("click", () => {
    document.querySelectorAll(".toggle button").forEach((b) => b.classList.remove("active"));
    btn.classList.add("active");
    currentMetric = btn.dataset.metric;
    if (currentData) drawChart();
  });
});
