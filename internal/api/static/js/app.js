const App = {
  state: {
    lang: localStorage.getItem('brics_lang') || 'fa',
    token: localStorage.getItem('brics_token') || null,
    user: JSON.parse(localStorage.getItem('brics_user') || 'null')
  },

  init() {
    this.bindEvents();
    this.updateAuthUI();
    this.switchTab('workflow');
  },

  bindEvents() {
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => this.switchTab(e.target.dataset.tab));
    });

    const langSel = document.getElementById('langSelect');
    if (langSel) {
      langSel.value = this.state.lang;
      langSel.addEventListener('change', (e) => {
        this.state.lang = e.target.value;
        localStorage.setItem('brics_lang', this.state.lang);
        document.documentElement.dir = this.state.lang === 'fa' ? 'rtl' : 'ltr';
      });
    }
    document.documentElement.dir = this.state.lang === 'fa' ? 'rtl' : 'ltr';
  },

  switchTab(tabId) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.toggle('active', btn.dataset.tab === tabId);
    });
    document.querySelectorAll('.tab-content').forEach(c => {
      c.classList.toggle('active', c.id === `tab-${tabId}`);
    });

    if (tabId === 'admin' && window.AdminModule) window.AdminModule.loadData();
    if (tabId === 'kyc' && window.KYCModule) window.KYCModule.checkStatus();
  },

  updateAuthUI() {
    const authBtn = document.getElementById('authNavBtn');
    const userBadge = document.getElementById('userNavInfo');
    const adminTab = document.querySelector('[data-tab="admin"]');

    if (this.state.token && this.state.user) {
      if (authBtn) authBtn.style.display = 'none';
      if (userBadge) {
        userBadge.style.display = 'flex';
        document.getElementById('navUserEmail').innerText = this.state.user.email;
      }
      if (adminTab) adminTab.style.display = this.state.user.role === 'admin' ? 'block' : 'none';
    } else {
      if (authBtn) authBtn.style.display = 'inline-block';
      if (userBadge) userBadge.style.display = 'none';
      if (adminTab) adminTab.style.display = 'none';
    }
  },

  async api(url, options = {}) {
    options.headers = options.headers || {};
    if (this.state.token) {
      options.headers['Authorization'] = `Bearer ${this.state.token}`;
    }
    if (!(options.body instanceof FormData)) {
      options.headers['Content-Type'] = 'application/json';
    }
    const res = await fetch(url, options);
    if (res.status === 401) {
      AuthModule.logout();
    }
    return res;
  }
};

document.addEventListener('DOMContentLoaded', () => App.init());
