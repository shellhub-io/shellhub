import { describe, it, expect } from "vitest";
import {
  draggedOffset,
  dropIndex,
  makeRoomOffset,
  type TabDrag,
} from "../tabDrag";

const slots = [
  { id: "a", left: 0, width: 100 },
  { id: "b", left: 102, width: 100 },
  { id: "c", left: 204, width: 100 },
];

const dragOf = (id: string, offset: number): TabDrag => ({
  id,
  slots,
  offset,
});

describe("tab drag", () => {
  it("keeps the dragged tab inside its group", () => {
    expect(draggedOffset(dragOf("a", -50))).toBe(0);
    expect(draggedOffset(dragOf("a", 500))).toBe(204);
  });

  it.each([
    ["a", 40, 0],
    ["a", 60, 1],
    ["a", 170, 2],
    ["c", -60, 1],
    ["c", -170, 0],
  ])("drops %s moved by %i at index %i", (id, offset, index) => {
    expect(dropIndex(dragOf(id, offset))).toBe(index);
  });

  it("slides the tabs it passes towards where it came from", () => {
    const drag = dragOf("a", 170);

    expect(makeRoomOffset(drag, "b")).toBe(-102);
    expect(makeRoomOffset(drag, "c")).toBe(-102);
  });

  it("slides the tabs it passes and leaves the rest where they are", () => {
    const drag = dragOf("c", -60);

    expect(makeRoomOffset(drag, "a")).toBe(0);
    expect(makeRoomOffset(drag, "b")).toBe(102);
  });

  it("makes room for the gap the tabs are laid out with", () => {
    const spaced = [
      { id: "a", left: 0, width: 100 },
      { id: "b", left: 108, width: 100 },
    ];

    expect(makeRoomOffset({ id: "a", slots: spaced, offset: 90 }, "b")).toBe(
      -108,
    );
  });
});
