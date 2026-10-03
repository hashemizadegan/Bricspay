landing_js = '''(function () {
  'use strict';

  var translations = {
    en: {
      lang_label: "Language",
      theme_toggle: "Toggle theme",
      cta_login: "Login",
      cta_app: "Open App",
      stats_assets: "Assets listed",
      stats_apy: "Staking APY",
      stats_continents: "Continents Served",
      email_label: "Email address",
      ph_email: "Example@google.com",
      btn_signup: "Sign Up",
      promo1_t: "Global settlements in BRICS currencies",
      promo1_d: "Send and receive payments across borders without the usual banking friction.",
      promo2_t: "Built for business and personal use",
      promo2_d: "Top-level security, fast processing and 24/7 support for every transaction.",
      steps_title: "Get Started with BricsPay in 5 Easy Steps",
      step1_t: "Register",
      step1_d: "Register an Account with Globiance by navigating to the “Sign up” button visible throughout this website.",
      step2_t: "Perform KYC",
      step2_d: "Complete the third party KYC on the Globiance platform and wait for Approval.",
      step3_t: "Set up 2FA",
      step3_d: "In the meantime take a moment to set up 2 Factor Authentication.",
      step4_t: "Deposit Funds",
      step4_d: "Deposit Fiat or Crypto from your Bank account or an external wallet.",
      step5_t: "Start BricsPay",
      step5_d: "Experience fast, secure, and easy banking with BricsPay.",
      trust1_t: "SUPPORT",
      trust1_d: "Our team is available 24/7 to help you.",
      trust2_t: "PRIVACY & SECURITY",
      trust2_d: "Your data and transactions are fully encrypted.",
      trust3_t: "SAFETY OF FUNDS",
      trust3_d: "User funds are 100% insured, with manual transaction reviews.",
      trust4_t: "2 FACTOR AUTHENTICATION",
      trust4_d: "2FA is required for every transaction to protect your assets.",
      feat_title: "Our Services",
      feat1_t: "BricsPay Cards",
      feat1_d: "Access your funds anytime, anywhere.",
      feat1_cta: "ORDER NOW",
      feat2_t: "Staking",
      feat2_d: "Earn rewards by staking your funds.",
      feat2_cta: "STAKE NOW",
      feat3_t: "Quick Swap",
      feat3_d: "Instant currency swaps made easy.",
      feat3_cta: "SWAP NOW",
      about_title: "About BricsPay",
      about_text: "BricsPay is a platform for borderless transactions, offering secure, fast, and easy global payments. We make sending and receiving money across borders simple, with top-level security and 24/7 support. Whether for personal or business use, BricsPay ensures a seamless, borderless banking experience for everyone.",
      cta_text: "a seamless, borderless banking experience for everyone.",
      cta_learn: "LEARN MORE",
      foot_about: "About",
      foot_about_l: "About",
      foot_terms: "Terms and Conditions",
      foot_privacy: "Privacy Policy",
      foot_data: "Data Policy",
      foot_cookies: "Cookies",
      foot_support: "Support",
      foot_ticket: "Submit Ticket",
      foot_security: "Security Policy",
      foot_contact: "Contact",
      foot_resources: "Resources",
      foot_api: "API",
      foot_docs: "Documentation",
      foot_fees: "Fees",
      foot_kyc: "KYC Requirements",
      foot_license: "BricsPay License",
      foot_follow: "Follow BricsPay on other platforms",
      msg_invalid_email: "Please enter a valid email address."
    },
    fa: {
      lang_label: "زبان",
      theme_toggle: "تغییر پوسته",
      cta_login: "ورود",
      cta_app: "ورود به سامانه",
      stats_assets: "دارایی‌های فهرست‌شده",
      stats_apy: "بازده استیکینگ سالانه",
      stats_continents: "قاره تحت پوشش",
      email_label: "نشانی ایمیل",
      ph_email: "Example@google.com",
      btn_signup: "ثبت‌نام",
      promo1_t: "تسویه جهانی با ارزهای کشورهای بریکس",
      promo1_d: "ارسال و دریافت پرداخت‌های برون‌مرزی بدون محدودیت‌ها و هزینه‌های اضافی بانکی سنتی.",
      promo2_t: "طراحی‌شده برای مصارف تجاری و فردی",
      promo2_d: "امنیت در بالاترین سطح، سرعت پردازش بالا و پشتیبانی ۲۴/۷ برای تمامی مبادلات.",
      steps_title: "شروع کار با بریکس‌پی در ۵ گام ساده",
      step1_t: "ثبت‌نام",
      step1_d: "با استفاده از دکمه «ثبت‌نام» در سایت، حساب کاربری خود را در پلتفرم ایجاد کنید.",
      step2_t: "احراز هویت (KYC)",
      step2_d: "مراحل احراز هویت را در سامانه تکمیل کرده و منتظر تأیید بمانید.",
      step3_t: "فعال‌سازی ۲FA",
      step3_d: "در این فاصله، تأیید دومرحله‌ای را برای ارتقای امنیت حساب خود فعال نمایید.",
      step4_t: "واریز موجودی",
      step4_d: "ارز فیات یا رمزارز را از حساب بانکی یا کیف پول خارجی خود واریز کنید.",
      step5_t: "آغاز مبادلات با بریکس‌پی",
      step5_d: "تجربه‌ای سریع، امن و آسان از بانکداری بین‌المللی با بریکس‌پی.",
      trust1_t: "پشتیبانی ۲۴/۷",
      trust1_d: "تیم پشتیبانی ما در تمام ساعات شبانه‌روز آماده پاسخگویی به شماست.",
      trust2_t: "حفظ حریم خصوصی و امنیت",
      trust2_d: "تمامی داده‌ها و تراکنش‌های شما به‌طور کامل رمزنگاری می‌شوند.",
      trust3_t: "امنیت ۱۰۰٪ دارایی‌ها",
      trust3_d: "دارایی کاربران تحت پوشش کامل بیمه و بررسی دقیق تراکنش‌ها قرار دارد.",
      trust4_t: "احراز هویت دومرحله‌ای",
      trust4_d: "استفاده از ۲FA برای هر تراکنش جهت حفاظت حداکثری الزامی است.",
      feat_title: "خدمات ما",
      feat1_t: "کارت‌های بریکس‌پی",
      feat1_d: "دسترسی به دارایی‌ها در هر زمان و هر مکان در سراسر جهان.",
      feat1_cta: "سفارش کارت",
      feat2_t: "استیکینگ",
      feat2_d: "کسب سود و پاداش با استیک کردن دارایی‌های رمزارزی و فیات.",
      feat2_cta: "استیک کنید",
      feat3_t: "تبدیل سریع (سواپ)",
      feat3_d: "تبدیل آنی و بدون دردسر ارزها با بهترین نرخ روز.",
      feat3_cta: "تبدیل آنی",
      about_title: "درباره بریکس‌پی",
      about_text: "بریکس‌پی یک پلتفرم پیشرفته برای مبادلات و تسویه‌های برون‌مرزی است که امکان پرداخت‌های جهانی امن، سریع و آسان را فراهم می‌کند. ما انتقال وجه بین‌المللی را با بالاترین استانداردهای امنیتی و پشتیبانی ۲۴ ساعته ساده کرده‌ایم تا تجربه‌ای یکپارچه و بدون مرز برای همه رقم بخورد.",
      cta_text: "تجربه‌ای یکپارچه و بدون مرز از بانکداری بین‌المللی برای همگان.",
      cta_learn: "اطلاعات بیشتر",
      foot_about: "درباره ما",
      foot_about_l: "درباره بریکس‌پی",
      foot_terms: "قوانین و مقررات",
      foot_privacy: "حریم خصوصی",
      foot_data: "سیاست داده‌ها",
      foot_cookies: "کوکی‌ها",
      foot_support: "پشتیبانی",
      foot_ticket: "ارسال تیکت",
      foot_security: "سیاست امنیت",
      foot_contact: "تماس با ما",
      foot_resources: "منابع",
      foot_api: "مستندات API",
      foot_docs: "راهنما و اسناد",
      foot_fees: "کارمزدها",
      foot_kyc: "الزامات احراز هویت",
      foot_license: "مجوز بریکس‌پی",
      foot_follow: "ما را در شبکه‌های اجتماعی دنبال کنید",
      msg_invalid_email: "لطفاً یک نشانی ایمیل معتبر وارد کنید."
    },
    ru: {
      lang_label: "Язык",
      theme_toggle: "Сменить тему",
      cta_login: "Войти",
      cta_app: "В приложение",
      stats_assets: "Активов доступно",
      stats_apy: "Годовая доходность (APY)",
      stats_continents: "Континентов охвачено",
      email_label: "Электронная почта",
      ph_email: "Example@google.com",
      btn_signup: "Регистрация",
      promo1_t: "Международные расчеты в валютах БРИКС",
      promo1_d: "Отправляйте и принимайте трансграничные платежи без лишних банковских барьеров.",
      promo2_t: "Для бизнеса и личного использования",
      promo2_d: "Максимальный уровень безопасности, быстрые расчеты и круглосуточная поддержка 24/7.",
      steps_title: "Начните работу с BricsPay за 5 простых шагов",
      step1_t: "Регистрация",
      step1_d: "Зарегистрируйте аккаунт в системе через кнопку «Регистрация» на этом сайте.",
      step2_t: "Верификация (KYC)",
      step2_d: "Пройдите верификацию личности на платформе и дождитесь одобрения.",
      step3_t: "Настройка 2FA",
      step3_d: "Настройте двухфакторную аутентификацию для защиты аккаунта.",
      step4_t: "Пополнение баланса",
      step4_d: "Внесите фиатные средства или криптовалюту с банковского счета или внешнего кошелька.",
      step5_t: "Запуск BricsPay",
      step5_d: "Начните пользоваться быстрыми, надежными и удобными сервисами BricsPay.",
      trust1_t: "ПОДДЕРЖКА 24/7",
      trust1_d: "Наша команда всегда готова помочь вам в любое время суток.",
      trust2_t: "КОНФИДЕНЦИАЛЬНОСТЬ И БЕЗОПАСНОСТЬ",
      trust2_d: "Ваши данные и финансовые транзакции полностью зашифрованы.",
      trust3_t: "БЕЗОПАСНОСТЬ СРЕДСТВ",
      trust3_d: "Средства пользователей застрахованы на 100% с ручным контролем транзакций.",
      trust4_t: "2-ФАКТОРНАЯ АУТЕНТИФИКАЦИЯ",
      trust4_d: "2FA обязательна для каждой транзакции для надежной защиты ваших активов.",
      feat_title: "Наши сервисы",
      feat1_t: "Карты BricsPay",
      feat1_d: "Доступ к вашим средствам в любое время и в любой точке мира.",
      feat1_cta: "ЗАКАЗАТЬ",
      feat2_t: "Стейкинг",
      feat2_d: "Получайте регулярное вознаграждение за стейкинг средств.",
      feat2_cta: "СТЕЙКИНГ",
      feat3_t: "Быстрый обмен",
      feat3_d: "Мгновенный обмен валют и криптоактивов по выгодному курсу.",
      feat3_cta: "ОБМЕНЯТЬ",
      about_title: "О платформе BricsPay",
      about_text: "BricsPay is a platform for borderless transactions, offering secure, fast, and easy global payments. We make sending and receiving money across borders simple, with top-level security and 24/7 support. Whether for personal or business use, BricsPay ensures a seamless, borderless banking experience for everyone.",
      cta_text: "a seamless, borderless banking experience for everyone.",
      cta_learn: "ПОДРОБНЕЕ",
      foot_about: "О нас",
      foot_about_l: "О платформе",
      foot_terms: "Условия использования",
      foot_privacy: "Политика конфиденциальности",
      foot_data: "Политика данных",
      foot_cookies: "Файлы cookie",
      foot_support: "Поддержка",
      foot_ticket: "Создать тикет",
      foot_security: "Политика безопасности",
      foot_contact: "Контакты",
      foot_resources: "Ресурсы",
      foot_api: "API документация",
      foot_docs: "Документация",
      foot_fees: "Тарифы и комиссии",
      foot_kyc: "Требования KYC",
      foot_license: "Лицензия BricsPay",
      foot_follow: "Мы в других платформах",
      msg_invalid_email: "Пожалуйста, введите корректный адрес электронной почты."
    }
  };

  function applyLanguage(lang) {
    if (!translations[lang]) {
      lang = 'en';
    }
    try {
      localStorage.setItem('bricspay_lang', lang);
    } catch (e) {}

    document.documentElement.lang = lang;
    document.documentElement.dir = (lang === 'fa') ? 'rtl' : 'ltr';

    var dict = translations[lang] || translations.en;

    var elms = document.querySelectorAll('[data-i18n]');
    for (var i = 0; i < elms.length; i++) {
      var key = elms[i].getAttribute('data-i18n');
      if (key === 'theme_toggle') {
        var currentTheme = document.documentElement.getAttribute('data-theme') || 'dark';
        elms[i].textContent = currentTheme === 'light' ? '☀️' : '🌙';
        continue;
      }
      if (dict[key] !== undefined) {
        elms[i].textContent = dict[key];
      }
    }

    var placeholders = document.querySelectorAll('[data-i18n-placeholder]');
    for (var j = 0; j < placeholders.length; j++) {
      var pKey = placeholders[j].getAttribute('data-i18n-placeholder');
      if (dict[pKey] !== undefined) {
        placeholders[j].setAttribute('placeholder', dict[pKey]);
      }
    }

    var sel = document.getElementById('langSelect');
    if (sel && sel.value !== lang) {
      sel.value = lang;
    }
  }

  function applyTheme(theme) {
    if (theme !== 'light' && theme !== 'dark') {
      theme = 'dark';
    }
    try {
      localStorage.setItem('bricspay_theme', theme);
    } catch (e) {}

    if (theme === 'light') {
      document.documentElement.setAttribute('data-theme', 'light');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }

    var btn = document.getElementById('themeBtn');
    if (btn) {
      btn.textContent = theme === 'light' ? '☀️' : '🌙';
    }
  }

  document.addEventListener('DOMContentLoaded', function () {
    var storedLang = 'en';
    try {
      storedLang = localStorage.getItem('bricspay_lang') || 'en';
    } catch (e) {}
    if (!translations[storedLang]) {
      storedLang = 'en';
    }
    applyLanguage(storedLang);

    var storedTheme = 'dark';
    try {
      storedTheme = localStorage.getItem('bricspay_theme') || 'dark';
    } catch (e) {}
    applyTheme(storedTheme);

    var langSel = document.getElementById('langSelect');
    if (langSel) {
      langSel.addEventListener('change', function (e) {
        applyLanguage(e.target.value);
      });
    }

    var themeBtn = document.getElementById('themeBtn');
    if (themeBtn) {
      themeBtn.addEventListener('click', function () {
        var current = document.documentElement.getAttribute('data-theme') === 'light' ? 'light' : 'dark';
        var next = current === 'light' ? 'dark' : 'light';
        applyTheme(next);
      });
    }

    var menuBtn = document.getElementById('menuBtn');
    var navLinks = document.getElementById('navLinks');
    if (menuBtn && navLinks) {
      menuBtn.addEventListener('click', function () {
        var open = navLinks.classList.toggle('open');
        menuBtn.setAttribute('aria-expanded', open ? 'true' : 'false');
      });
    }

    var yr = document.getElementById('year');
    if (yr) {
      yr.textContent = new Date().getFullYear();
    }

    var signupForm = document.getElementById('signupForm');
    if (signupForm) {
      signupForm.addEventListener('submit', function (e) {
        e.preventDefault();
        var emailInput = document.getElementById('signupEmail');
        var msg = document.getElementById('signupMsg');
        var val = emailInput ? emailInput.value.trim() : '';
        var emailRe = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!val || !emailRe.test(val)) {
          if (msg) {
            var currentLang = document.documentElement.lang || 'en';
            var dict = translations[currentLang] || translations.en;
            msg.textContent = dict.msg_invalid_email || 'Please enter a valid email address.';
          }
          if (emailInput) {
            emailInput.focus();
          }
          return;
        }
        if (msg) {
          msg.textContent = '';
        }
        var params = new URLSearchParams();
        params.set('auth', 'register');
        params.set('email', val);
        window.location.href = '/app?' + params.toString();
      });
    }
  });
})();
'''
import os
os.makedirs('/mnt/data/work/Bricspay-main/internal/api/static/js', exist_ok=True)
with open('/mnt/data/work/Bricspay-main/internal/api/static/js/landing.js', 'w', encoding='utf-8') as f:
    f.write(landing_js.strip() + '
')
print('wrote landing.js size:', os.path.getsize('/mnt/data/work/Bricspay-main/internal/api/static/js/landing.js'))
