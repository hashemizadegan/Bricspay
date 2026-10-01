document.addEventListener("DOMContentLoaded", () => {
  const container = document.getElementById("bric-pay-external-section");
  if (!container) return;

  fetch("/static/bric-pay-content.html")
    .then((res) => {
      if (!res.ok) throw new Error("خطا در بارگذاری محتوای BRICS Pay");
      return res.text();
    })
    .then((html) => {
      container.innerHTML = html;
      fetch("/static/bric-pay-content.html")
  .then((res) => res.text())
  .then((html) => {
    container.innerHTML = html;
    if (window.App && typeof App.applyLanguage === 'function') {
      App.applyLanguage(App.state.lang);
    }
  });

    })
    .catch((err) => {
      console.warn("Could not load external BRICS Pay content:", err);
    });
});
