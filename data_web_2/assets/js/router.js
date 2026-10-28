const routes = [];
export function route(pattern, render) { routes.push({ pattern, render }); }
function match(pattern, path) { const names = []; const regex = new RegExp(`^${pattern.replace(/:[^/]+/g, (name) => { names.push(name.slice(1)); return "([^/]+)"; })}$`); const result = path.match(regex); return result ? Object.fromEntries(names.map((name, index) => [name, decodeURIComponent(result[index + 1])])) : null; }
export function resolve(path = location.pathname) { for (const item of routes) { const params = match(item.pattern, path); if (params) return { ...item, params }; } return null; }
export function navigate(path) { history.pushState({}, "", path); window.dispatchEvent(new PopStateEvent("popstate")); }
export function startRouter() { document.addEventListener("click", (event) => { const link = event.target.closest("[data-link]"); if (!link || event.metaKey || event.ctrlKey) return; event.preventDefault(); navigate(link.getAttribute("href")); }); window.addEventListener("popstate", () => window.dispatchEvent(new CustomEvent("routechange"))); }
