import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

async function loadStore() {
  vi.resetModules();
  return (await import("../sidebarStore")).useSidebarStore;
}

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("sidebarStore", () => {
  it("starts automatic when nothing was chosen", async () => {
    expect((await loadStore()).getState().pin).toBe("auto");
  });

  it.each([
    ["true", "pinned"],
    ["false", "rail"],
  ])("reads a choice the layout stored as %s as %s", async (stored, pin) => {
    localStorage.setItem("sidebarPinned", stored);
    expect((await loadStore()).getState().pin).toBe(pin);
  });

  it.each(["pinned", "rail"] as const)(
    "keeps a %s choice across a reload",
    async (pin) => {
      (await loadStore()).getState().setPin(pin);
      expect((await loadStore()).getState().pin).toBe(pin);
    },
  );

  it("forgets an earlier choice when set back to automatic", async () => {
    (await loadStore()).getState().setPin("pinned");
    (await loadStore()).getState().setPin("auto");
    expect((await loadStore()).getState().pin).toBe("auto");
  });

  it("still switches when the browser refuses storage", async () => {
    vi.spyOn(Storage.prototype, "getItem").mockImplementation(() => {
      throw new DOMException("denied", "SecurityError");
    });
    vi.spyOn(Storage.prototype, "setItem").mockImplementation(() => {
      throw new DOMException("denied", "SecurityError");
    });
    const store = await loadStore();

    store.getState().setPin("rail");

    expect(store.getState().pin).toBe("rail");
  });
});
