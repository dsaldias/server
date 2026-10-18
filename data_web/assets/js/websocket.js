(() => {
  "use strict";

  if (window.__authV2WebsocketStarted) return;
  window.__authV2WebsocketStarted = true;

  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.host}/ws`;
  const statusDots = () => document.querySelectorAll("[data-ws-status]");

  const setStatus = (status) => {
    const colors = {
      connecting: ["#f97316", "0 0 18px #f97316", "Conectando al servidor"],
      connected: ["#65c28c", "0 0 18px #65c28c", "WebSocket conectado"],
      error: ["#ef4444", "0 0 18px #ef4444", "WebSocket desconectado"],
    };
    const [color, shadow, label] = colors[status] || colors.error;
    statusDots().forEach((dot) => {
      dot.style.backgroundColor = color;
      dot.style.boxShadow = shadow;
      dot.setAttribute("title", label);
      dot.setAttribute("aria-label", label);
    });
  };

  const setCounters = (uniqueUsers, totalConnections) => {
    document.querySelectorAll("[data-ws-unique]").forEach((element) => {
      element.textContent = String(uniqueUsers);
    });
    document.querySelectorAll("[data-ws-total]").forEach((element) => {
      element.textContent = String(totalConnections);
    });
    document.querySelectorAll("[data-ws-connected-summary]").forEach((element) => {
      element.setAttribute("title", `${uniqueUsers} usuarios, ${totalConnections} conexiones`);
    });
  };

  const showVersionUpdate = (serverVersion) => {
    const currentVersion = document.querySelector("meta#x-asset-version")?.dataset.value;
    if (!currentVersion || !serverVersion || currentVersion === serverVersion) return;
    if (window.__authV2VersionUpdateShown) return;

    const host = document.getElementById("version-update-host");
    if (!host) return;
    window.__authV2VersionUpdateShown = true;

    const update = document.createElement("button");
    update.type = "button";
    update.className = "pointer-events-auto w-full max-w-lg cursor-pointer rounded-xl border border-primary/30 bg-card px-4 py-3 text-left text-sm font-medium text-card-foreground shadow-xl ring-1 ring-black/5 transition hover:bg-muted dark:ring-white/10";
    update.setAttribute("aria-label", "Actualizar la aplicación");
    update.textContent = "Hay una nueva versión disponible. Haz clic aquí para actualizar.";
    update.addEventListener("click", () => window.location.reload());
    host.replaceChildren(update);
  };

  const showNotification = (message) => {
    const title = String(message || "Nueva notificación").trim();
    const host = document.getElementById("notification-host");
    if (!host) return;

    const toast = document.createElement("div");
    toast.setAttribute("role", "status");
    toast.className = "pointer-events-auto flex items-start gap-3 rounded-xl border border-transparent p-4 text-sm font-medium shadow-xl";
    toast.style.backgroundColor = "#65c28c";
    toast.style.color = "#081027";

    const content = document.createElement("p");
    content.className = "min-w-0 flex-1 text-sm font-medium leading-5";
    content.textContent = title;

    const close = document.createElement("button");
    close.type = "button";
    close.className = "shrink-0 rounded-md px-1 text-lg leading-none opacity-75 hover:opacity-100";
    close.style.color = "#081027";
    close.setAttribute("aria-label", "Cerrar notificación");
    close.textContent = "×";
    close.addEventListener("click", () => toast.remove());

    toast.append(content, close);
    host.append(toast);

    while (host.children.length > 5) {
      host.firstElementChild?.remove();
    }

    window.setTimeout(() => toast.remove(), 5000);
  };

  const shownAlertIDs = new Set();
  let activeAlertsRequest = null;

  const waitForDOM = () => {
    if (document.readyState !== "loading") return Promise.resolve();
    return new Promise((resolve) => document.addEventListener("DOMContentLoaded", resolve, { once: true }));
  };

  const refreshActiveAlerts = async () => {
    const path = window.location.pathname;
    if (path === "/adminx" || path === "/adminx/" || path === "/adminx/login") return;
    if (activeAlertsRequest) return activeAlertsRequest;

    activeAlertsRequest = (async () => {
      await waitForDOM();
      const response = await fetch("/adminx/avisos/activos", {
        headers: { Accept: "text/html", "X-Requested-With": "fetch" },
        credentials: "same-origin",
      });
      if (!response.ok) return;

      const html = await response.text();
      const template = document.createElement("template");
      template.innerHTML = html.trim();
      const content = template.content.firstElementChild;
      if (!content) return;

      const ids = Array.from(content.querySelectorAll("[data-active-alert-id]")).map((element) => element.getAttribute("data-active-alert-id"));
      const newIDs = ids.filter((id) => id && !shownAlertIDs.has(id));
      ids.forEach((id) => id && shownAlertIDs.add(id));

      const current = document.getElementById("active-alerts-content");
      if (current) current.replaceWith(content);

      if (newIDs.length > 0 && window.templ?.dialog?.open) {
        window.templ.dialog.open("active-alerts-dialog");
      }
    })()
      .catch((error) => console.error("[WS] No se pudieron cargar los avisos activos:", error))
      .finally(() => {
        activeAlertsRequest = null;
      });

    return activeAlertsRequest;
  };

  const processMessage = (payload) => {
    const notification = payload?.data?.notificaciones_subs;
    if (!notification) return;

    if (!notification.data_json) {
      showNotification(notification.title);
    } else {
      try {
        const data = JSON.parse(notification.data_json);
        if (data.tipo === "conectados") {
          setCounters(data.datos?.conectados ?? 0, data.datos?.total_conectados ?? 0);
          showVersionUpdate(data.datos?.assets_version);
        } else if (data.tipo === "alerta") {
          refreshActiveAlerts();
        } else if (data.tipo === "chat") {
          document.getElementById("chat-dialog")?.dispatchEvent(new CustomEvent("auth-v2-chat", { detail: data }));
        } else if (!data.tipo) {
          showNotification(notification.title);
        }
      } catch (error) {
        console.error("[WS] No se pudo procesar la notificación:", error);
      }
    }

    window.dispatchEvent(new CustomEvent("auth-v2-notification", { detail: notification }));
  };

  setStatus("connecting");

  if (!window.graphqlWs?.createClient) {
    console.error("[WS] No está disponible graphql-ws.");
    setStatus("error");
    return;
  }

  const client = window.graphqlWs.createClient({
    url: wsUrl,
    retryAttempts: Infinity,
    shouldRetry: () => true,
    retryWait: async (retries) => {
      const delay = Math.min(1000 * 2 ** retries, 15000);
      await new Promise((resolve) => setTimeout(resolve, delay));
    },
    on: {
      connecting: () => setStatus("connecting"),
      opened: () => {
        setStatus("connected");
        refreshActiveAlerts();
      },
      closed: () => setStatus("error"),
      error: () => setStatus("error"),
    },
  });

  client.subscribe(
    {
      query: `subscription NotificacionesSubs { notificaciones_subs { title data_json } }`,
    },
    {
      next: processMessage,
      error: (error) => {
        console.error("[WS] Error en la suscripción:", error);
        setStatus("error");
      },
      complete: () => setStatus("error"),
    },
  );
})();
