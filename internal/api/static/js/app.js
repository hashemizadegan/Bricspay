const I18N = {
  fa: {
    brand_title: "سامانه یکپارچه تسویه فرامرزی BRICS Pay",
    nav_login: "ورود / ثبت‌نام",
    nav_logout: "خروج",
    tab_workflow: "معماری تسویه (Workflow)",
    tab_kyc: "پرونده تجاری و مدارک (KYC)",
    tab_admin: "پنل نظارتی ادمین (Audit)",
    workflow_heading: "مدل تسویه دوجانبه ایران - روسیه (VTB / میر-شتاب)",
    workflow_desc: "این سیستم تراکنش‌های مالی و اعتباری بین‌المللی را از طریق پیام‌رسان‌های مالی غیرسوئیفتی و تسویه مستقیم ارزهای ملی (روبل / ریال) بدون وابستگی به دلار مدیریت می‌کند.",
    kyc_title: "بارگذاری اسناد تجاری و احراز هویت (KYC)",
    kyc_doc_type: "نوع سند:",
    kyc_company_reg: "روزنامه رسمی / ثبت شرکت",
    kyc_license: "کارت بازرگانی / مجوز فعالیت",
    kyc_contract: "قرارداد تجاری / پروفورما",
    kyc_btn_upload: "ارسال سند جهت اعتبارسنجی",
    kyc_history_title: "تاریخچه اسناد بارگذاری شده",
    auth_title: "ورود به سامانه تسویه",
    auth_email: "ایمیل تجاری:",
    auth_password: "رمز عبور:",
    auth_company: "نام شرکت:",
    auth_country: "کشور ثبتی:",
    auth_role: "نقش سازمانی:",
    auth_btn_login: "ورود",
    auth_btn_register: "ثبت‌نام شرکت جدید",
    auth_switch_to_reg: "حساب ندارید؟ ثبت‌نام شرکت جدید",
    auth_switch_to_login: "قبلاً ثبت‌نام کرده‌اید؟ ورود به حساب"
  },
  en: {
    brand_title: "BRICS Pay Core Settlement Platform",
    nav_login: "Login / Register",
    nav_logout: "Logout",
    tab_workflow: "Settlement Workflow",
    tab_kyc: "KYC & Compliance Docs",
    tab_admin: "Admin Audit Panel",
    workflow_heading: "Iran - Russia Bilateral Settlement Framework (VTB / Mir-Shetab)",
    workflow_desc: "Direct multi-currency settlement platform managing cross-border transactions using national currencies (RUB / IRR) independent of legacy SWIFT mechanisms.",
    kyc_title: "Corporate KYC & Verification Documents",
    kyc_doc_type: "Document Type:",
    kyc_company_reg: "Certificate of Incorporation",
    kyc_license: "Trade License / Chamber ID",
    kyc_contract: "Commercial Contract / Proforma",
    kyc_btn_upload: "Upload Document for Verification",
    kyc_history_title: "Uploaded Documents History",
    auth_title: "Platform Authentication",
    auth_email: "Business Email:",
    auth_password: "Password:",
    auth_company: "Company Name:",
    auth_country: "Country of Registration:",
    auth_role: "Account Role:",
    auth_btn_login: "Sign In",
    auth_btn_register: "Register New Entity",
    auth_switch_to_reg: "No account? Register corporate entity",
    auth_switch_to_login: "Already registered? Sign in"
  },
  ru: {
    brand_title: "Платформа трансграничных расчетов BRICS Pay",
    nav_login: "Вход / Регистрация",
    nav_logout: "Выйти",
    tab_workflow: "Архитектура расчетов",
    tab_kyc: "Документы и верификация (KYC)",
    tab_admin: "Панель аудитора",
    workflow_heading: "Двусторонняя модель расчетов Россия — Иран (ВТБ / Мир-Шетаб)",
    workflow_desc: "Инфраструктура клиринга и прямых межбанковских расчетов в национальных валютах (RUB / IRR) без использования западных расчетных каналов.",
    kyc_title: "Загрузка корпоративных документов (KYC)",
    kyc_doc_type: "Тип документа:",
    kyc_company_reg: "Свидетельство о регистрации / ОГРН",
    kyc_license: "ВЭД лицензия / Выписка",
    kyc_contract: "Внешнеторговый контракт / Инвойс",
    kyc_btn_upload: "Отправить на проверку",
    kyc_history_title: "История загруженных документов",
    auth_title: "Авторизация в системе",
    auth_email: "Корпоративный Email:",
    auth_password: "Пароль:",
    auth_company: "Наименование организации:",
    auth_country: "Страна регистрации:",
    auth_role: "Роль учетной записи:",
    auth_btn_login: "Войти",
    auth_btn_register: "Зарегистрировать компанию",
    auth_switch_to_reg: "Нет аккаунта? Регистрация",
    auth_switch_to_login: "Уже зарегистрированы? Войти"
  }
};

