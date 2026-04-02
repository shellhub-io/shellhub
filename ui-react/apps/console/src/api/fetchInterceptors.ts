import { client } from "../client/client.gen";
import { useAuthStore } from "../stores/authStore";
import { useConnectivityStore } from "../stores/connectivityStore";

function isTokenExpired(token: string): boolean {
  try {
    const payload: unknown = JSON.parse(atob(token.split(".")[1]));
    if (typeof payload === "object" && payload !== null && "exp" in payload) {
      const { exp } = payload as { exp: unknown };
      return typeof exp === "number" && exp * 1000 < Date.now();
    }
    return false;
  } catch {
    return true;
  }
}

const GRACE_PERIOD_MS = 5000;
let downTimer: ReturnType<typeof setTimeout> | null = null;

function scheduleMarkDown() {
  if (downTimer) return;
  downTimer = setTimeout(() => {
    downTimer = null;
    useConnectivityStore.getState().markDown();
  }, GRACE_PERIOD_MS);
}

function cancelMarkDown() {
  if (downTimer) {
    clearTimeout(downTimer);
    downTimer = null;
  }
}

function isApiDown(status: number): boolean {
  return status === 502 || status === 503 || status === 504;
}

client.interceptors.request.use((request) => {
  const token = useAuthStore.getState().token;
  if (token) {
    if (isTokenExpired(token)) {
      const isTokenLogin = new URLSearchParams(window.location.search).has(
        "token",
      );
      if (!isTokenLogin) {
        useAuthStore.getState().logout();
        window.location.href = "/login";
      }
      throw new Error("Token expired");
    }
    request.headers.set("Authorization", `Bearer ${token}`);
  }
  return request;
});

client.interceptors.response.use((response) => {
  if (!isApiDown(response.status)) {
    cancelMarkDown();
    if (!useConnectivityStore.getState().apiReachable) {
      useConnectivityStore.getState().markUp();
    }
  }

  if (response.status === 401) {
    const mfaToken = response.headers.get("x-mfa-token");
    if (mfaToken) {
      useAuthStore.getState().setMfaToken(mfaToken);
    } else {
      const isLoginRequest = response.url.includes("/api/login");
      const isTokenLogin = new URLSearchParams(window.location.search).has(
        "token",
      );
      if (!isLoginRequest && !isTokenLogin) {
        useAuthStore.getState().logout();
        window.location.href = "/login";
      }
    }
  } else if (isApiDown(response.status)) {
    scheduleMarkDown();
  }

  return response;
});

client.interceptors.error.use((error, response) => {
  if (!response) {
    scheduleMarkDown();
    return error;
  }

  // Attach HTTP status and headers to the error so mutation handlers can
  // distinguish error codes (e.g., 402 vs 403 vs 409) and read headers
  // (e.g., x-account-lockout on 429).
  const enriched = typeof error === "object" && error !== null ? error : {};
  (enriched as Record<string, unknown>).status = response.status;
  (enriched as Record<string, unknown>).headers = response.headers;
  return enriched;
});
