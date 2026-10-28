import { graphql, queries, mutations } from "./api.js";
import { getState, setState } from "./store.js";
import { navigate, resolve, route, startRouter } from "./router.js";
import { date, errorBox, esc, formData, formField, loading, modal, pageHead, status, table, toast } from "./components.js";

const app = document.querySelector("#app");

function loginView(message = "") {
  app.innerHTML = `<main class="login-screen"><section class="card login-card"><div class="brand"><span class="brand-mark">A</span><span>Administración</span></div><h1>Bienvenido</h1><p class="muted">Ingresa para continuar al panel administrativo.</p>${message ? `<p class="error">${esc(message)}</p>` : ""}<form class="login-form" data-action="login"><div class="field"><label for="username">Usuario</label><input id="username" name="username" autocomplete="username" required></div><div class="field"><label for="password">Contraseña</label><input id="password" name="password" type="password" autocomplete="current-password" required></div><button class="btn btn-primary" type="submit">Ingresar</button></form></section></main>`;
}

function normalizeMenuPath(path) {
  if (!path) return "/";
  const normalized = path.replace(/^\/adminx(?=\/|$)/, "") || "/";
  return normalized.startsWith("/") ? normalized : `/${normalized}`;
}

function menuIsActive(menu, path = location.pathname) {
  const menuPath = normalizeMenuPath(menu.path);
  return menuPath === "/" ? path === "/" : path === menuPath || path.startsWith(`${menuPath}/`);
}

const menuIcons = {
  home: "⌂", dashboard: "▦", group: "♙", groups: "♙", people: "♙", person: "♙",
  admin_panel_settings: "◈", settings: "⚙", folder: "▰", menu: "☷", campaign: "◇",
  notifications: "◇", chat: "◌", message: "◌", support: "□", business: "⌗",
  domain: "⌗", local_movies: "▣", key: "⚿", lock: "▣", report: "▤", analytics: "▥",
};

function menuIcon(name, fallback = "•") {
  const key = String(name || "").trim().toLowerCase();
  return `<span class="menu-icon" aria-hidden="true">${menuIcons[key] || fallback}</span>`;
}

function navLink(path, label, icon) { return `<a class="nav-link ${location.pathname === path ? "active" : ""}" data-link href="${path}"><span>${icon || "•"}</span><span>${esc(label)}</span></a>`; }

function sidebarMenu() {
  const menus = getState().menus || [];
  if (!menus.length) {
    return `<div class="nav-section">General</div>${navLink("/", "Inicio", "⌂")}${navLink("/usuarios", "Usuarios", "♙")}${navLink("/roles", "Roles", "◈")}${navLink("/unidades", "Unidades", "⌗")}<div class="nav-section">Operación</div>${navLink("/avisos", "Avisos", "◇")}${navLink("/tickets", "Tickets", "□")}${navLink("/chat", "Chat", "◌")}${navLink("/perfil", "Mi perfil", "◎")}`;
  }

  const childrenByParent = new Map();
  for (const menu of menus) {
    const parentID = menu.padre_id == null ? "root" : String(menu.padre_id);
    if (!childrenByParent.has(parentID)) childrenByParent.set(parentID, []);
    childrenByParent.get(parentID).push(menu);
  }
  const hasActiveDescendant = (menu) => {
    const children = childrenByParent.get(String(menu.id)) || [];
    return children.some((child) => menuIsActive(child) || hasActiveDescendant(child));
  };
  const renderItems = (items, level = 0) => items.sort((a, b) => (a.orden || 0) - (b.orden || 0)).map((menu) => {
    const children = childrenByParent.get(String(menu.id)) || [];
    const hasChildren = children.length > 0;
    const active = menuIsActive(menu) || hasActiveDescendant(menu);
    const path = normalizeMenuPath(menu.path);
    if (!hasChildren) return `<a class="nav-link ${menuIsActive(menu) ? "active" : ""}" data-link href="${path}" style="padding-left:${12 + level * 18}px">${menuIcon(menu.icon)}<span>${esc(menu.label)}</span></a>`;
    const groupID = `menu-group-${esc(String(menu.id))}`;
    return `<div class="nav-group"><button class="nav-link nav-parent" type="button" data-menu-toggle="${groupID}" aria-expanded="${active ? "true" : "false"}" style="padding-left:${12 + level * 18}px">${menuIcon(menu.icon, "▸")}<span>${esc(menu.label)}</span><span class="menu-chevron">⌄</span></button><div id="${groupID}" class="nav-children ${active ? "" : "hidden"}">${renderItems(children, level + 1)}</div></div>`;
  }).join("");
  return renderItems(childrenByParent.get("root") || []);
}

