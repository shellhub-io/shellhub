import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";

function refuseStorage() {
  const refuse = () => {
    throw new DOMException("denied", "SecurityError");
  };
  vi.spyOn(Storage.prototype, "getItem").mockImplementation(refuse);
  vi.spyOn(Storage.prototype, "setItem").mockImplementation(refuse);
  vi.spyOn(Storage.prototype, "removeItem").mockImplementation(refuse);
}

async function loadStore() {
  vi.resetModules();
  return (await import("../themeStore")).useThemeStore;
}

beforeEach(() => {
  localStorage.clear();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("themeStore", () => {
  it.each(["light", "dark"] as const)(
    "keeps a %s choice across a reload",
    async (choice) => {
      (await loadStore()).getState().setPreference(choice);

      const reloaded = await loadStore();

      expect(reloaded.getState().preference).toBe(choice);
    },
  );

  it("forgets a fixed choice once the system is chosen again", async () => {
    const store = await loadStore();
    store.getState().setPreference("light");
    store.getState().setPreference("system");

    const reloaded = await loadStore();

    expect(reloaded.getState().preference).toBe("system");
  });

  it("follows the system when storage refuses to be read", async () => {
    localStorage.setItem("appTheme", "light");
    refuseStorage();

    const store = await loadStore();

    expect(store.getState().preference).toBe("system");
  });

  it("still switches theme when storage refuses the choice", async () => {
    const store = await loadStore();
    refuseStorage();

    store.getState().setPreference("light");

    expect(store.getState().theme).toBe("light");
    expect(document.documentElement.classList.contains("light")).toBe(true);
  });
});
