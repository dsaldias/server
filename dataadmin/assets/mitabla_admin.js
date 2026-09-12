function normalizarTexto(texto) {
  return texto
    .toLowerCase()
    .normalize("NFD")
    .replace(/[\u0300-\u036f]/g, "");
}

// ============================================================
// BUSQUEDA
// ============================================================

document.addEventListener("input", function (e) {
  const input = e.target.closest("[data-tabla-busqueda]");

  if (!input) {
    return;
  }

  const listaId = input.dataset.tablaBusqueda;
  const contenedor = document.getElementById(listaId);

  if (!contenedor) {
    return;
  }

  const texto = normalizarTexto(input.value);

  const filas = contenedor.querySelectorAll("tbody tr");

  filas.forEach((fila) => {
    const contenido = normalizarTexto(fila.textContent);

    fila.dataset.filtroVisible = contenido.includes(texto) ? "true" : "false";
  });

  contenedor.dataset.paginaActual = "1";

  if (contenedor.hasAttribute("data-tabla-paginable")) {
    actualizarPaginacion(contenedor);
  } else {
    filas.forEach((fila) => {
      fila.style.display = fila.dataset.filtroVisible === "true" ? "" : "none";
    });
  }
});

// ============================================================
// ORDENAMIENTO
// ============================================================

function actualizarIndicadoresOrden(tabla) {
  tabla.querySelectorAll("thead th").forEach((th) => {
    if (th.hasAttribute("no-ordenar")) {
      th.style.cursor = "";
      return;
    }

    th.style.cursor = "pointer";

    // Guardamos el texto original solamente una vez
    if (!th.dataset.textoOriginal) {
      th.dataset.textoOriginal = th.textContent.trim();
    }

    const direccion = th.dataset.ordenDireccion;

    if (direccion === "asc") {
      th.innerHTML = `${th.dataset.textoOriginal} <span class="text-xs">↑</span>`;
    } else if (direccion === "desc") {
      th.innerHTML = `${th.dataset.textoOriginal} <span class="text-xs">↓</span>`;
    } else {
      th.textContent = th.dataset.textoOriginal;
    }
  });
}

document.addEventListener("click", function (e) {
  const th = e.target.closest("[data-tabla-ordenable] thead th");

  if (!th) {
    return;
  }

  if (th.hasAttribute("no-ordenar")) {
    return;
  }

  const tabla = th.closest("table");

  if (!tabla) {
    return;
  }

  const tbody = tabla.querySelector("tbody");

  if (!tbody) {
    return;
  }

  const filas = Array.from(tbody.querySelectorAll("tr"));

  if (filas.length === 0) {
    return;
  }

  const columnas = Array.from(th.parentElement.children);

  const indice = columnas.indexOf(th);

  if (indice === -1) {
    return;
  }

  const direccionActual = th.dataset.ordenDireccion || "desc";

  const nuevaDireccion = direccionActual === "asc" ? "desc" : "asc";

  // Quitar orden de todas las demás columnas
  tabla.querySelectorAll("thead th[data-orden-direccion]").forEach((header) => {
    delete header.dataset.ordenDireccion;
  });

  th.dataset.ordenDireccion = nuevaDireccion;

  filas.sort((a, b) => {
    const celdaA = a.children[indice];
    const celdaB = b.children[indice];

    if (!celdaA || !celdaB) {
      return 0;
    }

    const valorA = normalizarTexto(celdaA.textContent.trim());

    const valorB = normalizarTexto(celdaB.textContent.trim());

    // Intentar ordenar como número
    const numeroA = Number(valorA.replace(",", "."));

    const numeroB = Number(valorB.replace(",", "."));

    if (!isNaN(numeroA) && !isNaN(numeroB) && valorA !== "" && valorB !== "") {
      return nuevaDireccion === "asc" ? numeroA - numeroB : numeroB - numeroA;
    }

    // Ordenar como texto
    const comparacion = valorA.localeCompare(valorB, undefined, {
      numeric: true,
      sensitivity: "base",
    });

    return nuevaDireccion === "asc" ? comparacion : -comparacion;
  });

  filas.forEach((fila) => {
    tbody.appendChild(fila);
  });

  actualizarIndicadoresOrden(tabla);

  const contenedor = tabla.closest("[data-tabla-paginable]");

  if (contenedor) {
    actualizarPaginacion(contenedor);
  }
});

function inicializarTablasOrdenables() {
  document.querySelectorAll("[data-tabla-ordenable] table").forEach((tabla) => {
    actualizarIndicadoresOrden(tabla);
  });
}

document.addEventListener("DOMContentLoaded", inicializarTablasOrdenables);

document.addEventListener("htmx:afterSwap", inicializarTablasOrdenables);

// ============================================================
// MENÚ CONTEXTUAL
// ============================================================

document.addEventListener("contextmenu", function (e) {
  const fila = e.target.closest("[data-tabla-contextual] tbody tr");

  if (!fila) {
    return;
  }

  // Buscar el menú contextual de esta fila
  const menu = fila.querySelector("[data-contextual-menu]");

  if (!menu) {
    return;
  }

  e.preventDefault();

  // Cerrar otros menús
  document.querySelectorAll("[data-contextual-menu]").forEach((item) => {
    item.classList.add("hidden");
  });

  menu.classList.remove("hidden");

  // Posición inicial
  let x = e.clientX;
  let y = e.clientY;

  // Mostrarlo antes de calcular para conocer dimensiones
  const rect = menu.getBoundingClientRect();

  // Evitar que salga por la derecha
  if (x + rect.width > window.innerWidth) {
    x = window.innerWidth - rect.width - 10;
  }

  // Evitar que salga por abajo
  if (y + rect.height > window.innerHeight) {
    y = window.innerHeight - rect.height - 10;
  }

  menu.style.left = `${Math.max(10, x)}px`;
  menu.style.top = `${Math.max(10, y)}px`;
});

