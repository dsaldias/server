function initWizards() {
  document.querySelectorAll("[data-wizard]").forEach((wizard) => {
    if (wizard.dataset.wizardInitialized) {
      return;
    }

    wizard.dataset.wizardInitialized = "true";

    const steps = wizard.querySelectorAll("[data-wizard-step]");
    const contents = wizard.querySelectorAll("[data-wizard-content]");
    const prev = wizard.querySelector("[data-wizard-prev]");
    const next = wizard.querySelector("[data-wizard-next]");

    let current = 0;

    function showStep(index) {
      current = index;

      contents.forEach((content, i) => {
        content.classList.toggle("hidden", i !== current);
      });

      steps.forEach((step, i) => {
        const number = step.querySelector("[data-wizard-number]");

        number.textContent = i + 1;

        number.classList.toggle("bg-primary", i === current);

        number.classList.toggle("text-white", i === current);
      });

      prev.disabled = current === 0;

      next.textContent =
        current === contents.length - 1 ? "Finalizar" : "Siguiente";
    }

    prev.addEventListener("click", () => {
      if (current > 0) {
        showStep(current - 1);
      }
    });

    next.addEventListener("click", () => {
      if (current < contents.length - 1) {
        showStep(current + 1);
      }
    });

    steps.forEach((step, index) => {
      const trigger = step.querySelector("[data-wizard-trigger]");

      trigger.addEventListener("click", () => {
        showStep(index);
      });
    });

    showStep(0);
  });
}

// Página normal
document.addEventListener("DOMContentLoaded", initWizards);

// Contenido cargado por HTMX
document.body.addEventListener("htmx:afterSwap", initWizards);
