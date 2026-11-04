import { describe, it, expect, beforeEach } from "vitest";
import { useCommandPaletteStore } from "../commandPaletteStore";

describe("commandPaletteStore", () => {
  beforeEach(() => {
    useCommandPaletteStore.setState({ view: null });
  });

  it("show opens the given view", () => {
    useCommandPaletteStore.getState().show("connect");
    expect(useCommandPaletteStore.getState().view).toBe("connect");
  });

  it("close closes whichever view is open", () => {
    useCommandPaletteStore.setState({ view: "connect" });
    useCommandPaletteStore.getState().close();
    expect(useCommandPaletteStore.getState().view).toBeNull();
  });
});
