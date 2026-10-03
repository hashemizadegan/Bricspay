(() => {
  "use strict";

  const tabs = ["workflow", "kyc", "admin"];

  function showTab(name) {
    if (!tabs.includes(name)) return;

    document.querySelectorAll(".tab-btn").forEach((button) => {
      const selected = button.dataset.tab === name;
      button.classList.toggle("active", selected);
      button.setAttribute("aria-selected", String(selected));
    });

    tabs.forEach((tab) => {
      const pane = document.getElementById(`tab-${tab}`);
      if (!pane) return;

      const selected = tab === name;
      pane.classList.toggle("active", selected);
      pane.hidden = !selected;
      pane.style.setProperty("display", selected ? "block" : "none", "important");
    });
  }

  function init() {
    document.addEventListener("click", (event) => {
      const button = event.target.closest(".tab-btn[data-tab]");
      if (!button) return;

      const name = button.dataset.tab;
      if (!tabs.includes(name)) return;

      event.preventDefault();
      showTab(name);
    });

    const initiallyActive = document.querySelector(".tab-btn.active[data-tab]");
    showTab(initiallyActive?.dataset.tab || "workflow");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init, { once: true });
  } else {
    init();
  }
})();
