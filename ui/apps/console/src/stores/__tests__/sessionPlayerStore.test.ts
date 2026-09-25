import { describe, it, expect, beforeEach, vi } from "vitest";
import {
  nextPlayerControls,
  useSessionPlayerStore,
} from "../sessionPlayerStore";

beforeEach(() => {
  localStorage.clear();
  useSessionPlayerStore.setState({ controls: "auto" });
});

describe("sessionPlayerStore", () => {
  it("remembers the choice across a reload", async () => {
    useSessionPlayerStore.getState().setControls("hidden");

    vi.resetModules();
    const { useSessionPlayerStore: reloaded } =
      await import("../sessionPlayerStore");

    expect(reloaded.getState().controls).toBe("hidden");
  });

  it.each([
    ["auto", "always"],
    ["always", "hidden"],
    ["hidden", "auto"],
  ] as const)("steps from %s to %s", (from, to) => {
    expect(nextPlayerControls(from).value).toBe(to);
  });
});
