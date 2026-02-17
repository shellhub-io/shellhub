import { AxiosInstance, InternalAxiosRequestConfig, AxiosError } from "axios";
import { useAuthStore } from "../stores/authStore";
import { useConnectivityStore } from "../stores/connectivityStore";

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split(".")[1]));
    return typeof payload.exp === "number" && payload.exp * 1000 < Date.now();
  } catch {
    return true;
  }
}

function isApiDown(error: AxiosError): boolean {
  // Network error — no response at all (ECONNREFUSED, ERR_NETWORK, timeout)
  if (!error.response) return true;

  // Gateway errors — API process is down but gateway is up
  const status = error.response.status;
  return status === 502 || status === 503 || status === 504;
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

export function setupInterceptors(instance: AxiosInstance) {
  instance.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token) {
      if (isTokenExpired(token)) {
        useAuthStore.getState().logout();
        window.location.href = "/v2/ui/login";
        return Promise.reject(new Error("Token expired"));
      }
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  });

  instance.interceptors.response.use(
    (response) => {
      cancelMarkDown();
      if (!useConnectivityStore.getState().apiReachable) {
        useConnectivityStore.getState().markUp();
      }
      return response;
    },
    (error: AxiosError) => {
      if (error.response?.status === 401) {
        useAuthStore.getState().logout();
        window.location.href = "/v2/ui/login";
      } else if (isApiDown(error)) {
        scheduleMarkDown();
      }
      return Promise.reject(error);
    },
  );
}
