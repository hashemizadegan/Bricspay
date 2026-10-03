(() => {
  "use strict";

  function initTabs() {
    const buttons = Array.from(
      document.querySelectorAll(".nav-btn[data-tab]")
    );

    const panes = Array.from(
      document.querySelectorAll(".tab-pane")
    );

    if (buttons.length === 0 || panes.length === 0) {
      console.warn("Tabs: buttons or panels were not found.");
      return;
    }

    function activateTab(tabName) {
      const targetPane = document.getElementById(`tab-${tabName}`);

      if (!targetPane) {
        console.warn(`Tabs: panel not found: tab-${tabName}`);
        return;
      }

      buttons.forEach((button) => {
        const isActive = button.dataset.tab === tabName;

        button.classList.toggle("active", isActive);
        button.setAttribute("aria-selected", String(isActive));
        button.setAttribute("tabindex", isActive ? "0" : "-1");
      });

      panes.forEach((pane) => {
        const isActive = pane === targetPane;

        pane.classList.toggle("active", isActive);

        if (isActive) {
          pane.removeAttribute("hidden");
          pane.style.display = "block";
        } else {
          pane.setAttribute("hidden", "");
          pane.style.display = "none";
        }
      });
    }

    buttons.forEach((button) => {
      button.addEventListener("click", (event) => {
        event.preventDefault();
        activateTab(button.dataset.tab);
      });

      button.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          activateTab(button.dataset.tab);
        }
      });
    });

    const activeButton =
      buttons.find((button) => button.classList.contains("active")) ||
      buttons[0];

    if (activeButton) {
      activateTab(activeButton.dataset.tab);
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", initTabs);
  } else {
    initTabs();
  }
})();
