"use strict";

const I18N = {
  fa: {
    nav_status: "وضعیت شبکه: عملیاتی",
    ticker_banner: "سامانه تسویه چندارزی برون‌مرزی • کریدور فعال: RUB / CNY / AED / IRR • مطابق ISO 20022",
    tab_overview: "نمای کلی تسویه",
    tab_workflow: "گردش کار تجاری",
    tab_kyc: "احراز هویت شرکتی",
    tab_admin: "انطباق / مدیریت",
    overview_title: "درگاه تسویه چندارزی",
    overview_desc: "تسویه برخط برای تراکنش‌های تجاری دوجانبه و برون‌مرزی.",
    workflow_title: "تأمین مالی تجاری و تسویه",
    workflow_desc: "مسیر مستندسازی، تسویه و امانی برای تجارت برون‌مرزی.",
    kyc_title: "ثبت‌نام سازمان و احراز هویت",
    kyc_guest: "برای بارگذاری امن مدارک، ابتدا وارد شوید.",
    kyc_doc_category: "دسته‌بندی مدرک",
    kyc_registration: "گواهی ثبت شرکت",
    kyc_tax: "گواهی مالیاتی / شناسه مالیاتی",
    kyc_bank: "صورت‌حساب بانکی",
    kyc_upload_label: "بارگذاری فایل (PDF، PNG، JPG)",
    kyc_upload_button: "بارگذاری مدرک",
    admin_title: "پنل بررسی انطباق",
    admin_entity: "شرکت",
    admin_jurisdiction: "حوزه قضایی",
    admin_status: "وضعیت",
    admin_actions: "عملیات",
    admin_empty: "موردی برای بررسی وجود ندارد",
    auth_open: "ورود / ثبت‌نام",
    auth_logout: "خروج",
    auth_title_login: "ورود",
    auth_title_register: "ثبت‌نام شرکت یا بانک",
    auth_email: "نشانی ایمیل",
    auth_password: "گذرواژه",
    auth_legal_name: "نام حقوقی شرکت",
    auth_jurisdiction: "حوزه قضایی (مانند IR، RU، CN، AE)",
    auth_submit_login: "ورود",
    auth_submit_register: "ثبت‌نام و ایجاد پروفایل",
    auth_register_prompt: "حساب ندارید؟ ثبت‌نام کنید",
    auth_login_prompt: "قبلاً ثبت‌نام کرده‌اید؟ وارد شوید",
    auth_metamask: "ورود با MetaMask",
    ticker_event1: "تسویه موفق پارت دوم محموله غلات با نماد RUB/IRR به ارزش ۲۴۰ میلیون روبل از طریق کارگزاری VTB.",
    ticker_event2: "اتصال پروتکل پیام‌رسان مالی SPFS به سامانه تسویه با موفقیت آزمایش شد.",
    ticker_event3: "مدارک شرکت بازرگانی پتروشیمی اروند توسط واحد ارزیابی انطباق تأیید شد.",
    ticker_event4: "کارمزد تسویه ارزی در معاملات کریدور شمال-جنوب به ۰.۱۵٪ کاهش یافت."
  },

  en: {
    nav_status: "Network Status: Operational",
    ticker_banner: "Cross-Border Multi-Currency Settlement Engine • Active Corridor: RUB / CNY / AED / IRR • ISO 20022 aligned",
    tab_overview: "Settlement Overview",
    tab_workflow: "Trade Workflow",
    tab_kyc: "Corporate KYC",
    tab_admin: "Compliance / Admin",
    overview_title: "Multi-Currency Settlement Gateway",
    overview_desc: "Real-time settlement for bilateral and cross-border trade transactions.",
    workflow_title: "Trade Finance and Settlement",
    workflow_desc: "End-to-end documentation, clearing, and escrow pathway for cross-border trade.",
    kyc_title: "Organization Onboarding & KYC",
    kyc_guest: "Please sign in to access secure document submission.",
    kyc_doc_category: "Document Category",
    kyc_registration: "Registration Certificate",
    kyc_tax: "Tax Certificate / TIN",
    kyc_bank: "Bank Statement",
    kyc_upload_label: "Upload File (PDF, PNG, JPG)",
    kyc_upload_button: "Upload Document",
    admin_title: "Compliance Review Panel",
    admin_entity: "Entity",
    admin_jurisdiction: "Jurisdiction",
    admin_status: "Status",
    admin_actions: "Actions",
    admin_empty: "No pending verifications",
    auth_open: "Sign In / Register",
    auth_logout: "Sign Out",
    auth_title_login: "Sign In",
    auth_title_register: "Register Legal Entity / Bank",
    auth_email: "Email Address",
    auth_password: "Password",
    auth_legal_name: "Legal Entity Name",
    auth_jurisdiction: "Jurisdiction (e.g. IR, RU, CN, AE)",
    auth_submit_login: "Sign In",
    auth_submit_register: "Register & Create Profile",
    auth_register_prompt: "No account yet? Register here",
    auth_login_prompt: "Already registered? Sign in",
    auth_metamask: "Login with MetaMask",
    ticker_event1: "Successful settlement of a grain shipment tranche valued at 240 million rubles via VTB clearing.",
    ticker_event2: "SPFS financial messaging connection to the settlement system was successfully tested.",
    ticker_event3: "Arvand Petrochemical Trading documents were approved by compliance review.",
    ticker_event4: "Foreign-exchange settlement fees on the North-South corridor were reduced to 0.15%."
  },

  ru: {
    nav_status: "Статус сети: работает",
    ticker_banner: "Мультивалютная система трансграничных расчетов • Активный коридор: RUB / CNY / AED / IRR • ISO 20022",
    tab_overview: "Обзор расчетов",
    tab_workflow: "Торговый процесс",
    tab_kyc: "Корпоративный KYC",
    tab_admin: "Комплаенс / Администрирование",
    overview_title: "Мультивалютный расчетный шлюз",
    overview_desc: "Расчеты в реальном времени по двусторонним и трансграничным торговым операциям.",
    workflow_title: "Торговое финансирование и расчеты",
    workflow_desc: "Полный цикл документооборота, клиринга и эскроу для трансграничной торговли.",
    kyc_title: "Регистрация организации и KYC",
    kyc_guest: "Войдите в систему для защищенной отправки документов.",
    kyc_doc_category: "Категория документа",
    kyc_registration: "Свидетельство о регистрации",
    kyc_tax: "Налоговый сертификат / ИНН",
    kyc_bank: "Банковская выписка",
    kyc_upload_label: "Загрузить файл (PDF, PNG, JPG)",
    kyc_upload_button: "Загрузить документ",
    admin_title: "Панель проверки комплаенса",
    admin_entity: "Организация",
    admin_jurisdiction: "Юрисдикция",
    admin_status: "Статус",
    admin_actions: "Действия",
    admin_empty: "Нет заявок на проверку",
    auth_open: "Вход / Регистрация",
    auth_logout: "Выйти",
    auth_title_login: "Вход",
    auth_title_register: "Регистрация компании или банка",
    auth_email: "Адрес электронной почты",
    auth_password: "Пароль",
    auth_legal_name: "Юридическое название компании",
    auth_jurisdiction: "Юрисдикция (например, IR, RU, CN, AE)",
    auth_submit_login: "Войти",
    auth_submit_register: "Зарегистрировать и создать профиль",
    auth_register_prompt: "Нет аккаунта? Зарегистрируйтесь",
    auth_login_prompt: "Уже зарегистрированы? Войдите",
    auth_metamask: "Войти через MetaMask",
    ticker_event1: "Успешно проведен расчет по партии зерна на сумму 240 млн рублей через клиринг ВТБ.",
    ticker_event2: "Связь финансовых сообщений SPFS с расчетной системой успешно протестирована.",
    ticker_event3: "Документы компании Arvand Petrochemical одобрены службой комплаенса.",
    ticker_event4: "Комиссия за валютные расчеты по коридору Север—Юг снижена до 0,15%."
  }
};

