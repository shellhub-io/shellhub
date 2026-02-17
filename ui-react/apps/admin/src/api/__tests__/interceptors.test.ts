import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import axios from "axios";
import { setupInterceptors } from "../interceptors";
import { useAuthStore } from "../../stores/authStore";
import { useConnectivityStore } from "../../stores/connectivityStore";

/* --- helpers --- */

/** Build a minimal JWT with the given expiry (unix seconds). */
function makeJwt(exp: number): string {
  const header = btoa(JSON.stringify({ alg: "HS256" }));
  const payload = btoa(JSON.stringify({ exp }));
  return `${header}.${payload}.sig`;
}

function futureExp() {
  return Math.floor(Date.now() / 1000) + 3600; // +1h
}

function pastExp() {
  return Math.floor(Date.now() / 1000) - 60; // -1m
}

/* --- setup --- */

let client: ReturnType<typeof axios.create>;

beforeEach(() => {
  client = axios.create({ baseURL: "http://localhost" });
  setupInterceptors(client);

  // Reset stores
  useAuthStore.setState({
    token: null,
    user: null,
    userId: null,
    email: null,
    tenant: null,
    role: null,
    name: null,
    loading: false,
    error: null,
  });

  useConnectivityStore.getState().markUp();

  // Stub window.location
  Object.defineProperty(window, "location", {
    writable: true,
    value: { href: "", replace: vi.fn() },
  });
});

afterEach(() => {
  vi.restoreAllMocks();
  vi.useRealTimers();
});

/* ================================================================
   Request interceptor
   ================================================================ */

describe("request interceptor", () => {
  it("attaches Bearer token when token exists and is valid", async () => {
    const token = makeJwt(futureExp());
    useAuthStore.setState({ token });

    // Intercept the actual network call to inspect the config
    const adapter = vi
      .fn()
      .mockResolvedValue({ data: {}, status: 200, headers: {}, config: {} });
    client.defaults.adapter = adapter;

    await client.get("/test");

    expect(adapter).toHaveBeenCalledTimes(1);
    const config = adapter.mock.calls[0][0];
    expect(config.headers.get("Authorization")).toBe(`Bearer ${token}`);
  });

  it("does not attach Authorization header when no token", async () => {
    const adapter = vi
      .fn()
      .mockResolvedValue({ data: {}, status: 200, headers: {}, config: {} });
    client.defaults.adapter = adapter;

    await client.get("/test");

    const config = adapter.mock.calls[0][0];
    expect(config.headers.get("Authorization")).toBeFalsy();
  });

  it("rejects and redirects to login when token is expired", async () => {
    const token = makeJwt(pastExp());
    useAuthStore.setState({ token });

    const adapter = vi.fn();
    client.defaults.adapter = adapter;

    await expect(client.get("/test")).rejects.toThrow("Token expired");

    // Should not reach the network
    expect(adapter).not.toHaveBeenCalled();

    // Should logout
    expect(useAuthStore.getState().token).toBeNull();

    // Should redirect
    expect(window.location.href).toBe("/v2/ui/login");
  });

  it("rejects when token is malformed", async () => {
    useAuthStore.setState({ token: "not-a-jwt" });

    const adapter = vi.fn();
    client.defaults.adapter = adapter;

    await expect(client.get("/test")).rejects.toThrow("Token expired");
    expect(adapter).not.toHaveBeenCalled();
  });
});

/* ================================================================
   Response interceptor
   ================================================================ */

describe("response interceptor", () => {
  it("logs out and redirects on 401", async () => {
    useAuthStore.setState({ token: makeJwt(futureExp()) });

    const adapter = vi.fn().mockRejectedValue({
      response: { status: 401 },
      isAxiosError: true,
    });
    client.defaults.adapter = adapter;

    await expect(client.get("/test")).rejects.toBeDefined();

    expect(useAuthStore.getState().token).toBeNull();
    expect(window.location.href).toBe("/v2/ui/login");
  });

  it("marks API as up on successful response", async () => {
    useConnectivityStore.getState().markDown();
    expect(useConnectivityStore.getState().apiReachable).toBe(false);

    const adapter = vi
      .fn()
      .mockResolvedValue({ data: {}, status: 200, headers: {}, config: {} });
    client.defaults.adapter = adapter;

    await client.get("/test");

    expect(useConnectivityStore.getState().apiReachable).toBe(true);
  });

  it("schedules markDown after grace period on network error", async () => {
    vi.useFakeTimers();

    const adapter = vi.fn().mockRejectedValue({
      response: undefined, // network error
      isAxiosError: true,
    });
    client.defaults.adapter = adapter;

    await expect(client.get("/test")).rejects.toBeDefined();

    // Not yet marked down (grace period)
    expect(useConnectivityStore.getState().apiReachable).toBe(true);

    // After grace period
    vi.advanceTimersByTime(5000);
    expect(useConnectivityStore.getState().apiReachable).toBe(false);
  });

  it("schedules markDown on 502/503/504", async () => {
    vi.useFakeTimers();

    for (const status of [502, 503, 504]) {
      useConnectivityStore.getState().markUp();

      const adapter = vi.fn().mockRejectedValue({
        response: { status },
        isAxiosError: true,
      });
      client.defaults.adapter = adapter;

      await expect(client.get("/test")).rejects.toBeDefined();

      vi.advanceTimersByTime(5000);
      expect(useConnectivityStore.getState().apiReachable).toBe(false);
    }
  });

  it("cancels markDown if a successful response arrives during grace period", async () => {
    vi.useFakeTimers();

    // First: network error
    const failAdapter = vi.fn().mockRejectedValue({
      response: undefined,
      isAxiosError: true,
    });
    client.defaults.adapter = failAdapter;
    await expect(client.get("/test")).rejects.toBeDefined();

    // Advance partially through grace period
    vi.advanceTimersByTime(2000);
    expect(useConnectivityStore.getState().apiReachable).toBe(true);

    // Then: successful response
    const okAdapter = vi
      .fn()
      .mockResolvedValue({ data: {}, status: 200, headers: {}, config: {} });
    client.defaults.adapter = okAdapter;
    await client.get("/test");

    // Finish grace period — should NOT be marked down
    vi.advanceTimersByTime(5000);
    expect(useConnectivityStore.getState().apiReachable).toBe(true);
  });

  it("does not mark down on regular client errors (400, 404, 422)", async () => {
    vi.useFakeTimers();

    for (const status of [400, 404, 422]) {
      const adapter = vi.fn().mockRejectedValue({
        response: { status },
        isAxiosError: true,
      });
      client.defaults.adapter = adapter;

      await expect(client.get("/test")).rejects.toBeDefined();

      vi.advanceTimersByTime(5000);
      expect(useConnectivityStore.getState().apiReachable).toBe(true);
    }
  });
});