function shell(content, title = "Panel") {
  const state = getState(); const user = state.user?.usuario;
  app.innerHTML = `<div class="shell"><aside class="sidebar" id="sidebar"><div class="brand"><span class="brand-mark">A</span><span>Administración</span></div><nav class="nav">${sidebarMenu()}</nav></aside><section class="main"><header class="topbar"><div class="topbar-actions"><button class="btn btn-secondary icon-btn mobile-menu" data-toggle-sidebar>☰</button><strong>${esc(title)}</strong></div><div class="topbar-actions"><span class="small muted">${esc(user ? `${user.nombres} ${user.apellido1}` : "")}</span><button class="btn btn-secondary" data-action="logout">Salir</button></div></header><main class="page" id="page-content">${content}</main></section></div>`;
}

async function renderPage() {
  if (!getState().user) return loginView();
  const match = resolve();
  if (!match) return navigate("/");
  shell(loading(), pageTitle(match.pattern));
  try { document.querySelector("#page-content").innerHTML = await match.render(match.params); } catch (error) { document.querySelector("#page-content").innerHTML = errorBox(error); }
}
function pageTitle(pattern) { return ({ "/": "Inicio", "/usuarios": "Usuarios", "/roles": "Roles", "/unidades": "Unidades", "/avisos": "Avisos", "/tickets": "Tickets", "/chat": "Chat", "/perfil": "Mi perfil" })[pattern] || "Panel"; }

async function dashboardPage() {
  const data = await graphql(queries.dashboard).catch(() => ({ reporte1: [], reporte2: [], usuarios_conectados: [] }));
  const stats = data.reporte1 || [];
  return `${pageHead("Inicio", "Resumen general del sistema")}<div class="grid grid-4">${stats.slice(0, 4).map((item) => `<article class="card stat"><div class="stat-label">${esc(item.nombre)}</div><div class="stat-value">${item.valor}</div></article>`).join("") || `<article class="card stat"><div class="stat-label">Sesión</div><div class="stat-value">Activa</div></article>`}</div><div class="grid grid-2" style="margin-top:16px"><section class="card card-pad"><h2>Usuarios conectados</h2><p class="muted small">Actividad actual del sistema.</p>${table(["Usuario", "Username", "Conexiones"], (data.usuarios_conectados || []).map((item) => `<tr><td>${esc(`${item.nombres} ${item.apellido1}`)}</td><td>${esc(item.username)}</td><td>${item.conexiones}</td></tr>`).join(""))}</section><section class="card card-pad"><h2>Actividad reciente</h2>${table(["Fecha", "Valor"], (data.reporte2 || []).slice(-8).map((item) => `<tr><td>${esc(item.fecha_str || date(item.fecha))}</td><td>${item.valor}</td></tr>`).join(""))}</section></div>`;
}

async function usersPage() {
  const query = { pagina: Number(new URLSearchParams(location.search).get("pagina") || 1), size: 10, nombre: new URLSearchParams(location.search).get("nombre") || null, rol: null };
  const data = await graphql(queries.users, { query }); const result = data.usuarios2; const users = result?.datos || [];
  return `${pageHead("Usuarios", `${result?.paginacion?.registros || users.length} registros`, `<button class="btn btn-primary" data-modal-form="user">Nuevo usuario</button>`)}<section class="card"><form class="toolbar" data-action="user-filter"><div class="field"><label for="user-search">Buscar</label><input id="user-search" name="nombre" value="${esc(query.nombre || "")}" placeholder="Nombre o usuario"></div><button class="btn btn-secondary" type="submit">Buscar</button></form>${table(["Nombre", "Usuario", "Correo", "Estado", "Último acceso", "Acciones"], users.map((item) => `<tr><td>${esc(`${item.nombres} ${item.apellido1} ${item.apellido2 || ""}`)}</td><td>${esc(item.username)}</td><td>${esc(item.correo || "—")}</td><td>${status(item.estado)}</td><td>${date(item.last_login)}</td><td><div class="actions"><button class="btn btn-secondary small" data-view-user="${esc(item.id)}">Ver</button><button class="btn btn-secondary small" data-edit-user="${esc(item.id)}">Editar</button></div></td></tr>`).join(""))}</section>`;
}

