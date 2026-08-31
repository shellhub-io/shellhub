import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { client } from "@/client/client.gen";
import { useAuthStore } from "@/stores/authStore";
import { useConnectivityStore } from "@/stores/connectivityStore";
import { isSdkError } from "@/api/errors";
import "@/api/fetchInterceptors";

const GRACE_PERIOD_MS = 5000;

function makeJwt(exp: number): string {
  const header = btoa(JSON.stringify({ alg: "HS256" }));
  const payload = btoa(JSON.stringify({ exp }));
  return `${header}.${payload}.sig`;
}

function futureExp() {
  return Math.floor(Date.now() / 1000) + 3600;
}

function pastExp() {
  return Math.floor(Date.now() / 1000) - 60;
}

function respondWith(status: number, headers: Record<string, string> = {}) {
  const fetchMock = vi.fn().mockImplementation((request: Request) => {
    const response = new Response(JSON.stringify({}), {
      status,
      headers: { "Content-Type": "application/json", ...headers },
    });
    Object.defineProperty(response, "url", { value: request.url });
    return Promise.resolve(response);
  });
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function failToConnect() {
  const fetchMock = vi.fn().mockRejectedValue(new TypeError("Failed to fetch"));
  vi.stubGlobal("fetch", fetchMock);
  return fetchMock;
}

function setLocation(search = "") {
  Object.defineProperty(window, "location", {
    writable: true,
    value: { href: "", search, replace: vi.fn() },
  });
}

beforeEach(() => {
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
    mfaToken: null,
  });

  useConnectivityStore.getState().markUp();
  setLocation();
});

afterEach(() => {
  if (vi.isFakeTimers()) vi.advanceTimersByTime(GRACE_PERIOD_MS);
  vi.useRealTimers();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

describe("request interceptor", () => {
  it("attaches the bearer token when the token is valid", async () => {
    const token = makeJwt(futureExp());
    useAuthStore.setState({ token });
    const fetchMock = respondWith(200);

    await client.get({ url: "/test" });

    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.headers.get("Authorization")).toBe(`Bearer ${token}`);
  });

  it("sends no Authorization header when there is no token", async () => {
    const fetchMock = respondWith(200);

    await client.get({ url: "/test" });

    const request = fetchMock.mock.calls[0][0] as Request;
    expect(request.headers.get("Authorization")).toBeNull();
  });

  it.each([
    ["expired", () => makeJwt(pastExp())],
    ["malformed", () => "not-a-jwt"],
  ])("rejects a %s token before it reaches the network, and logs out", async (_label, makeToken) => {
    useAuthStore.setState({ token: makeToken() });
    const fetchMock = respondWith(200);

    await expect(client.get({ url: "/test" })).rejects.toThrow("Token expired");

    expect(fetchMock).not.toHaveBeenCalled();
    expect(useAuthStore.getState().token).toBeNull();
    expect(window.location.href).toBe("/login");
  });

  it("keeps the session on an expired token during a token login", async () => {
    setLocation("?token=abc");
    useAuthStore.setState({ token: makeJwt(pastExp()) });
    respondWith(200);

    await expect(client.get({ url: "/test" })).rejects.toThrow("Token expired");

    expect(useAuthStore.getState().token).not.toBeNull();
    expect(window.location.href).toBe("");
  });
});

describe("response interceptor", () => {
  it("logs out and redirects on 401 from a non-login route", async () => {
    useAuthStore.setState({ token: makeJwt(futureExp()) });
    respondWith(401);

    await client.get({ url: "/test" });

    expect(useAuthStore.getState().token).toBeNull();
    expect(window.location.href).toBe("/login");
  });

  it("keeps the session on 401 from the login route", async () => {
    useAuthStore.setState({ token: makeJwt(futureExp()) });
    respondWith(401);

    await client.post({ url: "/api/login" });

    expect(useAuthStore.getState().token).not.toBeNull();
    expect(window.location.href).not.toBe("/login");
  });

  it("marks the API up again on a successful response", async () => {
    useConnectivityStore.getState().markDown();
    respondWith(200);

    await client.get({ url: "/test" });

    expect(useConnectivityStore.getState().apiReachable).toBe(true);
  });

  it("stores the MFA token a 401 carries instead of logging out", async () => {
    useAuthStore.setState({ token: makeJwt(futureExp()) });
    respondWith(401, { "x-mfa-token": "mfa-temp-token-456" });

    await client.get({ url: "/test" });

    expect(useAuthStore.getState().mfaToken).toBe("mfa-temp-token-456");
    expect(useAuthStore.getState().token).not.toBeNull();
  });

  it("ignores an MFA token on a status other than 401", async () => {
    const token = makeJwt(futureExp());
    useAuthStore.setState({ token });
    respondWith(403, { "x-mfa-token": "should-be-ignored" });

    await client.get({ url: "/test" });

    expect(useAuthStore.getState().mfaToken).toBeNull();
    expect(useAuthStore.getState().token).toBe(token);
  });
});

describe("connectivity tracking", () => {
  it("marks the API down after the grace period when the request cannot connect", async () => {
    vi.useFakeTimers();
    failToConnect();

    await client.get({ url: "/test" });

    expect(useConnectivityStore.getState().apiReachable).toBe(true);

    vi.advanceTimersByTime(GRACE_PERIOD_MS);
    expect(useConnectivityStore.getState().apiReachable).toBe(false);
  });

  it.each([502, 503, 504])("marks the API down after the grace period on %i", async (status) => {
    vi.useFakeTimers();
    respondWith(status);

    await client.get({ url: "/test" });

    vi.advanceTimersByTime(GRACE_PERIOD_MS);
    expect(useConnectivityStore.getState().apiReachable).toBe(false);
  });

  it.each([400, 404, 422])("leaves the API marked up on %i", async (status) => {
    vi.useFakeTimers();
    respondWith(status);

    await client.get({ url: "/test" });

    vi.advanceTimersByTime(GRACE_PERIOD_MS);
    expect(useConnectivityStore.getState().apiReachable).toBe(true);
  });

  it("cancels the pending mark-down when a success arrives inside the grace period", async () => {
    vi.useFakeTimers();
    failToConnect();
    await client.get({ url: "/test" });

    vi.advanceTimersByTime(2000);
    expect(useConnectivityStore.getState().apiReachable).toBe(true);

    respondWith(200);
    await client.get({ url: "/test" });

    vi.advanceTimersByTime(GRACE_PERIOD_MS);
    expect(useConnectivityStore.getState().apiReachable).toBe(true);
  });
});

describe("error interceptor", () => {
  it("attaches the status and headers that isSdkError reads", async () => {
    respondWith(409, { "x-account-lockout": "60" });

    const { error } = await client.get({ url: "/test" });

    expect(isSdkError(error)).toBe(true);
    expect((error as { status: number }).status).toBe(409);
    expect((error as { headers: Headers }).headers.get("x-account-lockout")).toBe("60");
  });
});