const App = {
  state: {
    lang: localStorage.getItem('brics_lang') || 'fa',
    token: localStorage.getItem('brics_token') || null,
    user: JSON.parse(localStorage.getItem('brics_user') || 'null')
  },

  init() {
    this.bindEvents();
    this.applyLanguage(this.state.lang);
    this.updateAuthUI();
  },

  bindEvents() {
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const targetTab = e.currentTarget.getAttribute('data-tab');
        if (targetTab) this.switchTab(targetTab);
      });
    });

    const langSel = document.getElementById('langSelect');
    if (langSel) {
      langSel.value = this.state.lang;
      langSel.addEventListener('change', (e) => {
        this.applyLanguage(e.target.value);
      });
    }
  },

  applyLanguage(lang) {
    if (!I18N[lang]) lang = 'fa';
    this.state.lang = lang;
    localStorage.setItem('brics_lang', lang);

    document.documentElement.lang = lang;
    document.documentElement.dir = (lang === 'fa') ? 'rtl' : 'ltr';

    const langSel = document.getElementById('langSelect');
    if (langSel && langSel.value !== lang) {
      langSel.value = lang;
    }

    const dict = I18N[lang];
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.getAttribute('data-i18n');
      if (dict[key]) {
        el.textContent = dict[key];
      }
    });
  },

  switchTab(tabName) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.toggle('active', btn.getAttribute('data-tab') === tabName);
    });

    document.querySelectorAll('.tab-content').forEach(section => {
      section.classList.remove('active');
    });

    const targetSection = document.getElementById(`tab-${tabName}`);
    if (targetSection) {
      targetSection.classList.add('active');
    }

    if (tabName === 'kyc' && window.Kyc) {
      Kyc.loadHistory();
    }
    if (tabName === 'admin' && window.Admin) {
      Admin.loadAll();
    }
  },

  updateAuthUI() {
    const authBtn = document.getElementById('authActionBtn');
    const userDisplay = document.getElementById('userDisplay');
    const adminTabBtn = document.querySelector('.tab-btn[data-tab="admin"]');

    if (this.state.token && this.state.user) {
      if (authBtn) authBtn.style.display = 'none';
      if (userDisplay) {
        userDisplay.style.display = 'flex';
        userDisplay.innerHTML = `
          <span>👤 ${this.state.user.company_name || this.state.user.email} (${this.state.user.role || 'Member'})</span>
          <button id="logoutBtn" class="btn" style="background:#ef4444; border:none; padding:4px 8px; font-size:12px; margin-right:8px; margin-left:8px;" data-i18n="nav_logout">${I18N[this.state.lang].nav_logout}</button>
        `;
        const logoutBtn = document.getElementById('logoutBtn');
        if (logoutBtn) logoutBtn.addEventListener('click', () => Auth.logout());
      }

      if (adminTabBtn && (this.state.user.role === 'admin' || this.state.user.role === 'auditor')) {
        adminTabBtn.style.display = 'inline-block';
      }
    } else {
      if (authBtn) authBtn.style.display = 'inline-block';
      if (userDisplay) userDisplay.style.display = 'none';
      if (adminTabBtn) adminTabBtn.style.display = 'none';
    }
  },

  async api(endpoint, options = {}) {
    const headers = options.headers || {};
    if (this.state.token) {
      headers['Authorization'] = `Bearer ${this.state.token}`;
    }
    if (!options.isFormData && !headers['Content-Type']) {
      headers['Content-Type'] = 'application/json';
    }

    const config = {
      ...options,
      headers
    };

    const response = await fetch(endpoint, config);
    if (response.status === 401) {
      if (window.Auth) Auth.logout();
      throw new Error('نشست شما منقضی شده است. مجدداً وارد شوید.');
    }
    return response;
  }
};

document.addEventListener('DOMContentLoaded', () => {
  App.init();
});