// Cerrar menú con click izquierdo
document.addEventListener("click", function (e) {
  if (e.target.closest("[data-contextual-menu]")) {
    return;
  }

  document.querySelectorAll("[data-contextual-menu]").forEach((menu) => {
    menu.classList.add("hidden");
  });
});

// Cerrar con Escape
document.addEventListener("keydown", function (e) {
  if (e.key !== "Escape") {
    return;
  }

  document.querySelectorAll("[data-contextual-menu]").forEach((menu) => {
    menu.classList.add("hidden");
  });
});

function inicializarMenusContextuales() {
  document
    .querySelectorAll("[data-tabla-contextual] tbody tr")
    .forEach((fila) => {
      const menu = fila.querySelector("[data-contextual-menu]");

      fila.classList.toggle("cursor-context-menu", !!menu);
    });
}

document.addEventListener("DOMContentLoaded", inicializarMenusContextuales);

document.addEventListener("htmx:afterSwap", inicializarMenusContextuales);

// ============================================================
// PAGINACIÓN
// ============================================================

function actualizarPaginacion(contenedor) {

  const listaId = contenedor.id;

  const paginacion = document.querySelector(
    `[data-tabla-paginacion="${listaId}"]`,
  );

  if (!paginacion) {
    console.error("❌ No encuentra el componente de paginación");
    return;
  }

  const tabla = contenedor.querySelector("table");

  if (!tabla) {
    console.error("❌ No encuentra la tabla");
    return;
  }

  const tbody = tabla.querySelector("tbody");

  if (!tbody) {
    console.error("❌ No encuentra tbody");
    return;
  }

  const filas = Array.from(tbody.querySelectorAll("tr"));

  const filasPorPagina = parseInt(contenedor.dataset.filasPorPagina, 10);

  if (!filasPorPagina || filasPorPagina <= 0) {
    console.error("❌ filasPorPagina inválido");
    return;
  }

  // Filas que coinciden con la búsqueda
  const filasVisibles = filas.filter(
    (fila) => fila.dataset.filtroVisible !== "false",
  );

  const totalFilas = filasVisibles.length;

  const totalPaginas = Math.max(1, Math.ceil(totalFilas / filasPorPagina));

  let paginaActual = parseInt(contenedor.dataset.paginaActual || "1", 10);

  if (paginaActual > totalPaginas) {
    paginaActual = totalPaginas;
  }

  if (paginaActual < 1) {
    paginaActual = 1;
  }

  contenedor.dataset.paginaActual = paginaActual;

  // Ocultar todas las filas
  filas.forEach((fila) => {
    fila.style.display = "none";
  });

  // Determinar rango
  const inicio = (paginaActual - 1) * filasPorPagina;
  const fin = inicio + filasPorPagina;

  filasVisibles.slice(inicio, fin).forEach((fila) => {
    fila.style.display = "";
  });

  // ==========================================================
  // INFORMACIÓN
  // ==========================================================

  const info = paginacion.querySelector("[data-paginacion-info]");

  if (info) {
    if (totalFilas === 0) {
      info.textContent = "Sin resultados";
    } else {
      const desde = inicio + 1;
      const hasta = Math.min(fin, totalFilas);

      info.textContent = `Mostrando ${desde} - ${hasta} de ${totalFilas}`;
    }
  }

  // ==========================================================
  // CONTROLES
  // ==========================================================

  const controles = paginacion.querySelector("[data-paginacion-controles]");

  if (!controles) {
    return;
  }

  controles.innerHTML = "";

  if (totalPaginas <= 1) {
    return;
  }

  const crearBoton = (texto, pagina, disabled = false) => {
    const boton = document.createElement("button");

    boton.type = "button";
    boton.textContent = texto;

    boton.className =
      "inline-flex h-8 min-w-8 items-center justify-center rounded-md border px-2 text-sm hover:bg-muted disabled:pointer-events-none disabled:opacity-50";

    boton.disabled = disabled;

    boton.addEventListener("click", () => {
      contenedor.dataset.paginaActual = pagina;
      actualizarPaginacion(contenedor);
    });

    return boton;
  };

  // Anterior
  controles.appendChild(crearBoton("‹", paginaActual - 1, paginaActual === 1));

  // Páginas
  for (let pagina = 1; pagina <= totalPaginas; pagina++) {
    const boton = crearBoton(pagina, pagina, false);

    if (pagina === paginaActual) {
      boton.classList.add("bg-primary", "text-primary-foreground");
    }

    controles.appendChild(boton);
  }

  // Siguiente
  controles.appendChild(
    crearBoton("›", paginaActual + 1, paginaActual === totalPaginas),
  );
}
function inicializarFilasTabla(contenedor) {
  const filas = contenedor.querySelectorAll("tbody tr");

  filas.forEach((fila) => {
    if (fila.dataset.filtroVisible === undefined) {
      fila.dataset.filtroVisible = "true";
    }
  });
}

function inicializarTablasPaginables() {
  document.querySelectorAll("[data-tabla-paginable]").forEach((contenedor) => {
    const tabla = contenedor.querySelector("table");

    if (!tabla) {
      return;
    }

    const tbody = tabla.querySelector("tbody");

    if (!tbody) {
      return;
    }

    inicializarFilasTabla(contenedor);

    if (!contenedor.dataset.paginaActual) {
      contenedor.dataset.paginaActual = "1";
    }

    actualizarPaginacion(contenedor);
  });
}

document.addEventListener("DOMContentLoaded", inicializarTablasPaginables);

document.addEventListener("htmx:afterSwap", inicializarTablasPaginables);
