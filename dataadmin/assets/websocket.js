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
      const total = datos.datos?.total_conectados ?? 0;
      const conectados = datos.datos?.conectados ?? 0;
      const assets_v = datos.datos?.assets_version ?? "";

      const contadores = document.querySelectorAll("[data-conectados-1]");
      const badges = document.querySelectorAll("[data-conectados-badge]");

      contadores.forEach((contador) => {
        contador.textContent = conectados;
      });
      
      badges.forEach((indicador) => {
        indicador.title = `Conectados: ${total}`;
      });

      if (flag_ver_alertas_on_refresh_page) {
        flag_ver_alertas_on_refresh_page = false;
        mostrarAlertas();
      }

      verificar_new_version(assets_v);
      //
    } else if (datos.tipo == "alerta") {
      // const mensaje = datos.datos?.mensaje || "Nueva alerta del sistema";
      mostrarAlertas();
      //
    } else if (!datos.tipo) {
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

async function mostrarAlertas() {
  const response = await fetch("/adminx/avisos/modal");

  if (!response.ok || response.status === 204) {
    return;
  }

  document.body.insertAdjacentHTML("beforeend", await response.text());

  const modal = document.getElementById("xmodal-alerta");

  modal?.classList.remove("hidden");
  modal?.classList.add("flex");
}

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

const cambiar_estado_ws = (estado) => {
  const badges = document.querySelectorAll("[data-conectados-badge]");

  if (!badges) return;

  for(let i=0;i<badges.length;i++){
    const badge = badges[i];
    badge.classList.remove("bg-orange-500", "bg-[#479066]", "bg-red-500");
    const punto = badge?.querySelector("[data-ws-status]");
    if (!punto) continue;
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
