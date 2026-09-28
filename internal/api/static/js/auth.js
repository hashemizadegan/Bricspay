const AuthModule = {
  openModal(mode = 'login') {
    document.getElementById('authModal').classList.add('active');
    this.switchMode(mode);
  },

  closeModal() {
    document.getElementById('authModal').classList.remove('active');
  },

  switchMode(mode) {
    const isLogin = mode === 'login';
    document.getElementById('authTitle').innerText = isLogin ? 'ورود به سامانه' : 'ثبت نام شخصیت حقوقی / بانک';
    document.getElementById('registerFields').style.display = isLogin ? 'none' : 'block';
    document.getElementById('authSubmitBtn').innerText = isLogin ? 'ورود' : 'ثبت نام و ایجاد پرونده';
    document.getElementById('authToggleText').innerText = isLogin ? 'حساب کاربری ندارید؟ ثبت نام کنید' : 'حساب دارید؟ وارد شوید';
    document.getElementById('authModal').dataset.mode = mode;
  },

  toggleMode() {
    const current = document.getElementById('authModal').dataset.mode;
    this.switchMode(current === 'login' ? 'register' : 'login');
  },

  async submit(e) {
    e.preventDefault();
    const mode = document.getElementById('authModal').dataset.mode;
    const email = document.getElementById('authEmail').value;
    const password = document.getElementById('authPassword').value;

    if (mode === 'login') {
      try {
        const res = await App.api('/api/v1/auth/login', {
          method: 'POST',
          body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'خطا در ورود');

        App.state.token = data.token;
        App.state.user = { email, role: data.role, status: data.status };
        localStorage.setItem('brics_token', data.token);
        localStorage.setItem('brics_user', JSON.stringify(App.state.user));

        this.closeModal();
        App.updateAuthUI();
        alert('ورود موفقیت‌آمیز بود');
      } catch (err) {
        alert(err.message);
      }
    } else {
      const payload = {
        email,
        password,
        entity_type: document.getElementById('authEntityType').value,
        legal_name: document.getElementById('authLegalName').value,
        registration_number: document.getElementById('authRegNo').value,
        tax_id: document.getElementById('authTaxId').value,
        jurisdiction: document.getElementById('authJurisdiction').value,
        contact_phone: document.getElementById('authPhone').value
      };

      try {
        const res = await App.api('/api/v1/auth/register', {
          method: 'POST',
          body: JSON.stringify(payload)
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'خطا در ثبت نام');

        alert('ثبت نام با موفقیت انجام شد. اکنون وارد شوید.');
        this.switchMode('login');
      } catch (err) {
        alert(err.message);
      }
    }
  },

  logout() {
    App.state.token = null;
    App.state.user = null;
    localStorage.removeItem('brics_token');
    localStorage.removeItem('brics_user');
    App.updateAuthUI();
    App.switchTab('workflow');
  }
};
