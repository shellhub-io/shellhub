import type { SdkHttpError } from "./errors";
import { useAuthStore } from "@/stores/authStore";
import { useConnectivityStore } from "@/stores/connectivityStore";

/** Error shape threaded through every generated hook. */
export type ErrorType<_Error> = SdkHttpError;

/** Request body passthrough — Orval requires this export. */
export type BodyType<BodyData> = BodyData;

const baseURL = window.location.origin;

function isTokenExpired(token: string): boolean {
  try {
    const payload: unknown = JSON.parse(atob(token.split(".")[1]));
    if (typeof payload === "object" && payload !== null && "exp" in payload) {
      const { exp } = payload;
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

interface MutatorOptions {
  method: "GET" | "POST" | "PUT" | "DELETE" | "PATCH";
  params?: Record<string, string>;
  body?: BodyType<unknown>;
  headers?: Record<string, string>;
  signal?: AbortSignal;
}

async function doFetch(
  url: string,
  { method, params, body, headers, signal }: MutatorOptions,
): Promise<Response> {
  let targetUrl = `${baseURL}${url}`;

  if (params) {
    const search = new URLSearchParams(params).toString();
    if (search) targetUrl += `?${search}`;
  }

  const requestHeaders: Record<string, string> = { ...headers };

  if (body !== undefined) {
    requestHeaders["Content-Type"] ??= "application/json";
  }

  const token = useAuthStore.getState().token;
  const isTokenLogin = new URLSearchParams(window.location.search).has("token");
  if (token) {
    if (isTokenExpired(token)) {
      if (!isTokenLogin) {
        useAuthStore.getState().logout();
        window.location.href = "/login";
      }
      throw new Error("Token expired");
    }
    requestHeaders["Authorization"] = `Bearer ${token}`;
  }

  let response: Response;
  try {
    response = await fetch(targetUrl, {
      method,
      headers: requestHeaders,
      body: body !== undefined ? JSON.stringify(body) : undefined,
      signal,
    });
  } catch (error) {
    scheduleMarkDown();
    throw error;
  }

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
      if (!isLoginRequest && !isTokenLogin) {
        useAuthStore.getState().logout();
        window.location.href = "/login";
      }
    }
  } else if (isApiDown(response.status)) {
    scheduleMarkDown();
  }

  if (!response.ok) {
    const errorBody: unknown = await response.json().catch(() => ({}));
    const fields =
      typeof errorBody === "object" && errorBody !== null
        ? (errorBody as Record<string, unknown>)
        : {};
    const error = Object.assign(new Error(String(response.status)), {
      ...fields,
      status: response.status,
      headers: response.headers,
    });
    throw error;
  }

  return response;
}

/** Orval custom mutator — every generated SDK function calls this. */
export const customInstance = async <T>(
  url: string,
  options: MutatorOptions,
): Promise<T> => {
  const response = await doFetch(url, options);
  if (response.status === 204) return undefined as T;
  const ct = response.headers.get("content-type") ?? "";
  if (ct.includes("application/json")) return (await response.json()) as T;
  return (await response.text()) as unknown as T;
};

export default customInstance;

/** Like `customInstance` but also exposes the response headers (for `X-Total-Count`). */
export async function fetchWithHeaders<T>(
  url: string,
  options: MutatorOptions,
): Promise<{ data: T; headers: Headers }> {
  const response = await doFetch(url, options);
  if (response.status === 204)
    return { data: undefined as T, headers: response.headers };
  const data = (await response.json()) as T;
  return { data, headers: response.headers };
}
