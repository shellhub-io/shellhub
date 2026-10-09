import { describe, it, expect, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { http } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { seedAuthStore } from "@/tests/seedAuthStore";
import { defaultHandlers } from "@/tests/handlers";
import {
  ACCOUNT_TAB_ID,
  PREFERENCES_TAB_ID,
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
  server.use(...defaultHandlers);
  server.use(http.get("*/api/namespaces", () => jsonWithTotal([])));
});

describe("useSyncWorkspaceTab", () => {
  it("opens the admin console's tab on its pages", () => {
    syncAt("/admin/instance");

    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toEqual([
      "admin",
    ]);
  });

  it("opens the account's tab on its pages", () => {
    syncAt("/account");

    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toEqual([
      ACCOUNT_TAB_ID,
    ]);
  });

  it("opens the preferences' tab on their pages, remembering the section", () => {
    syncAt("/preferences/terminal");

    expect(useWorkspaceTabsStore.getState().tabs).toEqual([
      expect.objectContaining({
        id: PREFERENCES_TAB_ID,
        path: "/preferences/terminal",
      }),
    ]);
  });

  it("opens no tab for someone the admin console turned away", () => {
    syncAt(ADMIN_UNAUTHORIZED_PATH);

    expect(useWorkspaceTabsStore.getState().tabs).toEqual([]);
  });
});