async function rolesPage() {
  const data = await graphql(queries.roles, { q: { pagina: 1, size: 50 } }); const roles = data.roles2?.datos || [];
  return `${pageHead("Roles", "Perfiles de acceso y permisos", `<button class="btn btn-primary" data-modal-form="role">Nuevo rol</button>`)}<section class="card">${table(["Nombre", "Descripción", "Jerarquía", "Menús", "Permisos", "Usuarios", "Acciones"], roles.map((item) => `<tr><td>${esc(item.nombre)}</td><td>${esc(item.descripcion || "—")}</td><td>${item.jerarquia}</td><td>${item.menus}</td><td>${item.permisos}</td><td>${item.usuarios}</td><td><button class="btn btn-secondary small" data-edit-role="${esc(item.id)}">Editar</button></td></tr>`).join(""))}</section>`;
}

async function unitsPage() {
  const data = await graphql(queries.units, { q: { pagina: 1, size: 50, nombre: null } }); const units = data.unidades2?.datos || [];
  return `${pageHead("Unidades", "Sedes o contextos operativos", `<button class="btn btn-primary" data-modal-form="unit">Nueva unidad</button>`)}<section class="card">${table(["Nombre", "Descripción", "Orden", "Ubicación", "Acciones"], units.map((item) => `<tr><td>${esc(item.nombre)}</td><td>${esc(item.descripcion || "—")}</td><td>${item.orden}</td><td>${item.latitud ?? "—"}, ${item.longitud ?? "—"}</td><td><button class="btn btn-secondary small" data-edit-unit="${esc(item.id)}">Editar</button></td></tr>`).join(""))}</section>`;
}

async function noticesPage() {
  const data = await graphql(queries.notices); const notices = data.notificaciones || [];
  return `${pageHead("Avisos", "Mensajes visibles para los usuarios", `<button class="btn btn-primary" data-modal-form="notice">Nuevo aviso</button>`)}<section class="card">${table(["Mensaje", "Desde", "Hasta", "Creado"], notices.map((item) => `<tr><td>${esc(item.mensaje)}</td><td>${date(item.desde)}</td><td>${date(item.hasta)}</td><td>${date(item.fecha_registro)}</td></tr>`).join(""))}</section>`;
}

async function ticketsPage() {
  const data = await graphql(queries.tickets, { q: { estado: null } }); const tickets = data.all_tickets || [];
  return `${pageHead("Tickets", "Solicitudes y soporte", `<button class="btn btn-primary" data-modal-form="ticket">Crear ticket</button>`)}<section class="card">${table(["Usuario", "Problema", "Estado", "Fecha", "Acciones"], tickets.map((item) => `<tr><td>${esc(item.usuario)}</td><td>${esc(item.problema)}</td><td>${status(item.estado)}</td><td>${date(item.fecha_registro)}</td><td><div class="actions"><button class="btn btn-secondary small" data-answer-ticket="${esc(item.id)}">Responder</button>${item.estado !== "cerrado" ? `<button class="btn btn-danger small" data-close-ticket="${esc(item.id)}">Cerrar</button>` : ""}</div></td></tr>`).join(""))}</section>`;
}

async function chatPage() {
  const user = getState().user.usuario; const data = await graphql(queries.chats, { id: user.id }); const chats = data.chats_by_user || [];
  return `${pageHead("Chat", "Conversaciones internas") }<div class="grid grid-2"><section class="card">${table(["Conversación", "Tipo", "No leídos", "Acción"], chats.map((item) => `<tr><td>${esc(item.nombre || "Conversación")}</td><td>${esc(item.tipo)}</td><td>${item.no_leidos}</td><td><button class="btn btn-secondary small" data-open-chat="${esc(item.id)}">Abrir</button></td></tr>`).join(""))}</section><section class="card card-pad" id="chat-messages"><div class="empty">Selecciona una conversación.</div></section></div>`;
}

