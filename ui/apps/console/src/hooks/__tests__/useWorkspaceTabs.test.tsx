import { describe, it, expect, beforeEach, vi } from "vitest";
import { act, renderHook } from "@testing-library/react";
import { QueryClient } from "@tanstack/react-query";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore, VALID_JWT } from "@/tests/seedAuthStore";
import { useAuthStore } from "@/stores/authStore";
import { useTerminalStore } from "@/stores/terminalStore";
import { useVaultStore } from "@/stores/vaultStore";
import {
  ACCOUNT_TAB_ID,
  namespaceTab,
  useWorkspaceTabsStore,
} from "@/stores/workspaceTabsStore";
import { mockNamespace } from "@/tests/factories";
import { ADMIN_UNAUTHORIZED_PATH } from "@/utils/adminRoute";
import { useNamespaces } from "../useNamespaces";
import { useWorkspaceTabs } from "../useWorkspaceTabs";

const home = namespaceTab("tenant-home", "home", "/devices");
const other = namespaceTab("tenant-other", "other", "/sessions");

function openTerminal() {
  useTerminalStore.getState().open({
    deviceUid: "dev-1",
    deviceName: "dev-1",
    username: "root",
    password: "",
    tenant: "tenant-home",
  });
}

function renderTabs(queryClient = new QueryClient()) {
  return renderHook(() => useWorkspaceTabs(), {
    wrapper: createTestWrapper({ queryClient, initialEntries: ["/devices"] }),
  });
}

beforeEach(() => {
  seedAuthStore({ tenant: "tenant-home" });
  useWorkspaceTabsStore.setState({ tabs: [home, other], failures: {} });
  useTerminalStore.setState({ sessions: [], restoreAfterNavigation: null });
  useVaultStore.setState({ status: "unlocked" });
  server.use(
    http.get("*/api/namespaces", () =>
      jsonWithTotal([
        mockNamespace({ tenant_id: "tenant-home", name: "home" }),
        mockNamespace({ tenant_id: "tenant-other", name: "other" }),
      ]),
    ),
  );
});

describe("useWorkspaceTabs", () => {
  it("records why a namespace could not be entered and moves nothing", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    openTerminal();
    const { result } = renderTabs();

    let entered = true;
    await act(async () => {
      entered = await result.current.activate(other);
    });

    expect(entered).toBe(false);
    expect(useWorkspaceTabsStore.getState().failures[other.id]).toBeTruthy();
    expect(useAuthStore.getState().tenant).toBe("tenant-home");
    expect(useTerminalStore.getState().sessions[0].state).toBe("docked");
    expect(useVaultStore.getState().status).toBe("unlocked");
  });

  it("enters a namespace in place, dropping its cache but keeping the terminals", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: VALID_JWT, role: "owner" }),
      ),
    );
    const queryClient = new QueryClient();
    queryClient.setQueryData([{ _id: "getDevices" }], ["dev-1"]);
    queryClient.setQueryData([{ _id: "getNamespaces" }], ["home", "other"]);
    openTerminal();
    const { result } = renderTabs(queryClient);

    await act(async () => {
      await result.current.activate(other);
    });

    expect(useAuthStore.getState().tenant).toBe("tenant-other");
    expect(queryClient.getQueryData([{ _id: "getDevices" }])).toBeUndefined();
    expect(queryClient.getQueryData([{ _id: "getNamespaces" }])).toEqual([
      "home",
      "other",
    ]);
    expect(useTerminalStore.getState().sessions).toHaveLength(1);
  });

  it("drops the vault of the namespace being left and reads the next one's", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: VALID_JWT, role: "owner" }),
      ),
    );
    const { result } = renderTabs();

    await act(async () => {
      await result.current.activate(other);
    });

    await vi.waitFor(() =>
      expect(useVaultStore.getState().status).toBe("uninitialized"),
    );
  });

  it("clears a recorded failure once the namespace opens", async () => {
    useWorkspaceTabsStore.setState({ failures: { [home.id]: "offline" } });
    const { result } = renderTabs();

    await act(async () => {
      await result.current.activate(home);
    });

    expect(useWorkspaceTabsStore.getState().failures).toEqual({});
  });

  it("shows a terminal from another namespace by entering that namespace", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({ token: VALID_JWT, role: "owner" }),
      ),
    );
    useTerminalStore.getState().open({
      deviceUid: "dev-2",
      deviceName: "dev-2",
      username: "root",
      password: "",
      tenant: "tenant-other",
    });
    useTerminalStore.getState().minimizeAll();
    const [session] = useTerminalStore.getState().sessions;
    const { result } = renderHook(
      () => ({ tabs: useWorkspaceTabs(), list: useNamespaces() }),
      { wrapper: createTestWrapper({ initialEntries: ["/devices"] }) },
    );
    await vi.waitFor(() =>
      expect(result.current.list.namespaces).toHaveLength(2),
    );

    await act(async () => {
      await result.current.tabs.showTerminal(session);
    });

    expect(useAuthStore.getState().tenant).toBe("tenant-other");
    expect(useTerminalStore.getState().restoreAfterNavigation).toBe(session.id);
  });

  it("keeps the active tab when the neighbour it would move to cannot be entered", async () => {
    server.use(
      http.get("*/api/auth/token/:tenant", () =>
        HttpResponse.json({}, { status: 403 }),
      ),
    );
    const { result } = renderTabs();

    act(() => {
      result.current.close(home);
    });

    await vi.waitFor(() =>
      expect(useWorkspaceTabsStore.getState().failures[other.id]).toBeTruthy(),
    );
    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toContain(
      home.id,
    );
  });

  it("treats the profile as the account's context, not the namespace's", () => {
    const { result } = renderHook(() => useWorkspaceTabs(), {
      wrapper: createTestWrapper({ initialEntries: ["/profile"] }),
    });

    expect(result.current.activeId).toBe(ACCOUNT_TAB_ID);
  });

  it("keeps the namespace tab active where the admin console turned the user away", () => {
    const { result } = renderHook(() => useWorkspaceTabs(), {
      wrapper: createTestWrapper({ initialEntries: [ADMIN_UNAUTHORIZED_PATH] }),
    });

    expect(result.current.activeId).toBe(home.id);
  });
});
