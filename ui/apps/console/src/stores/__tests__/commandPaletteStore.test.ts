import { describe, it, expect, beforeEach } from "vitest";
import { useCommandPaletteStore } from "../commandPaletteStore";

describe("commandPaletteStore", () => {
  beforeEach(() => {
    useCommandPaletteStore.setState({ open: false });
  });

  it("is closed by default", () => {
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });

  it("openPalette opens it", () => {
    useCommandPaletteStore.getState().openPalette();
    expect(useCommandPaletteStore.getState().open).toBe(true);
  });

  it("closePalette closes it", () => {
    useCommandPaletteStore.setState({ open: true });
    useCommandPaletteStore.getState().closePalette();
    expect(useCommandPaletteStore.getState().open).toBe(false);
  });
});