async function profilePage() {
  const user = getState().user.usuario;
  return `${pageHead("Mi perfil", "Actualiza tus datos personales")}<section class="card card-pad"><form class="grid grid-2" data-action="profile"><input type="hidden" name="id" value="${esc(user.id)}">${formField("Nombres", "nombres", user.nombres, "text", "required")}${formField("Primer apellido", "apellido1", user.apellido1, "text", "required")}${formField("Segundo apellido", "apellido2", user.apellido2)}${formField("Documento", "documento", user.documento)}${formField("Celular", "celular", user.celular)}${formField("Correo", "correo", user.correo, "email")}${formField("Usuario", "username", user.username)}${formField("Nueva contraseña", "password", "", "password", "autocomplete=\"new-password\"")}<div class="field" style="grid-column:1/-1"><label for="field-direccion">Dirección</label><textarea id="field-direccion" name="direccion">${esc(user.direccion || "")}</textarea></div><div class="form-actions" style="grid-column:1/-1"><button class="btn btn-primary" type="submit">Guardar cambios</button></div></form></section>`;
}

function formShell(type, value = {}) {
  const forms = {
    user: `${formField("Nombres", "nombres", value.nombres, "text", "required")}${formField("Primer apellido", "apellido1", value.apellido1, "text", "required")}${formField("Usuario", "username", value.username, "text", "required")}${formField("Correo", "correo", value.correo, "email")}${formField("Contraseña", "password", "", "password", value.id ? "" : "required")}<input type="hidden" name="id" value="${esc(value.id || "")}">`,
    role: `${formField("Nombre", "nombre", value.nombre, "text", "required")}${formField("Descripción", "descripcion", value.descripcion)}${formField("Jerarquía", "jerarquia", value.jerarquia || 1, "number", "min=\"1\"")}<input type="hidden" name="id" value="${esc(value.id || "")}">`,
    unit: `${formField("Nombre", "nombre", value.nombre, "text", "required")}${formField("Descripción", "descripcion", value.descripcion)}${formField("Orden", "orden", value.orden || 1, "number", "required")}${formField("Latitud", "latitud", value.latitud, "number", "step=\"any\"")}${formField("Longitud", "longitud", value.longitud, "number", "step=\"any\"")}<input type="hidden" name="id" value="${esc(value.id || "")}">`,
    notice: `${formField("Mensaje", "mensaje", value.mensaje, "text", "required")}${formField("Desde", "desde", value.desde ? value.desde.slice(0, 16) : "", "datetime-local", "required")}${formField("Hasta", "hasta", value.hasta ? value.hasta.slice(0, 16) : "", "datetime-local", "required")}<input type="hidden" name="id" value="${esc(value.id || "")}">`,
    ticket: `<div class="field"><label for="field-problema">Problema</label><textarea id="field-problema" name="problema" required>${esc(value.problema || "")}</textarea></div>`,
  };
  const names = { user: "Usuario", role: "Rol", unit: "Unidad", notice: "Aviso", ticket: "Ticket" };
  return modal(`${value.id ? "Editar" : "Nuevo"} ${names[type]}`, `<form class="grid grid-2" data-action="${type}">${forms[type]}<div class="form-actions" style="grid-column:1/-1"><button type="button" class="btn btn-secondary" data-close-modal>Cancelar</button><button class="btn btn-primary" type="submit">Guardar</button></div></form>`);
}

function inputNumber(value) { const number = Number(value); return Number.isFinite(number) && value !== "" ? number : null; }
async function submitMutation(type, values) {
  if (type === "user") { const input = { nombres: values.nombres, apellido1: values.apellido1, username: values.username, correo: values.correo || null, password: values.password || null, roles: [], permisos_sueltos: [], menus_sueltos: [] }; return graphql(values.id ? mutations.updateUser : mutations.createUser, { input: values.id ? { ...input, id: values.id } : input }); }
  if (type === "role") { const input = { nombre: values.nombre, descripcion: values.descripcion || null, jerarquia: Number(values.jerarquia || 1), permisos: [], menus: [] }; return graphql(values.id ? mutations.updateRole : mutations.createRole, { input: values.id ? { ...input, id: values.id } : input }); }
  if (type === "unit") { const input = { nombre: values.nombre, descripcion: values.descripcion || null, orden: Number(values.orden || 1), latitud: inputNumber(values.latitud), longitud: inputNumber(values.longitud) }; return graphql(values.id ? mutations.updateUnit : mutations.createUnit, { input: values.id ? { ...input, id: values.id } : input }); }
  if (type === "notice") { const input = { mensaje: values.mensaje, desde: new Date(values.desde).toISOString(), hasta: new Date(values.hasta).toISOString() }; return graphql(values.id ? mutations.updateNotice : mutations.createNotice, { input: values.id ? { ...input, id: values.id } : input }); }
  if (type === "ticket") return graphql(mutations.createTicket, { input: { problema: values.problema } });
  if (type === "profile") return graphql(mutations.updateProfile, { input: { ...values, password: values.password || null } });
}

