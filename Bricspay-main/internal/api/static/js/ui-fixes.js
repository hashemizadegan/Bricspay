(() => {
  "use strict";

  const tabs = ["overview", "workflow", "kyc", "admin"];

  function showTab(name) {
    if (!tabs.includes(name)) return;

    const buttons = document.querySelectorAll(".nav-btn");

    tabs.forEach((tab, index) => {
      const pane = document.getElementById(`tab-${tab}`);
      const selected = tab === name;

      if (buttons[index]) {
        buttons[index].classList.toggle("active", selected);
        buttons[index].setAttribute("aria-selected", String(selected));
      }

      if (pane) {
        pane.classList.toggle("active", selected);
        pane.hidden = !selected;

        // Do not depend on an unknown existing .tab-pane CSS rule.
        pane.style.setProperty(
          "display",
          selected ? "block" : "none",
          "important"
        );
      }
    });
  }

  function init() {
    // Capture the click before the existing inline onclick is reached.
    // Inline onclick may be blocked by CSP or may call a broken switchTab.
    document.addEventListener(
      "click",
      (event) => {
        const button = event.target.closest(".nav-btn");
        if (!button) return;

        const buttons = [...document.querySelectorAll(".nav-btn")];
        const index = buttons.indexOf(button);
        if (index < 0 || index >= tabs.length) return;

        event.preventDefault();
        event.stopImmediatePropagation();
        showTab(tabs[index]);
      },
      true
    );

    showTab("overview");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init, { once: true });
  } else {
    init();
  }
})();