function storageGet(key) {
  try {
    return window.localStorage.getItem(key);
  } catch (error) {
    return null;
  }
}

function storageSet(key, value) {
  try {
    window.localStorage.setItem(key, value);
  } catch (error) {
    // The application can continue if browser storage is unavailable.
  }
}

function storageRemove(key) {
  try {
    window.localStorage.removeItem(key);
  } catch (error) {
    // The application can continue if browser storage is unavailable.
  }
}

function readStoredUser() {
  const value = storageGet("bricspay_user");

  if (!value) {
    return null;
  }

  try {
    return JSON.parse(value);
  } catch (error) {
    return value;
  }
}

function getInitialLanguage() {
  const savedLanguage = storageGet("bricspay_lang");
  return Object.prototype.hasOwnProperty.call(I18N, savedLanguage)
    ? savedLanguage
    : "fa";
}

const App = {
  state: {
    lang: getInitialLanguage(),
    currentTab: "overview",
    token: storageGet("bricspay_token"),
    user: readStoredUser()
  },

  api(path, options = {}) {
    const token = storageGet("bricspay_token");
    const headers = new Headers(options.headers || {});
    const isFormData =
      typeof FormData !== "undefined" && options.body instanceof FormData;

    if (options.body && !isFormData && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }

    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    return fetch(path, {
      ...options,
      headers,
      credentials: "include"
    });
  },

  t(key) {
    const dictionary = I18N[this.state.lang] || I18N.fa;
    return dictionary[key] || I18N.en[key] || key;
  },

  init() {
    this.initTabs();
    this.initLanguageSelector();
    this.initKYCForm();
    this.applyLanguage(this.state.lang);
    this.updateAuthUI();
    this.switchTab(this.state.currentTab);

    if (window.NewsTickerModule) {
      window.NewsTickerModule.init(this.state.lang);
    }
  },

  initTabs() {
    document.querySelectorAll(".nav-btn[data-tab]").forEach((button) => {
      button.addEventListener("click", (event) => {
        event.preventDefault();
        this.switchTab(button.dataset.tab);
      });
    });
  },

  initLanguageSelector() {
    const select = document.getElementById("langSelect");

    if (!select) {
      return;
    }

    select.value = this.state.lang;

    select.addEventListener("change", (event) => {
      this.applyLanguage(event.target.value);

      if (window.NewsTickerModule) {
        window.NewsTickerModule.init(this.state.lang);
      }
    });
  },

  initKYCForm() {
    const form = document.getElementById("kycUploadForm");

    if (!form) {
      return;
    }

    form.addEventListener("submit", (event) => {
      event.preventDefault();

      if (
        window.KYCModule &&
        typeof window.KYCModule.uploadDocument === "function"
      ) {
        window.KYCModule.uploadDocument(event);
      } else {
        console.error("KYCModule.uploadDocument is not available.");
      }
    });
  },

  applyLanguage(language) {
    const selectedLanguage = Object.prototype.hasOwnProperty.call(I18N, language)
      ? language
      : "fa";

    this.state.lang = selectedLanguage;
    storageSet("bricspay_lang", selectedLanguage);

    document.documentElement.lang = selectedLanguage;
    document.documentElement.dir = selectedLanguage === "fa" ? "rtl" : "ltr";

    const dictionary = I18N[selectedLanguage];

    document.querySelectorAll("[data-i18n]").forEach((element) => {
      const key = element.dataset.i18n;

      if (Object.prototype.hasOwnProperty.call(dictionary, key)) {
        element.textContent = dictionary[key];
      }
    });

    const select = document.getElementById("langSelect");
    if (select) {
      select.value = selectedLanguage;
    }

    this.updateAuthUI();

    if (window.AuthModule && typeof window.AuthModule.refreshLanguage === "function") {
      window.AuthModule.refreshLanguage();
    }
  },

  switchTab(tabName) {
    if (!tabName) {
      return;
    }

    const target = document.getElementById(`tab-${tabName}`);

    if (!target || !target.classList.contains("tab-pane")) {
      console.warn(`Tab panel not found: tab-${tabName}`);
      return;
    }

    this.state.currentTab = tabName;

    document.querySelectorAll(".nav-btn[data-tab]").forEach((button) => {
      const isActive = button.dataset.tab === tabName;
      button.classList.toggle("active", isActive);
      button.setAttribute("aria-selected", String(isActive));
    });

    document.querySelectorAll(".tab-pane").forEach((pane) => {
      const isActive = pane === target;
      pane.classList.toggle("active", isActive);
      pane.hidden = !isActive;
      pane.style.display = isActive ? "block" : "none";
    });

    if (
      tabName === "admin" &&
      window.AdminModule &&
      typeof window.AdminModule.loadKYCList === "function"
    ) {
      window.AdminModule.loadKYCList();
    }
  },

  updateAuthUI() {
    const token = storageGet("bricspay_token");
    const user = readStoredUser();

    this.state.token = token;
    this.state.user = user;

    const loginButton = document.getElementById("authOpenBtn");
    const logoutButton = document.getElementById("logoutBtn");
    const isLoggedIn = Boolean(token);

    if (loginButton) {
      loginButton.textContent = isLoggedIn ? this.t("auth_open") : this.t("auth_open");
      loginButton.classList.toggle("hidden", isLoggedIn);
    }

    if (logoutButton) {
      logoutButton.textContent = this.t("auth_logout");
      logoutButton.classList.toggle("hidden", !isLoggedIn);
    }

    const guestNotice = document.getElementById("kycGuestNotice");
    const kycForm = document.getElementById("kycFormContainer");

    if (guestNotice) {
      guestNotice.classList.toggle("hidden", isLoggedIn);
    }

    if (kycForm) {
      kycForm.classList.toggle("hidden", !isLoggedIn);
    }
  },

  clearAuth() {
    storageRemove("bricspay_token");
    storageRemove("bricspay_user");
    storageRemove("bricspay_role");
    this.updateAuthUI();
  }
};

window.App = App;
window.I18N = I18N;

const NewsTickerModule = {
  items: [],
  currentIndex: 0,
  timer: null,

  init(language) {
    const dictionary = I18N[language] || I18N.fa;
    this.items = [
      dictionary.ticker_event1,
      dictionary.ticker_event2,
      dictionary.ticker_event3,
      dictionary.ticker_event4
    ].filter(Boolean);

    this.currentIndex = 0;
    this.render();
    this.startCycle();
  },

  render() {
    const ticker = document.getElementById("tickerContent");

    if (!ticker || this.items.length === 0) {
      return;
    }

    ticker.textContent = this.items[this.currentIndex];
  },

  startCycle() {
    if (this.timer) {
      window.clearInterval(this.timer);
    }

    this.timer = window.setInterval(() => {
      if (this.items.length === 0) {
        return;
      }

      this.currentIndex = (this.currentIndex + 1) % this.items.length;
      this.render();
    }, 6000);
  }
};

window.NewsTickerModule = NewsTickerModule;

document.addEventListener("DOMContentLoaded", () => {
  App.init();
});
