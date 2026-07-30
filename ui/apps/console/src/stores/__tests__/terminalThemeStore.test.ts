import { beforeEach, describe, expect, it, vi } from "vitest";

const loadStore = async () => {
  const { useTerminalThemeStore } = await import("../terminalThemeStore");
  return useTerminalThemeStore;
};

describe("terminalThemeStore encoding", () => {
  beforeEach(() => {
    localStorage.clear();
    vi.resetModules();
  });

  it("defaults to UTF-8", async () => {
    const store = await loadStore();

    expect(store.getState().encoding).toBe("utf-8");
  });

  it("round-trips a selected encoding", async () => {
    const store = await loadStore();

    store.getState().setEncoding("gbk");

    expect(store.getState().encoding).toBe("gbk");
  });

  it("restores the selected encoding on reload", async () => {
    const store = await loadStore();
    store.getState().setEncoding("gbk");

    vi.resetModules();
    const reloaded = await loadStore();

    expect(reloaded.getState().encoding).toBe("gbk");
  });

  it("ignores an encoding it does not offer", async () => {
    const store = await loadStore();

    store.getState().setEncoding("not-a-real-encoding" as never);

    expect(store.getState().encoding).toBe("utf-8");
  });
});
