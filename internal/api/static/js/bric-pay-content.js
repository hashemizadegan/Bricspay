document.addEventListener('DOMContentLoaded', () => {
  const container = document.getElementById('bric-pay-external-section');
  if (!container) return;

  fetch('/static/bric-pay-content.html')
    .then(response => {
      if (!response.ok) {
        throw new Error('Failed to load BRICS Pay external component');
      }
      return response.text();
    })
    .then(html => {
      container.innerHTML = html;
    })
    .catch(err => {
      console.warn('[BRICS Pay] Component load notice:', err.message);
    });
});
