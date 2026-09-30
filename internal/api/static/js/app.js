const I18N = {
  fa: {
    brand_title: "تسویه مرکزی بریکس پی",
    nav_login: "ورود / ثبت‌نام",
    nav_logout: "خروج",
    tab_workflow: "معماری تسویه (Workflow)",
    tab_kyc: "پرونده تجاری و KYC",
    tab_admin: "پنل مدیریت (Admin)",
    workflow_title: "مدل تسویه دوجانبه ایران - روسیه (میر / VTB - شتاب)",
    workflow_desc: "این سیستم تراکنش‌های مالی و اعتباری بین‌المللی را از طریق پیام‌رسان‌های مالی غیرسوئیفتی مدیریت می‌کند.",
    diagram_heading: "فرایند ۹ مرحله‌ای تسویه و اعتبارات اسنادی (VTB Trade Finance)",
    kyc_title: "ثبت پرونده تجاری بین‌الملل (KYC)",
    admin_title: "پنل مدیریت اعتبارات اسنادی و نظارت",
    auth_login_title: "ورود به سامانه تسویه",
    auth_register_title: "ثبت‌نام شرکت / بازرگان",
    auth_email_label: "پست الکترونیک تجاری",
    auth_password_label: "رمز عبور امن",
    auth_company_label: "نام شرکت یا مؤسسه",
    auth_role_label: "نوع دسترسی سامانه",
    auth_btn_login: "ورود به حساب",
    auth_btn_register: "ثبت حساب کاربری",
    auth_switch_to_register: "حساب کاربری ندارید؟ ثبت‌نام کنید",
    auth_switch_to_login: "قبلاً ثبت‌نام کرده‌اید؟ ورود به حساب"
  },
  en: {
    brand_title: "BRICS Pay Core Settlement",
    nav_login: "Login / Register",
    nav_logout: "Logout",
    tab_workflow: "Settlement Workflow",
    tab_kyc: "Commercial KYC",
    tab_admin: "Admin Console",
    workflow_title: "Iran-Russia Bilateral Settlement Flow (MIR / VTB - Shetab)",
    workflow_desc: "Non-SWIFT international financial messaging and bilateral settlement platform.",
    diagram_heading: "9-Step Trade Finance & Settlement Workflow (VTB Platform)",
    kyc_title: "International Trade Profile Registration (KYC)",
    admin_title: "Letter of Credit & Oversight Administration",
    auth_login_title: "Sign in to Settlement Gateway",
    auth_register_title: "Commercial Entity Registration",
    auth_email_label: "Corporate Email Address",
    auth_password_label: "Secure Password",
    auth_company_label: "Company / Institution Name",
    auth_role_label: "System Access Tier",
    auth_btn_login: "Sign In",
    auth_btn_register: "Register Account",
    auth_switch_to_register: "Need an account? Register now",
    auth_switch_to_login: "Already registered? Sign in"
  },
  ru: {
    brand_title: "Клиринговый Центр BRICS Pay",
    nav_login: "Вход / Регистрация",
    nav_logout: "Выход",
    tab_workflow: "Архитектура расчетов",
    tab_kyc: "Торговое досье (KYC)",
    tab_admin: "Панель администратора",
    workflow_title: "Двусторонняя модель расчетов Иран – Россия (МИР / ВТБ – Шетаб)",
    workflow_desc: "Система международных финансовых расчетов без использования SWIFT и доллара.",
    diagram_heading: "9-этапный процесс торгового финансирования и расчетов (ВТБ)",
    kyc_title: "Регистрация торгового профиля (KYC)",
    admin_title: "Администрирование аккредитивов и надзор",
    auth_login_title: "Вход в клиринговую систему",
    auth_register_title: "Регистрация компании / импортера",
    auth_email_label: "Корпоративная эл. почта",
    auth_password_label: "Безопасный пароль",
    auth_company_label: "Название компании",
    auth_role_label: "Уровень доступа",
    auth_btn_login: "Войти",
    auth_btn_register: "Зарегистрироваться",
    auth_switch_to_register: "Нет аккаунта? Зарегистрироваться",
    auth_switch_to_login: "Уже зарегистрированы? Войти"
  }
};

