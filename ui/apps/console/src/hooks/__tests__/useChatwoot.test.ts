import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook, act, waitFor } from "@testing-library/react";
import { getConfig, defaultConfig } from "@/env";
import { createTestWrapper } from "@/tests/wrapper";
import { useAuthStore } from "@/stores/authStore";
import { mockSdkResponse, makeSdkError } from "@/tests/sdk";
import { tearDownChatwoot } from "../chatwootRuntime";
import { useChatwoot } from "../useChatwoot";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getNamespace: vi.fn(),
    getNamespaceSupport: vi.fn(),
  }),
);

const mockGetConfig = vi.mocked(getConfig);

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

describe("useChatwoot", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    document.body.innerHTML = "";
    delete window.$chatwoot;
    delete window.chatwootSDK;
    delete window.chatwootSettings;
    tearDownChatwoot("logout");

    mockGetConfig.mockReturnValue({
      ...defaultConfig,
      edition: "cloud",
      chatwootWebsiteToken: "token-abc",
      chatwootBaseUrl: "https://chat.example.com",
    });
    useAuthStore.setState({
      userId: "user-1",
      email: "user@example.com",
      name: "Test User",
      tenant: "tenant-abc",
    });
    sdk.getNamespace.mockResolvedValue(
      mockSdkResponse({ name: "my-ns", billing: { active: true } }),
    );
    sdk.getNamespaceSupport.mockResolvedValue(
      mockSdkResponse({ identifier: "abc123" }),
    );
  });

  describe("status: non-cloud", () => {
    it("returns status 'non-cloud' when cloud=false", () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      expect(result.current.status).toBe("non-cloud");
    });

    it("does not inject a script when cloud=false", () => {
      mockGetConfig.mockReturnValue({ ...defaultConfig });

      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("status: unavailable", () => {
    it("returns 'unavailable' when chatwootWebsiteToken is missing", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
        chatwootWebsiteToken: "",
        chatwootBaseUrl: "https://chat.example.com",
      });

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      expect(result.current.status).toBe("unavailable");
    });

    it("returns 'unavailable' when chatwootBaseUrl is missing", () => {
      mockGetConfig.mockReturnValue({
        ...defaultConfig,
        edition: "cloud",
        chatwootWebsiteToken: "token-abc",
        chatwootBaseUrl: "",
      });

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      expect(result.current.status).toBe("unavailable");
    });
  });

  describe("status: loading (namespace not resolved)", () => {
    it("returns 'loading' when namespace is null", () => {
      sdk.getNamespace.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      expect(result.current.status).toBe("loading");
    });
  });

  describe("status: no-subscription", () => {
    it("returns 'no-subscription' when billing.active !== true", async () => {
      sdk.getNamespace.mockResolvedValue(
        mockSdkResponse({ name: "my-ns", billing: { active: false } }),
      );

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() =>
        expect(result.current.status).toBe("no-subscription"),
      );
    });

    it("returns 'no-subscription' when namespace has no billing object", async () => {
      sdk.getNamespace.mockResolvedValue(mockSdkResponse({ name: "my-ns" }));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() =>
        expect(result.current.status).toBe("no-subscription"),
      );
    });
  });

  describe("status: loading (identifier fetching)", () => {
    it("returns 'loading' while support identifier is being fetched", async () => {
      sdk.getNamespaceSupport.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => expect(sdk.getNamespaceSupport).toHaveBeenCalled());
      expect(result.current.status).toBe("loading");
    });

    it("returns 'loading' when identifier response has no identifier field", async () => {
      sdk.getNamespaceSupport.mockResolvedValue(mockSdkResponse({}));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => expect(sdk.getNamespaceSupport).toHaveBeenCalled());
      expect(result.current.status).toBe("loading");
    });
  });

  describe("status: ready", () => {
    it("returns 'ready' after the chatwoot:ready event fires", async () => {
      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() => expect(result.current.status).toBe("ready"));
    });
  });

  describe("script injection", () => {
    it("injects a script with the correct src when prerequisites are met", async () => {
      const baseURL = "https://chat.example.com";

      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        const script = document.getElementById("shellhub-chatwoot-sdk");
        expect(script).not.toBeNull();
        expect(script!.getAttribute("src")).toBe(`${baseURL}/packs/js/sdk.js`);
      });
    });

    it("does not inject a second script when the hook renders twice (StrictMode safety)", async () => {
      const { rerender } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      rerender();

      const scripts = document.querySelectorAll("#shellhub-chatwoot-sdk");
      expect(scripts.length).toBe(1);
    });
  });

  describe("window.chatwootSettings", () => {
    it("sets chatwootSettings with the correct values before script injection", async () => {
      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

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
      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() =>
        expect(window.$chatwoot!.setUser).toHaveBeenCalledWith("user-1", {
          email: "user@example.com",
          name: "Test User",
          identifier_hash: "abc123",
        }),
      );
    });

    it("does not call setUser again when re-rendered with the same identity", async () => {
      const { rerender } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() =>
        expect(window.$chatwoot!.setUser).toHaveBeenCalledTimes(1),
      );

      rerender();

      expect(window.$chatwoot!.setUser).toHaveBeenCalledTimes(1);
    });
  });

  describe("setConversationCustomAttributes", () => {
    it("calls setConversationCustomAttributes with namespace data on chatwoot:on-message", async () => {
      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

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
      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      act(() => {
        makeWidgetReady();
      });

      await waitFor(() => expect(result.current.status).toBe("ready"));

      act(() => {
        result.current.openWidget();
      });

      expect(window.$chatwoot!.toggle).toHaveBeenCalledWith("open");
    });

    it("does not call toggle when status is not ready (loading)", () => {
      sdk.getNamespace.mockReturnValue(new Promise(() => {}));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

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
    it("flips to 'unavailable' when /support returns an error", async () => {
      sdk.getNamespaceSupport.mockRejectedValue(makeSdkError(500));

      const { result } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => expect(result.current.status).toBe("unavailable"));
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("status: unavailable (bootstrap watchdog)", () => {
    it("flips to 'unavailable' when chatwoot:ready never fires after script load", async () => {
      vi.useFakeTimers();
      try {
        const { result, rerender } = renderHook(() => useChatwoot(), {
          wrapper: createTestWrapper(),
        });

        await act(async () => {
          await vi.advanceTimersByTimeAsync(100);
        });

        expect(result.current.status).toBe("loading");
        const script = document.getElementById(
          "shellhub-chatwoot-sdk",
        ) as HTMLScriptElement | null;
        expect(script).not.toBeNull();

        window.chatwootSDK = { run: vi.fn() };
        script!.onload?.(new Event("load"));

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
      const { result, rerender } = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      expect(result.current.status).toBe("loading");
      const script = document.getElementById(
        "shellhub-chatwoot-sdk",
      ) as HTMLScriptElement | null;

      await act(async () => {
        script!.onerror?.(new Event("error"));
      });
      rerender();

      expect(result.current.status).toBe("unavailable");
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });
  });

  describe("teardown via runtime helper", () => {
    it("tearDownChatwoot('logout') removes the script tag and clears window globals", async () => {
      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      makeWidgetReady();
      await waitFor(() => {
        expect(window.$chatwoot).toBeDefined();
      });

      tearDownChatwoot("logout");

      expect(window.$chatwoot).toBeUndefined();
      expect(window.chatwootSDK).toBeUndefined();
      expect(window.chatwootSettings).toBeUndefined();
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();
    });

    it("after teardown, a re-mount cleanly re-injects the script", async () => {
      const first = renderHook(() => useChatwoot(), {
        wrapper: createTestWrapper(),
      });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      first.unmount();

      tearDownChatwoot("logout");
      expect(document.getElementById("shellhub-chatwoot-sdk")).toBeNull();

      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });
    });

    it("removes SDK-injected DOM (chat bubble holders, iframe) on teardown", async () => {
      renderHook(() => useChatwoot(), { wrapper: createTestWrapper() });

      await waitFor(() => {
        expect(document.getElementById("shellhub-chatwoot-sdk")).not.toBeNull();
      });

      makeWidgetReady();

      const bubble = document.createElement("div");
      bubble.className = "woot-widget-holder";
      const altBubble = document.createElement("div");
      altBubble.className = "woot--bubble-holder";
      const iframe = document.createElement("iframe");
      iframe.src = "https://chatwoot.example.com/widget?token=abc";
      document.body.append(bubble, altBubble, iframe);

      tearDownChatwoot("logout");

      expect(document.querySelector(".woot-widget-holder")).toBeNull();
      expect(document.querySelector(".woot--bubble-holder")).toBeNull();
      expect(document.querySelector('iframe[src*="chatwoot"]')).toBeNull();
    });
  });

  describe("bootstrapFailed recovery", () => {
    it("clears bootstrapFailed when a fresh injection proceeds", async () => {
      vi.useFakeTimers();
      try {
        const first = renderHook(() => useChatwoot(), {
          wrapper: createTestWrapper(),
        });

        await act(async () => {
          await vi.advanceTimersByTimeAsync(100);
        });

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

        first.unmount();
        sdk.getNamespaceSupport.mockResolvedValue(
          mockSdkResponse({ identifier: "xyz789" }),
        );
        const second = renderHook(() => useChatwoot(), {
          wrapper: createTestWrapper(),
        });

        await act(async () => {
          await vi.advanceTimersByTimeAsync(100);
        });

        const script2 = document.getElementById(
          "shellhub-chatwoot-sdk",
        ) as HTMLScriptElement | null;
        expect(script2).not.toBeNull();

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