route("/", dashboardPage); route("/usuarios", usersPage); route("/roles", rolesPage); route("/unidades", unitsPage); route("/avisos", noticesPage); route("/tickets", ticketsPage); route("/chat", chatPage); route("/perfil", profilePage);

document.addEventListener("click", async (event) => {
  const target = event.target;
  if (target.closest("[data-toggle-sidebar]")) document.querySelector("#sidebar")?.classList.toggle("open");
  const menuToggle = target.closest("[data-menu-toggle]");
  if (menuToggle) {
    const group = document.getElementById(menuToggle.dataset.menuToggle);
    const expanded = menuToggle.getAttribute("aria-expanded") === "true";
    menuToggle.setAttribute("aria-expanded", String(!expanded));
    group?.classList.toggle("hidden", expanded);
  }
  if (target.closest("[data-close-modal]")) document.querySelector("[data-modal]")?.remove();
  const formButton = target.closest("[data-modal-form]"); if (formButton) document.body.insertAdjacentHTML("beforeend", formShell(formButton.dataset.modalForm));
  for (const [selector, query, title] of [["data-view-user", queries.user, "Detalle de usuario"], ["data-edit-user", queries.user, "Editar usuario"]]) { const button = target.closest(`[${selector}]`); if (button) { try { const data = await graphql(query, { id: button.getAttribute(selector) }); document.body.insertAdjacentHTML("beforeend", formShell("user", data.usuario_by_id)); } catch (error) { toast(error.message, true); } } }
  const closeTicket = target.closest("[data-close-ticket]"); if (closeTicket && confirm("¿Cerrar este ticket?")) { try { await graphql(mutations.closeTicket, { id: closeTicket.dataset.closeTicket }); toast("Ticket cerrado"); await renderPage(); } catch (error) { toast(error.message, true); } }
  const answer = target.closest("[data-answer-ticket]"); if (answer) document.body.insertAdjacentHTML("beforeend", modal("Responder ticket", `<form class="grid" data-action="answer-ticket"><input type="hidden" name="tickets_id" value="${esc(answer.dataset.answerTicket)}"><div class="field"><label>Respuesta</label><textarea name="respuesta" required></textarea></div><div class="form-actions"><button class="btn btn-primary">Enviar respuesta</button></div></form>`));
  const openChat = target.closest("[data-open-chat]"); if (openChat) { try { const data = await graphql(queries.messages, { id: openChat.dataset.openChat }); document.querySelector("#chat-messages").innerHTML = `<div class="card-pad"><h2>Mensajes</h2>${(data.chat_mensajes || []).map((item) => `<p><strong>${esc(item.sender_id)}</strong><br>${esc(item.texto)} <span class="small muted">${date(item.created_at)}</span></p>`).join("") || `<p class="muted">Sin mensajes.</p>`}</div>`; } catch (error) { toast(error.message, true); } }
});

document.addEventListener("submit", async (event) => {
  const form = event.target; const action = form.dataset.action; if (!action) return; event.preventDefault(); const button = form.querySelector("button[type=submit],button:not([type])"); if (button) button.disabled = true;
  try {
    const values = formData(form);
    if (action === "login") { const data = await graphql(queries.login, { input: { username: values.username, password: values.password } }); const login = data.login_v2; if (!login?.me) throw new Error("No se recibió la sesión del usuario."); setState({ user: login.me, menus: login.me.menus || [], roles: login.me.roles || [] }); toast("Sesión iniciada"); navigate("/"); }
    else if (action === "user-filter") { const params = new URLSearchParams(); if (values.nombre) params.set("nombre", values.nombre); navigate(`/usuarios${params.toString() ? `?${params}` : ""}`); }
    else if (action === "answer-ticket") { await graphql(mutations.answerTicket, { input: { tickets_id: values.tickets_id, respuesta: values.respuesta } }); document.querySelector("[data-modal]")?.remove(); toast("Respuesta enviada"); await renderPage(); }
    else { await submitMutation(action, values); document.querySelector("[data-modal]")?.remove(); toast("Cambios guardados"); await renderPage(); }
  } catch (error) { toast(error.message, true); if (button) button.disabled = false; }
});

document.addEventListener("click", (event) => { if (event.target.closest("[data-action=logout]")) { setState({ user: null }); navigate("/"); } });
window.addEventListener("routechange", renderPage);
startRouter();
renderPage();