const App = {
  state: {
    lang: localStorage.getItem('bricspay_lang') || 'fa',
    currentTab: 'workflow'
  },

  init() {
    this.initTabs();
    this.initLang();
    this.applyLanguage(this.state.lang);
    this.updateAuthUI();
    NewsTickerModule.init(this.state.lang);
  },

  initTabs() {
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.addEventListener('click', (e) => {
        const targetTab = e.currentTarget.getAttribute('data-tab');
        if (targetTab) this.switchTab(targetTab);
      });
    });
  },

  initLang() {
    const langSel = document.getElementById('langSelect');
    if (langSel) {
      langSel.value = this.state.lang;
      langSel.addEventListener('change', (e) => {
        this.applyLanguage(e.target.value);
      });
    }
  },

  applyLanguage(lang) {
    this.state.lang = lang;
    localStorage.setItem('bricspay_lang', lang);

    document.documentElement.lang = lang;
    document.documentElement.dir = (lang === 'fa') ? 'rtl' : 'ltr';

    const dict = I18N[lang] || I18N.en;
    document.querySelectorAll('[data-i18n]').forEach(el => {
      const key = el.getAttribute('data-i18n');
      if (dict[key]) {
        el.textContent = dict[key];
      }
    });

    NewsTickerModule.render(lang);
  },

  switchTab(tabName) {
    this.state.currentTab = tabName;

    // تنظیم وضعیت دکمه‌های ناوبری تب
    document.querySelectorAll('.tab-btn').forEach(btn => {
      btn.classList.toggle('active', btn.getAttribute('data-tab Ward') === tabName || btn.getAttribute('data-tab') === tabName);
    });

    // نمایش پنل مرتبط و مخفی‌سازی سایر پنل‌ها
    document.querySelectorAll('.tab-pane').forEach(pane => {
      pane.classList.remove('active');
    });

    const targetPane = document.getElementById(`tab-${tabName}`);
    if (targetPane) {
      targetPane.classList.add('active');
    }

    // هماهنگی با ماژول‌های KYC و Admin با نام صحیح
    if (tabName === 'kyc' && window.KYCModule) {
      if (typeof KYCModule.checkStatus === 'function') KYCModule.checkStatus();
    }
    if (tabName === 'admin' && window.AdminModule) {
      if (typeof AdminModule.loadData === 'function') AdminModule.loadData();
    }
  },

  updateAuthUI() {
    const token = localStorage.getItem('token');
    const userJson = localStorage.getItem('user');
    const authBtn = document.getElementById('authNavBtn');
    const userInfo = document.getElementById('userNavInfo');
    const userEmailSpan = document.getElementById('navUserEmail');

    if (token && userJson) {
      const user = JSON.parse(userJson);
      if (authBtn) authBtn.classList.add('hidden');
      if (userInfo) userInfo.classList.remove('hidden');
      if (userEmailSpan) userEmailSpan.textContent = user.company_name || user.email;
    } else {
      if (authBtn) authBtn.classList.remove('hidden');
      if (userInfo) userInfo.classList.add('hidden');
    }

    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn && !logoutBtn.dataset.bound) {
      logoutBtn.dataset.bound = "true";
      logoutBtn.addEventListener('click', () => {
        if (window.AuthModule) {
          AuthModule.logout();
        } else {
          localStorage.removeItem('token');
          localStorage.removeItem('user');
          location.reload();
        }
      });
    }
  }
};

const NewsTickerModule = {
  feeds: {
    fa: [
      "فاز دوم پایلوت تسویه فرامرزی BRICS Pay فعال شد.",
      "اتصال مستقیم شبکه‌های شتاب و میر روسیه با موفقیت برقرار گردید.",
      "تأمین نقدینگی و برابری ریال-روبل در پلتفرم تسویه به‌روزرسانی شد."
    ],
    en: [
      "BRICS Pay Cross-Border Settlement pilot reaches phase 2.",
      "MIR and Shetab payment networks successfully interconnected.",
      "Liquidity pool updated with direct RUB/IRR cross-rate parity."
    ],
    ru: [
      "Пилотный проект трансграничных расчетов BRICS Pay перешел во вторую фазу.",
      "Успешно завершена интеграция платежных систем МИР и Шетаб.",
      "Обновлен пул ликвидности для прямых клиринговых расчетов в паре RUB/IRR."
    ]
  },
  init(lang) {
    this.render(lang);
  },
  render(lang) {
    const track = document.getElementById('tickerTrack') || document.querySelector('.ticker-track');
    if (!track) return;
    const items = this.feeds[lang] || this.feeds.en;
    track.innerHTML = items.map(t => `<span class="ticker-item">${t}</span>`).join('');
  }
};

document.addEventListener('DOMContentLoaded', () => {
  App.init();
});
