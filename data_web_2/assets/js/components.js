export const esc = (value = "") => String(value).replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" }[char]));
export const date = (value) => value ? new Intl.DateTimeFormat("es-BO", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) : "—";
export const status = (value) => `<span class="badge ${value === "activo" || value === "open" || value === true ? "badge-success" : "badge-danger"}">${esc(value === true ? "Activo" : value || "Inactivo")}</span>`;
export function pageHead(title, subtitle, actions = "") { return `<div class="page-head"><div><h1>${esc(title)}</h1><p class="muted">${esc(subtitle || "")}</p></div><div class="page-actions">${actions}</div></div>`; }
export function loading() { return `<div class="loading">Cargando información…</div>`; }
export function errorBox(error) { return `<div class="error">${esc(error?.message || error || "No fue posible cargar la información.")}</div>`; }
export function modal(title, body) { return `<div class="modal-backdrop" data-modal><section class="modal" role="dialog" aria-modal="true"><header class="modal-head"><h2>${esc(title)}</h2><button class="btn btn-secondary icon-btn" data-close-modal aria-label="Cerrar">×</button></header><div class="modal-body">${body}</div></section></div>`; }
export function formField(label, name, value = "", type = "text", extra = "") { return `<div class="field"><label for="field-${name}">${esc(label)}</label><input id="field-${name}" name="${esc(name)}" type="${type}" value="${esc(value ?? "")}" ${extra}></div>`; }
export function formData(form) {
  const data = Object.fromEntries(new FormData(form).entries());
  for (const field of form.querySelectorAll("[name][multiple]")) data[field.name] = Array.from(field.selectedOptions).map((option) => option.value);
  return data;
}
export function toast(message, isError = false) { const node = document.createElement("div"); node.className = `toast ${isError ? "error" : ""}`; node.textContent = message; document.querySelector("#toast-region").append(node); setTimeout(() => node.remove(), 4200); }
export function table(headers, rows, empty = "No hay registros") { return `<div class="table-wrap"><table><thead><tr>${headers.map((header) => `<th>${esc(header)}</th>`).join("")}</tr></thead><tbody>${rows || `<tr><td colspan="${headers.length}" class="empty">${esc(empty)}</td></tr>`}</tbody></table></div>`; }
export function pagination(meta) {
  if (!meta || Number(meta.paginas || 0) <= 1) return "";
  const page = Number(meta.pagina || 1);
  const pages = Number(meta.paginas || 1);
  const size = Number(meta.size || 10);
  return `<div class="pagination"><span class="muted small">Página ${page} de ${pages} · ${meta.registros || 0} registros</span><div class="pagination-actions"><label class="small muted" for="pagination-size">Mostrar</label><select id="pagination-size" data-page-size><option ${size === 10 ? "selected" : ""}>10</option><option ${size === 25 ? "selected" : ""}>25</option><option ${size === 50 ? "selected" : ""}>50</option></select><button class="btn btn-secondary small" data-page-nav="${Math.max(1, page - 1)}" ${page <= 1 ? "disabled" : ""}>Anterior</button><button class="btn btn-secondary small" data-page-nav="${Math.min(pages, page + 1)}" ${page >= pages ? "disabled" : ""}>Siguiente</button></div></div>`;
}
