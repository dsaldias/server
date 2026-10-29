import { graphql, queries } from "./api.js";
import { esc, toast } from "./components.js";

let socket;
let reconnectTimer;
let stopped = false;

function setStatus(status) {
  const labels = { connecting: "Conectando", connected: "Conectado", error: "Desconectado" };
  const colors = { connecting: "#f59e0b", connected: "#22c55e", error: "#ef4444" };
  document.querySelectorAll("[data-realtime-dot]").forEach((dot) => {
    dot.style.backgroundColor = colors[status] || colors.error;
    dot.title = labels[status] || labels.error;
    dot.setAttribute("aria-label", labels[status] || labels.error);
  });
  document.querySelectorAll("[data-realtime-label]").forEach((node) => { node.textContent = labels[status] || labels.error; });
  window.dispatchEvent(new CustomEvent("auth-v2-realtime-status", { detail: { status } }));
}

function updateConnectionCounters(data) {
  const connected = data?.datos?.conectados;
  const total = data?.datos?.total_conectados;
  document.querySelectorAll("[data-realtime-users]").forEach((node) => { node.textContent = String(connected ?? 0); });
  document.querySelectorAll("[data-realtime-users-detail]").forEach((node) => { node.textContent = `${connected ?? 0} usuarios · ${total ?? 0} conexiones`; });
}

function checkAssetVersion(version) {
  if (!version) return;
  const current = window.__AUTH_ASSET_VERSION__ || localStorage.getItem("auth_asset_version");
  if (!current) { localStorage.setItem("auth_asset_version", version); return; }
  if (current === version || window.__authV2VersionUpdateShown) return;
  window.__authV2VersionUpdateShown = true;
  const node = document.createElement("button");
  node.className = "toast";
  node.type = "button";
  node.textContent = "Hay una nueva versión disponible. Haz clic para actualizar.";
  node.addEventListener("click", () => { localStorage.setItem("auth_asset_version", version); window.location.reload(); });
  document.querySelector("#toast-region")?.append(node);
}

function websocketURL() {
  const apiURL = window.__AUTH_API_URL__ || `${location.protocol}//${location.host}/query_auth`;
  const url = new URL(apiURL, location.href);
  url.pathname = "/ws";
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  return url.toString();
}

function scheduleReconnect() {
  if (stopped || reconnectTimer) return;
  reconnectTimer = setTimeout(() => { reconnectTimer = null; connect(); }, 2000);
}

function handleNotification(payload) {
  const notification = payload?.data?.notificaciones_subs;
  if (!notification) return;
  let detail = notification;
  if (notification.data_json) {
    try { detail = { ...notification, data: JSON.parse(notification.data_json) }; } catch { /* keep the base notification */ }
  }
  window.dispatchEvent(new CustomEvent("auth-v2-notification", { detail }));
  const type = detail.data?.tipo;
  if (type === "conectados") {
    updateConnectionCounters(detail.data);
    checkAssetVersion(detail.data?.datos?.assets_version);
  } else if (type === "chat") {
    toast(notification.title || "Nuevo mensaje de chat");
  } else if (type === "alerta" || !type) {
    toast(notification.title || "Nueva notificación");
  }
}

function connect() {
  if (stopped) return;
  setStatus("connecting");
  try { socket = new WebSocket(websocketURL(), "graphql-transport-ws"); } catch { scheduleReconnect(); return; }
  socket.addEventListener("open", () => {
    setStatus("connected");
    socket.send(JSON.stringify({ type: "connection_init", payload: {} }));
  });
  socket.addEventListener("message", (event) => {
    let payload; try { payload = JSON.parse(event.data); } catch { return; }
    if (payload.type === "connection_ack") {
      socket.send(JSON.stringify({ id: "notifications", type: "subscribe", payload: { query: "subscription NotificacionesSubs { notificaciones_subs { title data_json } }" } }));
    } else if (payload.type === "next") handleNotification(payload.payload);
  });
  socket.addEventListener("close", () => { setStatus("error"); scheduleReconnect(); });
  socket.addEventListener("error", () => socket.close());
}

export function startRealtime() {
  if (window.__authV2RealtimeStarted) return;
  window.__authV2RealtimeStarted = true;
  stopped = false;
  connect();
}

export function stopRealtime() {
  stopped = true;
  clearTimeout(reconnectTimer);
  reconnectTimer = null;
  socket?.close();
  socket = null;
  window.__authV2RealtimeStarted = false;
}

export async function refreshChatMessages(conversationID, target) {
  const data = await graphql(queries.messages, { id: conversationID });
  const messages = data.chat_mensajes || [];
  target.innerHTML = `<div class="card-pad"><h2>Mensajes</h2>${messages.map((item) => `<p><strong>${esc(item.sender_id)}</strong><br>${esc(item.texto)} <span class="small muted">${esc(item.created_at || "")}</span></p>`).join("") || `<p class="muted">Sin mensajes.</p>`}</div>`;
}
