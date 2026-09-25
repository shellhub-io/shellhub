import { describe, it, expect } from "vitest";
import { moveById } from "../moveById";

const list = [{ id: "a" }, { id: "b" }, { id: "c" }];
const ids = (items: { id: string }[]) => items.map((i) => i.id);

describe("moveById", () => {
  it.each([
    ["a", 2, ["b", "c", "a"]],
    ["c", 0, ["c", "a", "b"]],
  ])("moves %s to %i", (id, to, expected) => {
    expect(ids(moveById(list, id, to))).toEqual(expected);
  });

  it.each([
    ["x", 0],
    ["a", -1],
    ["a", 3],
    ["b", 1],
  ])("returns the list itself when %s cannot move to %i", (id, to) => {
    expect(moveById(list, id, to)).toBe(list);
  });

  it("leaves its input alone", () => {
    moveById(list, "a", 2);

    expect(ids(list)).toEqual(["a", "b", "c"]);
  });
});
