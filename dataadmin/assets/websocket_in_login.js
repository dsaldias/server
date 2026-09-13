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
    },

    opened: () => {
      console.log("[WS] Conexión establecida");
    },

    closed: (event) => {
      console.log("[WS] Conexión cerrada", event);
    },

    error: (error) => {
      console.error("[WS] Error", error);
      mostrar_notificacion_ws("Error conexion con el servidor: "+error)
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

let flag_ver_alertas_on_refresh_page = true;
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
      const assets_v = datos.datos?.assets_version ?? "";
      verificar_new_version(assets_v);
    }
    if (!datos.tipo) {
      mostrar_notificacion_ws(notificacion.title || "Nueva notificación");
    }
  } catch (error) {
    console.error("[WS] Error procesando mensaje:", error);
  }
};

let recargando_version = false;
const verificar_new_version = (assets_version) => {
  const meta = document.getElementById("x-asset-version");
  if (!meta) return;
  const current_version = meta.dataset.value;

  console.log("current_version:", current_version);
  console.log("server_version:", assets_version);

  if (assets_version !== current_version && !recargando_version) {
    recargando_version = true;
    Toastify({
      text: "↻  Hay una nueva versión disponible. Haz clic aquí para actualizar",
      duration: -1,
      gravity: "top",
      position: "center",
      close: false,
      className: "cursor-pointer",
      onClick: function () {
        window.location.reload();
      },
    }).showToast();
  }
};

const mostrar_notificacion_ws = (mensaje) => {
  Toastify({
    text: mensaje,
    duration: 3000,
    gravity: "top",
    position: "right",
    close: true,
    style: {
      background: "#21ba45",
    },
  }).showToast();
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
    },

    complete: () => {
      console.log("[WS] Subscription completada");
    },
  },
);
