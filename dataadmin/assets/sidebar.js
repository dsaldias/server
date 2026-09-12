const theme = localStorage.getItem("theme");

if (theme === "dark") {
  document.documentElement.classList.add("dark");
}

document.addEventListener("htmx:pushedIntoHistory", function () {
  const path = window.location.pathname;

  document.querySelectorAll("[data-menu-path]").forEach((el) => {
    const activo = el.dataset.menuPath === path;

    el.removeAttribute("data-tui-sidebar-active");

    if (activo) {
      el.setAttribute("data-tui-sidebar-active", "true");
    }
  });
});

document.addEventListener("htmx:responseError", function (event) {
  const xhr = event.detail.xhr;
  const mensaje = xhr.responseText.toLowerCase();

  if (mensaje.includes("proporcione un token")) {
    abrirModalLogin();
    return;
  }

  Toastify({
    text: xhr.responseText,
    style: {
      background: "#ce2324",
      zIndex: 999999,
    },
    escapeMarkup: false,
    duration: 5000,
    close: true,
    gravity: "top",
    position: "right",
  }).showToast();
});

function abrirModalLogin() {
  const modal = document.getElementById("xmodal-relogin");
  modal.classList.remove("hidden");
  modal.classList.add("flex");
  document.getElementById("login-usuario")?.focus();
}
