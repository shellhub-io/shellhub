import { describe, it, expect } from "vitest";
import { nextFontSize } from "../fontSizeShortcut";
import { DEFAULT_FONT_SIZE } from "@/stores/terminalThemeStore";

const event = (init: Partial<KeyboardEventInit> & { key: string }) =>
  new KeyboardEvent("keydown", { ctrlKey: false, metaKey: false, altKey: false, shiftKey: false, ...init });

const CURRENT = 14;

describe("nextFontSize", () => {
  describe("with the ctrl modifier", () => {
    it.each([
      ["+", CURRENT + 1],
      ["=", CURRENT + 1],
      ["-", CURRENT - 1],
      ["0", DEFAULT_FONT_SIZE],
    ])("maps %s to %i", (key, expected) => {
      expect(nextFontSize(CURRENT, event({ key, ctrlKey: true }))).toBe(expected);
    });
  });

  describe("with the meta modifier", () => {
    it.each([
      ["+", CURRENT + 1],
      ["=", CURRENT + 1],
      ["-", CURRENT - 1],
      ["0", DEFAULT_FONT_SIZE],
    ])("maps %s to %i so macOS behaves the same", (key, expected) => {
      expect(nextFontSize(CURRENT, event({ key, metaKey: true }))).toBe(expected);
    });
  });

  it("steps from the size it is given, not from the default", () => {
    expect(nextFontSize(20, event({ key: "+", ctrlKey: true }))).toBe(21);
    expect(nextFontSize(20, event({ key: "-", ctrlKey: true }))).toBe(19);
  });

  it("resets to the default from any size", () => {
    expect(nextFontSize(24, event({ key: "0", ctrlKey: true }))).toBe(DEFAULT_FONT_SIZE);
  });

  it("treats a shifted plus as an increase, so US layouts need no extra keypress", () => {
    expect(nextFontSize(CURRENT, event({ key: "+", ctrlKey: true, shiftKey: true }))).toBe(CURRENT + 1);
  });

  it.each([
    ["NumpadAdd", "+", CURRENT + 1],
    ["NumpadSubtract", "-", CURRENT - 1],
  ])("handles %s via its key", (code, key, expected) => {
    expect(nextFontSize(CURRENT, event({ key, code, ctrlKey: true }))).toBe(expected);
  });

  it("ignores ctrl+_, leaving 0x1F (readline/emacs undo) for the remote shell", () => {
    expect(nextFontSize(CURRENT, event({ key: "_", ctrlKey: true }))).toBeNull();
  });

  it.each(["+", "-", "0", "="])("ignores %s without a modifier", (key) => {
    expect(nextFontSize(CURRENT, event({ key }))).toBeNull();
  });

  it.each(["+", "-", "0"])("ignores %s when alt is held, leaving it for the remote shell", (key) => {
    expect(nextFontSize(CURRENT, event({ key, ctrlKey: true, altKey: true }))).toBeNull();
  });

  it.each(["a", "c", "1", "9", "Enter", "ArrowUp", "constructor", "toString"])(
    "ignores the unrelated key %s",
    (key) => {
      expect(nextFontSize(CURRENT, event({ key, ctrlKey: true }))).toBeNull();
    },
  );
});
