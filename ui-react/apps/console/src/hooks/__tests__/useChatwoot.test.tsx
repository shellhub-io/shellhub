import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";

// ─── Module mocks (must come before any import of the module under test) ──────

const mockGetConfig = vi.fn();

vi.mock("@/env", () => ({
  getConfig: (): unknown => mockGetConfig(),
}));

const mockUseAuthStore = vi.fn();

vi.mock("@/stores/authStore", () => ({
  useAuthStore: (selector: (s: unknown) => unknown): unknown =>
    mockUseAuthStore(selector),
}));

const mockUseNamespace = vi.fn();

vi.mock("@/hooks/useNamespaces", () => ({
  useNamespace: (...args: unknown[]): unknown => mockUseNamespace(...args),
}));

const mockUseSupportIdentifier = vi.fn();

vi.mock("@/hooks/useSupportIdentifier", () => ({
  useSupportIdentifier: (...args: unknown[]): unknown =>
    mockUseSupportIdentifier(...args),
}));

// ─── Helpers ──────────────────────────────────────────────────────────────────

function makeConfig(
  overrides: Partial<{
    cloud: boolean;
    chatwootWebsiteToken: string;
    chatwootBaseUrl: string;
  }> = {},
) {
  return {
    cloud: false,
    chatwootWebsiteToken: "",
    chatwootBaseUrl: "",
    ...overrides,
  };
}

/** Simulate the Zustand selector pattern for authStore */
function setupAuthStore(values: {
  userId?: string | null;
  email?: string | null;
  name?: string | null;
  tenant?: string | null;
}) {
  const store = {
    userId: values.userId ?? "user-1",
    email: values.email ?? "user@example.com",
    name: values.name ?? "Test User",
    tenant: values.tenant ?? "tenant-abc",
  };
  mockUseAuthStore.mockImplementation(
    (selector: (s: typeof store) => unknown) => selector(store),
  );
}

function setupNamespace(
  ns: {
    name?: string;
    billing?: Record<string, unknown>;
  } | null,
) {
  mockUseNamespace.mockReturnValue({ namespace: ns, isLoading: !ns });
}

function setupIdentifier(identifier: string | null, isLoading = false) {
  mockUseSupportIdentifier.mockReturnValue({
    identifier,
    isLoading,
    isError: false,
  });
}

function setupIdentifierError() {
  mockUseSupportIdentifier.mockReturnValue({
    identifier: null,
    isLoading: false,
    isError: true,
  });
}

function makeWidgetReady() {
  window.$chatwoot = {
    setUser: vi.fn(),
    toggle: vi.fn(),
    reset: vi.fn(),
    setConversationCustomAttributes: vi.fn(),
    setCustomAttributes: vi.fn(),
    deleteCustomAttribute: vi.fn(),
  };
  window.dispatchEvent(new Event("chatwoot:ready"));
}

async function importHook() {
  return await import("../useChatwoot");
}

// ─── Lifecycle ────────────────────────────────────────────────────────────────

beforeEach(async () => {
  vi.clearAllMocks();
  // Reset DOM + window globals between tests
  document.body.innerHTML = "";
  delete window.$chatwoot;
  delete window.chatwootSDK;
  delete window.chatwootSettings;
  // Reset module-level chatwoot runtime state (script latch, watchdog timer,
  // bootstrapFailed flag) so each test starts from a clean slate.
  const runtime = await import("../chatwootRuntime");
  runtime.tearDownChatwoot("logout");
});

afterEach(() => {
  vi.resetModules();
});

// ─── Tests ────────────────────────────────────────────────────────────────────

