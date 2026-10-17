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

  const processMessage = (payload) => {
    const notification = payload?.data?.notificaciones_subs;
    if (!notification) return;

    if (notification.data_json) {
      try {
        const data = JSON.parse(notification.data_json);
        if (data.tipo === "conectados") {
          setCounters(data.datos?.conectados ?? 0, data.datos?.total_conectados ?? 0);
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
      opened: () => setStatus("connected"),
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
