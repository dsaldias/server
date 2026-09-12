const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
const wsUrl = `${protocol}//${window.location.host}/ws`;

console.log("[WS] Conectando a:", wsUrl);

const client = graphqlWs.createClient({
  url: wsUrl,

  retryAttempts: Infinity,

  shouldRetry: () => true,

  retryWait: async (retries) => {
    const delay = Math.min(1000 * 2 ** retries, 15000);

    console.log(`[WS] Reintentando en ${delay}ms...`);

    await new Promise((resolve) => setTimeout(resolve, delay));
  },

  on: {
    connecting: () => {
      console.log("[WS] Conectando...");
      cambiar_estado_ws("conectando");
    },

    opened: () => {
      console.log("[WS] Conexión establecida");
      cambiar_estado_ws("conectado");
    },

    closed: (event) => {
      console.log("[WS] Conexión cerrada", event);
      cambiar_estado_ws("error");
    },

    error: (error) => {
      console.error("[WS] Error", error);
      cambiar_estado_ws("error");
    },
  },
});

const NOTIFICACIONES_SUBS = `
    subscription notificaciones_subs {
        notificaciones_subs {
            title
            data_json
        }
    }
`;

const procesar_mensajes = (data) => {
  const notificacion = data?.data?.notificaciones_subs;

  if (!notificacion) {
    return;
  }

  const dataJson = notificacion.data_json;

  if (!dataJson) {
    mostrar_notificacion_ws(notificacion.title || "Nueva notificación");
    return;
  }

  try {
    const datos = JSON.parse(dataJson);

    if (datos.tipo === "conectados") {
      const total = datos.datos?.total_conectados ?? 0;
      const conectados = datos.datos?.conectados ?? 0;

      const contador = document.getElementById("ws_total_conectados");
      const tooltip = document.getElementById("ws_tabs_conectados");

      if (contador) {
        contador.textContent = conectados;
      }

      if (tooltip) {
        tooltip.title = `Conectados: ${total}`;
      }
    }else if (!datos.tipo){
      mostrar_notificacion_ws(notificacion.title || "Nueva notificación");
    }

  } catch (error) {
    console.error("[WS] Error procesando mensaje:", error);
  }
};

const mostrar_notificacion_ws = (mensaje) => {
  Toastify({
    text: mensaje,
    duration: 3000,
    gravity: "top",
    position: "right",
    close: true,
    // backgroundColor: "#21ba45",
    style: {
      background: "#21ba45",
    },
  }).showToast();
};

const cambiar_estado_ws = (estado) => {
  const badge = document.getElementById("ws_badge_indicator");
  const punto = badge?.querySelector("[data-ws-status]");

  if (!badge || !punto) return;

  badge.classList.remove("bg-orange-500", "bg-[#479066]", "bg-red-500");

  punto.classList.remove("bg-orange-500", "bg-green-500", "bg-red-500");

  switch (estado) {
    case "conectando":
      badge.classList.add("bg-orange-500");
      punto.classList.add("bg-orange-500");
      break;

    case "conectado":
      badge.classList.add("bg-[#01c4fb]");
      punto.classList.add("bg-green-500");
      break;

    case "error":
      badge.classList.add("bg-red-500");
      punto.classList.add("bg-red-500");
      break;
  }
};

client.subscribe(
  {
    query: NOTIFICACIONES_SUBS,
  },
  {
    next: (data) => {
      // console.log("[WS] Mensaje recibido:", data);
      procesar_mensajes(data);
    },

    error: (error) => {
      console.error("[WS] Error subscription:", error);
      cambiar_estado_ws("error");
    },

    complete: () => {
      console.log("[WS] Subscription completada");
    },
  },
);
