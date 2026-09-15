import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "@/tests/msw";
import { useAuthStore } from "@/stores/authStore";
import { useConnectivityStore } from "@/stores/connectivityStore";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { customInstance, fetchWithHeaders } from "../customInstance";
import { totalCount } from "../pagination";

const EXPIRED_JWT = "eyJhbGciOiJIUzI1NiJ9.eyJleHAiOjF9.fake-sig";
const GRACE_MS = 5000;

let hrefSpy: (v: string) => void;

beforeEach(() => {
  vi.useFakeTimers();
  seedAuthStore();
  useConnectivityStore.setState({ apiReachable: true });
  hrefSpy = vi.fn();
  Object.defineProperty(window, "location", {
    value: {
      ...window.location,
      origin: "http://localhost",
      href: "",
      search: "",
    },
    writable: true,
    configurable: true,
  });
  Object.defineProperty(window.location, "href", {
    set: hrefSpy,
    get: () => "http://localhost/",
    configurable: true,
  });
});

afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
});

describe("customInstance", () => {
  describe("request building", () => {
    it("attaches the bearer token when authenticated", async () => {
      let captured: string | null = null;
      server.use(
        http.get("*/api/test", ({ request }) => {
          captured = request.headers.get("Authorization");
          return HttpResponse.json({});
        }),
      );

      await customInstance("/api/test", { method: "GET" });
      expect(captured).toBe(`Bearer ${VALID_JWT}`);
    });

    it("skips auth header when no token is stored", async () => {
      useAuthStore.setState({ token: null });
      let captured: string | null = null;
      server.use(
        http.get("*/api/test", ({ request }) => {
          captured = request.headers.get("Authorization");
          return HttpResponse.json({});
        }),
      );

      await customInstance("/api/test", { method: "GET" });
      expect(captured).toBeNull();
    });

    it("appends query params to the URL", async () => {
      let capturedUrl = "";
      server.use(
        http.get("*/api/devices", ({ request }) => {
          capturedUrl = request.url;
          return HttpResponse.json([]);
        }),
      );

      await customInstance("/api/devices", {
        method: "GET",
        params: { status: "accepted", page: "2" },
      });
      const url = new URL(capturedUrl);
      expect(url.searchParams.get("status")).toBe("accepted");
      expect(url.searchParams.get("page")).toBe("2");
    });

    it("sets Content-Type to application/json for string bodies", async () => {
      let captured: string | null = null;
      server.use(
        http.post("*/api/users", ({ request }) => {
          captured = request.headers.get("Content-Type");
          return HttpResponse.json({}, { status: 201 });
        }),
      );

      await customInstance("/api/users", {
        method: "POST",
        body: JSON.stringify({ name: "alice" }),
      });
      expect(captured).toBe("application/json");
    });
  });

  describe("token expiry", () => {
    it("throws and logs out when the token is expired", async () => {
      useAuthStore.setState({ token: EXPIRED_JWT });
      const logoutSpy = vi.spyOn(useAuthStore.getState(), "logout");

      await expect(
        customInstance("/api/test", { method: "GET" }),
      ).rejects.toThrow("Token expired");
      expect(logoutSpy).toHaveBeenCalled();
      expect(hrefSpy).toHaveBeenCalledWith("/login");
    });

    it("throws but does NOT redirect when on a token-login page", async () => {
      useAuthStore.setState({ token: EXPIRED_JWT });
      Object.defineProperty(window.location, "search", {
        get: () => "?token=abc",
        configurable: true,
      });

      await expect(
        customInstance("/api/test", { method: "GET" }),
      ).rejects.toThrow("Token expired");
      expect(hrefSpy).not.toHaveBeenCalled();
    });

    it("treats a malformed JWT as expired", async () => {
      useAuthStore.setState({ token: "not-a-jwt" });

      await expect(
        customInstance("/api/test", { method: "GET" }),
      ).rejects.toThrow("Token expired");
    });
  });

  describe("401 handling", () => {
    it("stores the MFA token from the response header", async () => {
      server.use(
        http.post(
          "*/api/login",
          () =>
            new HttpResponse(JSON.stringify({ message: "mfa required" }), {
              status: 401,
              headers: { "x-mfa-token": "mfa-temp-123" },
            }),
        ),
      );

      await expect(
        customInstance("/api/login", { method: "POST" }),
      ).rejects.toThrow();
      expect(useAuthStore.getState().mfaToken).toBe("mfa-temp-123");
    });

    it("logs out on 401 for a non-login request without MFA header", async () => {
      server.use(
        http.get(
          "*/api/devices",
          () => new HttpResponse(JSON.stringify({}), { status: 401 }),
        ),
      );
      const logoutSpy = vi.spyOn(useAuthStore.getState(), "logout");

      await expect(
        customInstance("/api/devices", { method: "GET" }),
      ).rejects.toThrow();
      expect(logoutSpy).toHaveBeenCalled();
      expect(hrefSpy).toHaveBeenCalledWith("/login");
    });

    it("does NOT log out on 401 for a login request", async () => {
      server.use(
        http.post(
          "*/api/login",
          () =>
            new HttpResponse(JSON.stringify({ message: "bad creds" }), {
              status: 401,
            }),
        ),
      );
      const logout = vi.fn();
      useAuthStore.setState({ logout });

      await expect(
        customInstance("/api/login", { method: "POST" }),
      ).rejects.toThrow();
      expect(logout).not.toHaveBeenCalled();
    });

    it("does NOT log out on 401 when on a token-login page", async () => {
      Object.defineProperty(window.location, "search", {
        get: () => "?token=abc",
        configurable: true,
      });
      server.use(
        http.get(
          "*/api/devices",
          () => new HttpResponse(JSON.stringify({}), { status: 401 }),
        ),
      );
      const logout = vi.fn();
      useAuthStore.setState({ logout });

      await expect(
        customInstance("/api/devices", { method: "GET" }),
      ).rejects.toThrow();
      expect(logout).not.toHaveBeenCalled();
      expect(hrefSpy).not.toHaveBeenCalled();
    });
  });

  describe("connectivity tracking", () => {
    it("schedules markDown on network error", async () => {
      server.use(http.get("*/api/test", () => HttpResponse.error()));

      await expect(
        customInstance("/api/test", { method: "GET" }),
      ).rejects.toThrow();

      vi.advanceTimersByTime(GRACE_MS);
      expect(useConnectivityStore.getState().apiReachable).toBe(false);
    });

    it.each([502, 503, 504])("schedules markDown on %i", async (status) => {
      server.use(
        http.get("*/api/test", () => new HttpResponse(null, { status })),
      );

      await expect(
        customInstance("/api/test", { method: "GET" }),
      ).rejects.toThrow();

      vi.advanceTimersByTime(GRACE_MS);
      expect(useConnectivityStore.getState().apiReachable).toBe(false);
    });

    it("cancels markDown and marks up on a successful response", async () => {
      useConnectivityStore.setState({ apiReachable: false });
      server.use(http.get("*/api/test", () => HttpResponse.json({})));

      await customInstance("/api/test", { method: "GET" });
      expect(useConnectivityStore.getState().apiReachable).toBe(true);
    });
  });

  describe("error enrichment", () => {
    it("throws with status and body fields from the API response", async () => {
      server.use(
        http.post("*/api/users", () =>
          HttpResponse.json(
            { message: "conflict", fields: { username: "taken" } },
            { status: 409 },
          ),
        ),
      );

      const err: unknown = await customInstance("/api/users", {
        method: "POST",
      }).catch((e: unknown) => e);
      expect(err).toHaveProperty("status", 409);
      expect(err).toHaveProperty("fields", { username: "taken" });
    });

    it("falls back to empty object when body is not JSON", async () => {
      server.use(
        http.get(
          "*/api/test",
          () => new HttpResponse("Internal Server Error", { status: 500 }),
        ),
      );

      const err: unknown = await customInstance("/api/test", {
        method: "GET",
      }).catch((e: unknown) => e);
      expect(err).toHaveProperty("status", 500);
    });
  });

  describe("response parsing", () => {
    it("returns undefined for 204", async () => {
      server.use(
        http.delete(
          "*/api/devices/uid-1",
          () => new HttpResponse(null, { status: 204 }),
        ),
      );

      const result = await customInstance("/api/devices/uid-1", {
        method: "DELETE",
      });
      expect(result).toBeUndefined();
    });

    it("attaches totalCount from X-Total-Count on array responses", async () => {
      server.use(
        http.get("*/api/devices", () =>
          HttpResponse.json([{ uid: "a" }, { uid: "b" }], {
            headers: { "X-Total-Count": "50" },
          }),
        ),
      );

      const data = await customInstance<unknown[]>("/api/devices", {
        method: "GET",
      });
      expect(data).toHaveLength(2);
      expect(totalCount(data)).toBe(50);
    });

    it("returns text when content-type is not JSON", async () => {
      server.use(
        http.get(
          "*/api/test",
          () =>
            new HttpResponse("plain text", {
              headers: { "Content-Type": "text/plain" },
            }),
        ),
      );

      const result = await customInstance<string>("/api/test", {
        method: "GET",
      });
      expect(result).toBe("plain text");
    });
  });
});

describe("fetchWithHeaders", () => {
  it("returns data and response headers together", async () => {
    server.use(
      http.get("*/api/devices", () =>
        HttpResponse.json([{ uid: "a" }], {
          headers: { "X-Total-Count": "10" },
        }),
      ),
    );

    const { data, headers } = await fetchWithHeaders<unknown[]>(
      "/api/devices",
      {
        method: "GET",
      },
    );
    expect(data).toHaveLength(1);
    expect(headers.get("X-Total-Count")).toBe("10");
  });

  it("returns undefined data for 204", async () => {
    server.use(
      http.delete("*/api/test", () => new HttpResponse(null, { status: 204 })),
    );

    const { data } = await fetchWithHeaders("/api/test", { method: "DELETE" });
    expect(data).toBeUndefined();
  });
});
