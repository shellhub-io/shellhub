import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import {
  ACCOUNT_TAB_ID,
  useWorkspaceTabsStore,
} from "@/stores/workspaceTabsStore";
import { ADMIN_UNAUTHORIZED_PATH, isAdminPath } from "@/utils/adminRoute";
import { useSyncWorkspaceTab } from "../useSyncWorkspaceTab";

function syncAt(pathname: string) {
  return renderHook(
    () => useSyncWorkspaceTab(pathname, isAdminPath(pathname)),
    {
      wrapper: createTestWrapper(),
    },
  );
}

beforeEach(() => {
  seedAuthStore();
  useWorkspaceTabsStore.setState({ tabs: [], failures: {} });
  server.use(http.get("*/api/namespaces", () => jsonWithTotal([])));
});

describe("useSyncWorkspaceTab", () => {
  it("opens the admin console's tab on its pages", () => {
    syncAt("/admin/dashboard");

    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toEqual([
      "admin",
    ]);
  });

  it("opens the account's tab on its pages", () => {
    syncAt("/profile");

    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toEqual([
      ACCOUNT_TAB_ID,
    ]);
  });

  it("opens no tab for someone the admin console turned away", () => {
    syncAt(ADMIN_UNAUTHORIZED_PATH);

    expect(useWorkspaceTabsStore.getState().tabs).toEqual([]);
  });
});
