const AuthModule = {
  openModal(mode = 'login') {
    document.getElementById('authModal').classList.add('active');
    this.switchMode(mode);
  },

  closeModal() {
    document.getElementById('authModal').classList.remove('active');
    this.clearStatus();
  },

  switchMode(mode) {
    const isLogin = mode === 'login';
    document.getElementById('authTitle').innerText = isLogin ? 'ورود به سامانه' : 'ثبت نام شخصیت حقوقی / بانک';
    document.getElementById('registerFields').style.display = isLogin ? 'none' : 'block';
    document.getElementById('authSubmitBtn').innerText = isLogin ? 'ورود' : 'ثبت نام و ایجاد پرونده';
    document.getElementById('authSwitchText').innerHTML = isLogin 
      ? 'حساب کاربری ندارید؟ <a href="javascript:void(0)" onclick="AuthModule.switchMode('register')">ثبت نام کنید</a>'
      : 'قبلاً ثبت نام کرده‌اید؟ <a href="javascript:void(0)" onclick="AuthModule.switchMode('login')">وارد شوید</a>';
    this.clearStatus();
  },

  clearStatus() {
    const statusEl = document.getElementById('authStatusMessage');
    if (statusEl) {
      statusEl.innerText = '';
      statusEl.className = 'auth-status';
    }
  },

  showStatus(msg, isError = false) {
    const statusEl = document.getElementById('authStatusMessage');
    if (statusEl) {
      statusEl.innerText = msg;
      statusEl.className = 'auth-status ' + (isError ? 'error' : 'success');
    } else {
      alert(msg);
    }
  },

  async handleFormSubmit() {
    const isLogin = document.getElementById('registerFields').style.display === 'none';
    const email = document.getElementById('authEmail').value.trim();
    const password = document.getElementById('authPassword').value;

    if (!email || !password) {
      this.showStatus('لطفاً ایمیل و رمز عبور را وارد کنید.', true);
      return;
    }

    try {
      this.showStatus('در حال پردازش درخواست...', false);
      if (isLogin) {
        const res = await fetch('/auth/login', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'خطا در ورود به سامانه');
        this.saveAuth(data.token, data.role, email);
      } else {
        const payload = {
          email,
          password,
          entity_type: document.getElementById('regEntityType').value,
          legal_name: document.getElementById('regLegalName').value.trim(),
          registration_number: document.getElementById('regRegNo').value.trim(),
          tax_id: document.getElementById('regTaxId').value.trim(),
          jurisdiction: document.getElementById('regJurisdiction').value.trim(),
          contact_phone: document.getElementById('regPhone').value.trim()
        };
        const res = await fetch('/api/v1/auth/register', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
        const data = await res.json();
        if (!res.ok) throw new Error(data.error || 'خطا در ثبت نام پرونده حقوقی');
        this.showStatus('ثبت‌نام با موفقیت انجام شد. اکنون وارد شوید.', false);
        setTimeout(() => this.switchMode('login'), 1200);
      }
    } catch (err) {
      this.showStatus(err.message, true);
    }
  },

  // --- MetaMask Web3 EIP-191 Integration ---
  async loginWithMetaMask() {
    this.clearStatus();
    if (typeof window.ethereum === 'undefined') {
      this.showStatus('کیف پول MetaMask روی مرورگر شما شناسایی نشد. لطفاً اکستنشن متامسک را نصب کنید.', true);
      return;
    }

    try {
      this.showStatus('درخواست اتصال به کیف پول MetaMask...', false);
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      if (!accounts || accounts.length === 0) {
        throw new Error('هیچ آدرس ولتی انتخاب نشد.');
      }
      const walletAddress = accounts[0].toLowerCase();

      this.showStatus(`در حال دریافت چالش امضای دیجیتال برای ولت ${walletAddress.substring(0, 8)}...`, false);
      
      // 1. Fetch challenge from backend
      const challengeRes = await fetch('/api/v1/auth/wallet/challenge', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          wallet_address: walletAddress,
          domain: window.location.host || 'bricspay.local'
        })
      });
      const challengeData = await challengeRes.json();
      if (!challengeRes.ok) {
        throw new Error(challengeData.error || 'خطا در صدور چالش ورود');
      }

      this.showStatus('لطفاً پیام احراز هویت را در متامسک تأیید و امضا نمایید...', false);

      // 2. Request EIP-191 personal sign from user
      const signature = await window.ethereum.request({
        method: 'personal_sign',
        params: [challengeData.message, walletAddress]
      });

      this.showStatus('در حال اعتبارسنجی امضای رمزنگاری‌شده...', false);

      // 3. Verify signature on backend and receive JWT
      const verifyRes = await fetch('/api/v1/auth/wallet/verify', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          wallet_address: walletAddress,
          nonce: challengeData.nonce,
          signature: signature
        })
      });
      const verifyData = await verifyRes.json();
      if (!verifyRes.ok) {
        throw new Error(verifyData.error || 'اعتبارسنجی امضا ناموفق بود.');
      }

      this.saveAuth(verifyData.token, verifyData.role || 'member', walletAddress);
    } catch (err) {
      console.error('MetaMask Auth Error:', err);
      this.showStatus(err.message || 'خطا در احراز هویت متامسک', true);
    }
  },

  saveAuth(token, role, identifier) {
    localStorage.setItem('bricspay_token', token);
    localStorage.setItem('bricspay_role', role);
    localStorage.setItem('bricspay_user', identifier);
    this.showStatus('ورود موفقیت‌آمیز بود! در حال انتقال...', false);
    setTimeout(() => {
      window.location.reload();
    }, 800);
  },

  logout() {
    localStorage.removeItem('bricspay_token');
    localStorage.removeItem('bricspay_role');
    localStorage.removeItem('bricspay_user');
    window.location.reload();
  },

  initUserState() {
    const token = localStorage.getItem('bricspay_token');
    const user = localStorage.getItem('bricspay_user');
    const role = localStorage.getItem('bricspay_role');
    const authContainer = document.getElementById('userNavArea');
    if (!authContainer) return;

    if (token && user) {
      const displayId = user.startsWith('0x') ? `${user.substring(0, 6)}...${user.substring(user.length - 4)}` : user;
      authContainer.innerHTML = `
        <div class="user-badge">
          <span class="user-role-tag">${role === 'admin' ? 'مدیر ارشد' : 'عضو رسمی'}</span>
          <span class="user-id-text" title="${user}">${displayId}</span>
          <button class="btn-logout" onclick="AuthModule.logout()">خروج</button>
        </div>
      `;
    } else {
      authContainer.innerHTML = `
        <button class="btn btn-primary" onclick="AuthModule.openModal('login')">ورود / اتصال ولت</button>
      `;
    }
  }
};

document.addEventListener('DOMContentLoaded', () => {
  AuthModule.initUserState();
});