describe("useChatwoot", () => {
  describe("status: non-cloud", () => {
    it("returns status 'non-cloud' when cloud=false", async () => {
      mockGetConfig.mockReturnValue(makeConfig({ cloud: false }));
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("non-cloud");
    });

    it("does not inject a script when cloud=false", async () => {
      mockGetConfig.mockReturnValue(makeConfig({ cloud: false }));
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("status: unavailable", () => {
    it("returns 'unavailable' when cloud=true but chatwootWebsiteToken is missing", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("unavailable");
    });

    it("returns 'unavailable' when cloud=true but chatwootBaseUrl is missing", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("unavailable");
    });
  });

  describe("status: loading (namespace not resolved)", () => {
    it("returns 'loading' when namespace is null", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace(null);
      setupIdentifier(null, false);

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("loading");
    });
  });

  describe("status: no-subscription", () => {
    it("returns 'no-subscription' when namespace loaded but billing.active !== true", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: false } });
      setupIdentifier(null, false);

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("no-subscription");
    });

    it("returns 'no-subscription' when namespace has no billing object", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns" });
      setupIdentifier(null, false);

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("no-subscription");
    });
  });

  describe("status: loading (identifier fetching)", () => {
    it("returns 'loading' while support identifier is being fetched", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier(null, true); // isLoading=true

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("loading");
    });

    it("returns 'loading' when identifier has not arrived yet (isLoading=false but identifier=null)", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier(null, false);

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("loading");
    });
  });

  describe("status: ready", () => {
    it("returns 'ready' after the chatwoot:ready event fires with window.$chatwoot set", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() => expect(result.current.status).toBe("ready"));
    });
  });

  describe("script injection", () => {
    it("injects a script with the correct src when prerequisites are met", async () => {
      const baseURL = "https://chat.example.com";
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: baseURL,
        }),
      );
      setupAuthStore({ userId: "user-1" });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      await waitFor(() => {
        const script = document.getElementById("shellhub-chatwoot-sdk");
        expect(script).not.toBeNull();
        expect(script!.getAttribute("src")).toBe(`${baseURL}/packs/js/sdk.js`);
      });
    });

    it("does not inject a second script when the hook renders twice (StrictMode safety)", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({ userId: "user-1" });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { rerender } = renderHook(() => useChatwoot());
      rerender();

      await waitFor(() => {
        const scripts = document.querySelectorAll("#shellhub-chatwoot-sdk");
        expect(scripts.length).toBe(1);
      });
    });
  });

  describe("window.chatwootSettings", () => {
    it("sets chatwootSettings with the correct values before script injection", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({ userId: "user-1" });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      await waitFor(() =>
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull(),
      );

      expect(window.chatwootSettings).toEqual({
        locale: "en",
        position: "right",
        hideMessageBubble: true,
        type: "standard",
      });
    });
  });

  describe("setUser after widget ready", () => {
    it("calls setUser with identifier_hash when widget reports ready", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      const userId = "user-1";
      const email = "user@example.com";
      const name = "Test User";
      setupAuthStore({ userId, email, name, tenant: "tenant-abc" });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() =>
        expect(window.$chatwoot!.setUser).toHaveBeenCalledWith(userId, {
          email,
          name,
          identifier_hash: "abc123",
        }),
      );
    });

    it("does not call setUser again when re-rendered with the same identity", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({
        userId: "user-1",
        email: "user@example.com",
        name: "Test User",
        tenant: "tenant-abc",
      });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { rerender } = renderHook(() => useChatwoot());

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() =>
        expect(window.$chatwoot!.setUser).toHaveBeenCalledTimes(1),
      );

      rerender();

      // Still only called once
      expect(window.$chatwoot!.setUser).toHaveBeenCalledTimes(1);
    });
  });

  describe("setConversationCustomAttributes", () => {
    it("calls setConversationCustomAttributes with namespace data on chatwoot:on-message", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({ userId: "user-1", tenant: "tenant-abc" });
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() => expect(window.$chatwoot!.setUser).toHaveBeenCalled());

      act(() => {
        window.dispatchEvent(new Event("chatwoot:on-message"));
      });

      expect(
        window.$chatwoot!.setConversationCustomAttributes,
      ).toHaveBeenCalledWith({
        namespace: "my-ns",
        tenant: "tenant-abc",
        domain: window.location.hostname,
      });
    });
  });

  describe("openWidget", () => {
    it("calls window.$chatwoot.toggle('open') when status is ready", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() => expect(result.current.status).toBe("ready"));

      act(() => {
        result.current.openWidget();
      });

      expect(window.$chatwoot!.toggle).toHaveBeenCalledWith("open");
    });

    it("does not call toggle when status is not ready (loading)", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace(null); // loading state
      setupIdentifier(null, true);

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      // No $chatwoot set, so toggle would throw if called
      const toggleSpy = vi.fn();
      window.$chatwoot = {
        setUser: vi.fn(),
        toggle: toggleSpy,
        reset: vi.fn(),
        setConversationCustomAttributes: vi.fn(),
        setCustomAttributes: vi.fn(),
        deleteCustomAttribute: vi.fn(),
      };

      act(() => {
        result.current.openWidget();
      });

      expect(toggleSpy).not.toHaveBeenCalled();
    });
  });

  describe("status: unavailable (identifier endpoint errors)", () => {
    it("flips to 'unavailable' when /support returns an error (e.g. backend identity key missing)", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifierError();

      const { useChatwoot } = await importHook();
      const { result } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("unavailable");
      // No script injected — operator misconfiguration shouldn't load Chatwoot.
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("status: unavailable (bootstrap watchdog)", () => {
    it("flips to 'unavailable' when chatwoot:ready never fires after script load", async () => {
      vi.useFakeTimers();
      try {
        mockGetConfig.mockReturnValue(
          makeConfig({
            cloud: true,
            chatwootWebsiteToken: "token-abc",
            chatwootBaseUrl: "https://chat.example.com",
          }),
        );
        setupAuthStore({});
        setupNamespace({ name: "my-ns", billing: { active: true } });
        setupIdentifier("abc123");

        const { useChatwoot } = await importHook();
        const { result, rerender } = renderHook(() => useChatwoot());

        // Status starts at "loading"; script is injected.
        expect(result.current.status).toBe("loading");
        const script = document.getElementById(
          "shellhub-chatwoot-sdk",
        ) as HTMLScriptElement | null;
        expect(script).not.toBeNull();

        // Simulate script loaded but chatwootSDK.run never produces a ready event.
        // We need chatwootSDK present so .run() doesn't throw.
        window.chatwootSDK = { run: vi.fn() };
        script!.onload?.(new Event("load"));

        // Watchdog should fire after BOOTSTRAP_TIMEOUT_MS (15s).
        await act(async () => {
          await vi.advanceTimersByTimeAsync(15_001);
        });
        rerender();

        expect(result.current.status).toBe("unavailable");
        expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
      } finally {
        vi.useRealTimers();
      }
    });
  });

  describe("status: unavailable (script load error)", () => {
    it("flips to 'unavailable' when the SDK script fails to load (script.onerror)", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const { result, rerender } = renderHook(() => useChatwoot());

      expect(result.current.status).toBe("loading");
      const script = document.getElementById(
        "shellhub-chatwoot-sdk",
      ) as HTMLScriptElement | null;
      expect(script).not.toBeNull();

      // Network/CSP failure: <script> never loads.
      await act(async () => {
        script!.onerror?.(new Event("error"));
      });
      rerender();

      // Without the fix this stays "loading" forever (notify never fires).
      expect(result.current.status).toBe("unavailable");
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("teardown via runtime helper", () => {
    it("tearDownChatwoot('logout') removes the script tag and clears window globals", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());

      // Bootstrap a "ready" widget.
      makeWidgetReady();
      await waitFor(() => {
        expect(window.$chatwoot).toBeDefined();
      });

      const runtime = await import("../chatwootRuntime");
      runtime.tearDownChatwoot("logout");

      expect(window.$chatwoot).toBeUndefined();
      expect(window.chatwootSDK).toBeUndefined();
      expect(window.chatwootSettings).toBeUndefined();
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });

    it("after teardown, a re-mount cleanly re-injects the script", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      const first = renderHook(() => useChatwoot());
      expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      first.unmount();

      const runtime = await import("../chatwootRuntime");
      runtime.tearDownChatwoot("logout");
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();

      // Simulate re-login: same config, new mount — script injects again.
      renderHook(() => useChatwoot());
      expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
    });

    it("removes SDK-injected DOM (chat bubble holders, iframe) on teardown", async () => {
      mockGetConfig.mockReturnValue(
        makeConfig({
          cloud: true,
          chatwootWebsiteToken: "token-abc",
          chatwootBaseUrl: "https://chat.example.com",
        }),
      );
      setupAuthStore({});
      setupNamespace({ name: "my-ns", billing: { active: true } });
      setupIdentifier("abc123");

      const { useChatwoot } = await importHook();
      renderHook(() => useChatwoot());
      makeWidgetReady();

      // Simulate the DOM the Chatwoot SDK attaches outside the React tree.
      const bubble = document.createElement("div");
      bubble.className = "woot-widget-holder";
      const altBubble = document.createElement("div");
      altBubble.className = "woot--bubble-holder";
      const iframe = document.createElement("iframe");
      iframe.src = "https://chatwoot.example.com/widget?token=abc";
      document.body.append(bubble, altBubble, iframe);

      const runtime = await import("../chatwootRuntime");
      runtime.tearDownChatwoot("logout");

      expect(document.querySelector(".woot-widget-holder")).toBeNull();
      expect(document.querySelector(".woot--bubble-holder")).toBeNull();
      expect(document.querySelector('iframe[src*="chatwoot"]')).toBeNull();
    });
  });

  describe("bootstrapFailed recovery", () => {
    it("clears bootstrapFailed when a fresh injection proceeds (recovers from prior watchdog timeout)", async () => {
      vi.useFakeTimers();
      try {
        mockGetConfig.mockReturnValue(
          makeConfig({
            cloud: true,
            chatwootWebsiteToken: "token-abc",
            chatwootBaseUrl: "https://chat.example.com",
          }),
        );
        setupAuthStore({});
        setupNamespace({ name: "my-ns", billing: { active: true } });
        setupIdentifier("abc123");

        const { useChatwoot } = await importHook();
        const first = renderHook(() => useChatwoot());

        // Force a bootstrap timeout.
        const script1 = document.getElementById(
          "shellhub-chatwoot-sdk",
        ) as HTMLScriptElement | null;
        window.chatwootSDK = { run: vi.fn() };
        script1!.onload?.(new Event("load"));
        await act(async () => {
          await vi.advanceTimersByTimeAsync(15_001);
        });
        first.rerender();
        expect(first.result.current.status).toBe("unavailable");

        // Tenant change (simulating namespace switch) re-runs the injection
        // effect with a fresh identifier — bootstrapFailed must clear so
        // the new attempt can succeed.
        first.unmount();
        setupIdentifier("xyz789");
        const second = renderHook(() => useChatwoot());

        // Fresh script tag injected; bootstrapFailed flag cleared.
        const script2 = document.getElementById(
          "shellhub-chatwoot-sdk",
        ) as HTMLScriptElement | null;
        expect(script2).not.toBeNull();

        // Simulate successful bootstrap on the retry.
        window.chatwootSDK = { run: vi.fn() };
        script2!.onload?.(new Event("load"));
        await act(async () => {
          window.$chatwoot = {
            setUser: vi.fn(),
            toggle: vi.fn(),
            reset: vi.fn(),
            setConversationCustomAttributes: vi.fn(),
            setCustomAttributes: vi.fn(),
            deleteCustomAttribute: vi.fn(),
          };
          window.dispatchEvent(new Event("chatwoot:ready"));
        });
        second.rerender();

        expect(second.result.current.status).toBe("ready");
      } finally {
        vi.useRealTimers();
      }
    });
  });
});
