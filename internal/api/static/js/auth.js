"use strict";

function authStorageGet(key) {
  try {
    return window.localStorage.getItem(key);
  } catch (error) {
    return null;
  }
}

function authStorageSet(key, value) {
  try {
    window.localStorage.setItem(key, value);
  } catch (error) {
    // Authentication can report the problem if browser storage is unavailable.
  }
}

function authStorageRemove(key) {
  try {
    window.localStorage.removeItem(key);
  } catch (error) {
    // Continue with logout even if browser storage is unavailable.
  }
}

async function readResponse(response) {
  const text = await response.text();

  if (!text) {
    return {};
  }

  try {
    return JSON.parse(text);
  } catch (error) {
    return { error: text };
  }
}

const AuthModule = {
  mode: "login",

  init() {
    const openButton = document.getElementById("authOpenBtn");
    const logoutButton = document.getElementById("logoutBtn");
    const form = document.getElementById("authForm");
    const walletButton = document.getElementById("metaMaskBtn");

    if (openButton) {
      openButton.addEventListener("click", () => this.openModal("login"));
    }

    if (logoutButton) {
      logoutButton.addEventListener("click", () => this.logout());
    }

    document.querySelectorAll("[data-auth-close]").forEach((button) => {
      button.addEventListener("click", () => this.closeModal());
    });

    if (form) {
      form.addEventListener("submit", (event) => {
        event.preventDefault();
        this.handleFormSubmit();
      });
    }

    if (walletButton) {
      walletButton.addEventListener("click", () => this.loginWithMetaMask());
    }

    const modal = document.getElementById("authModal");
    if (modal) {
      modal.addEventListener("click", (event) => {
        if (event.target === modal) {
          this.closeModal();
        }
      });
    }

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape") {
        this.closeModal();
      }
    });

    this.refreshLanguage();
  },

  openModal(mode = "login") {
    const modal = document.getElementById("authModal");

    if (modal) {
      modal.classList.remove("hidden");
      modal.classList.add("active");
      modal.setAttribute("aria-hidden", "false");
      modal.style.display = "flex";
    }

    this.switchMode(mode);
  },

  closeModal() {
    const modal = document.getElementById("authModal");

    if (modal) {
      modal.classList.add("hidden");
      modal.classList.remove("active");
      modal.setAttribute("aria-hidden", "true");
      modal.style.display = "none";
    }

    this.clearStatus();
  },

  switchMode(mode) {
    this.mode = mode === "register" ? "register" : "login";
    this.refreshLanguage();
    this.clearStatus();
  },

  refreshLanguage() {
    const app = window.App;
    const isLogin = this.mode === "login";
    const title = document.getElementById("authTitle");
    const registerFields = document.getElementById("registerFields");
    const submitButton = document.getElementById("authSubmitBtn");
    const switchContainer = document.getElementById("authSwitchText");
    const GAPGPTMASKTOKENxmrefzcki1dX0X = document.getElementById("authPassword");

    const text = (key, fallback) => {
      return app && typeof app.t === "function" ? app.t(key) : fallback;
    };

    if (title) {
      title.textContent = text(
        isLogin ? "auth_title_login" : "auth_title_register",
        isLogin ? "Sign In" : "Register"
      );
    }

    if (submitButton) {
      submitButton.textContent = text(
        isLogin ? "auth_submit_login" : "auth_submit_register",
        isLogin ? "Sign In" : "Register"
      );
    }

    if (registerFields) {
      registerFields.classList.toggle("hidden", isLogin);
      registerFields.style.display = isLogin ? "none" : "block";
    }

    if (GAPGPTMASKTOKENxmrefzcki1dX1X) {
      GAPGPTMASKTOKENxmrefzcki1dX2X.autocomplete = isLogin
        ? "current-password"
        : "new-password";
    }

    if (switchContainer) {
      switchContainer.replaceChildren();

      const link = document.createElement("button");
      link.type = "button";
      link.className = "auth-switch-link";
      link.textContent = text(
        isLogin ? "auth_register_prompt" : "auth_login_prompt",
        isLogin ? "Register" : "Sign in"
      );

      link.addEventListener("click", () => {
        this.switchMode(isLogin ? "register" : "login");
      });

      switchContainer.appendChild(link);
    }
  },

  clearStatus() {
    const status = document.getElementById("authStatusMessage");

    if (status) {
      status.textContent = "";
      status.className = "auth-status";
    }
  },

  showStatus(message, isError = false) {
    const status = document.getElementById("authStatusMessage");

    if (!status) {
      window.alert(message);
      return;
    }

    status.textContent = message;
    status.className = isError
      ? "auth-status error"
      : "auth-status success";
  },

  async handleFormSubmit() {
    const emailInput = document.getElementById("authEmail");
    const GAPGPTMASKTOKENxmrefzcki1dX4X = document.getElementById("authPassword");
    const legalNameInput = document.getElementById("authLegalName");
    const jurisdictionInput = document.getElementById("authJurisdiction");
    const submitButton = document.getElementById("authSubmitBtn");

    const email = emailInput ? emailInput.value.trim() : "";
    const password = GAPGPTMASKTOKENxmrefzcki1dX5X ? GAPGPTMASKTOKENxmrefzcki1dX6X.value : "";

    if (!email || !password) {
      this.showStatus("Please enter your email and password.", true);
      return;
    }

    if (
      this.mode === "register" &&
      (!legalNameInput || !legalNameInput.value.trim())
    ) {
      this.showStatus("Please enter the legal entity name.", true);
      return;
    }

    if (submitButton) {
      submitButton.disabled = true;
    }

    try {
      this.showStatus("Processing request...", false);

      if (this.mode === "login") {
        const response = await fetch("/api/v1/auth/login", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ email, password })
        });

        const data = await readResponse(response);

        if (!response.ok) {
          throw new Error(
            data.error || data.message || "Login failed."
          );
        }

        if (!data.token) {
          throw new Error(
            "The server response did not include an authentication token."
          );
        }

        this.saveAuth(data.token, data.role || "member", email);
        return;
      }

      const payload = {
        email,
        password,
        legal_name: legalNameInput
          ? legalNameInput.value.trim()
          : "",
        jurisdiction: jurisdictionInput
          ? jurisdictionInput.value.trim()
          : ""
      };

      const response = await fetch("/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify(payload)
      });

      const data = await readResponse(response);

      if (!response.ok) {
        throw new Error(
          data.error || data.message || "Registration failed."
        );
      }

      this.showStatus(
        "Registration successful. Please sign in.",
        false
      );

      window.setTimeout(() => this.switchMode("login"), 1200);
    } catch (error) {
      this.showStatus(
        error.message || "Authentication request failed.",
        true
      );
    } finally {
      if (submitButton) {
        submitButton.disabled = false;
      }
    }
  },

  saveAuth(token, role = "member", email = "") {
    if (typeof token !== "string" || token.trim() === "") {
      throw new Error("The server did not return a valid authentication token.");
    }

    const normalizedRole =
      typeof role === "string" && role.trim()
        ? role.trim()
        : "member";

    const user = {
      email: typeof email === "string" ? email.trim() : "",
      role: normalizedRole
    };

    authStorageSet("bricspay_user", JSON.stringify(user));
    authStorageSet("bricspay_role", normalizedRole);
    authStorageSet("bricspay_token", token);

    if (authStorageGet("bricspay_token") !== token) {
      authStorageRemove("bricspay_token");
      authStorageRemove("bricspay_user");
      authStorageRemove("bricspay_role");

      throw new Error(
        "Could not save the authentication token in this browser."
      );
    }

    window.location.reload();
  },

  async loginWithMetaMask() {
    const walletButton = document.getElementById("metaMaskBtn");

    if (walletButton) {
      walletButton.disabled = true;
    }

    try {
      if (!window.ethereum) {
        throw new Error(
          "MetaMask was not detected. Please install or enable the MetaMask wallet."
        );
      }

      throw new Error(
        "MetaMask sign-in is not connected to a verified server login flow. Please sign in with email and password."
      );
    } catch (error) {
      this.showStatus(
        error.message || "MetaMask sign-in failed.",
        true
      );
    } finally {
      if (walletButton) {
        walletButton.disabled = false;
      }
    }
  },

  logout() {
    const app = window.App;

    if (app && typeof app.clearAuth === "function") {
      app.clearAuth();
    } else {
      authStorageRemove("bricspay_token");
      authStorageRemove("bricspay_user");
      authStorageRemove("bricspay_role");
    }

    this.closeModal();
    window.location.reload();
  }
};

if (typeof window !== "undefined") {
  window.AuthModule = AuthModule;
}
