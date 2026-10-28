const state = { user: null, menus: [], roles: [], unitId: "", roleId: "", theme: localStorage.getItem("theme") || "light" };
const listeners = new Set();
export function getState() { return state; }
export function setState(values) { Object.assign(state, values); listeners.forEach((listener) => listener(state)); }
export function subscribe(listener) { listeners.add(listener); return () => listeners.delete(listener); }
export function hasPermission(permission) { return state.user?.permisos_sueltos?.some((item) => item.metodo === permission) || state.user?.roles?.some((item) => item.rol?.id); }
