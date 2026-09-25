import { describe, it, expect, beforeEach } from "vitest";
import {
  adminTab,
  namespaceTab,
  useWorkspaceTabsStore,
} from "../workspaceTabsStore";

beforeEach(() => {
  useWorkspaceTabsStore.setState({ tabs: [], failures: {} });
});

describe("workspaceTabsStore", () => {
  it("keeps one tab per namespace, renaming it when the name changes", () => {
    const store = useWorkspaceTabsStore.getState();
    store.ensure(namespaceTab("t1", "old"));
    store.ensure(namespaceTab("t1", "new", "/devices"));

    const { tabs } = useWorkspaceTabsStore.getState();
    expect(tabs).toHaveLength(1);
    expect(tabs[0]).toMatchObject({ name: "new", path: "/dashboard" });
  });

  it("drops the tabs of namespaces the user left, and never the admin tab", () => {
    const store = useWorkspaceTabsStore.getState();
    store.ensure(namespaceTab("kept", "kept"));
    store.ensure(namespaceTab("gone", "gone"));
    store.ensure(adminTab());

    store.retain(new Set(["kept"]));

    expect(useWorkspaceTabsStore.getState().tabs.map((t) => t.id)).toEqual([
      "ns:kept",
      "admin",
    ]);
  });

  it("persists the tabs but not why one failed to open", () => {
    const store = useWorkspaceTabsStore.getState();
    store.ensure(namespaceTab("t1", "one"));
    store.fail("ns:t1", "offline");

    const persisted = JSON.parse(localStorage.getItem("workspaceTabs") ?? "{}");
    expect(persisted.state).toEqual({
      tabs: [namespaceTab("t1", "one")],
    });
  });
});
