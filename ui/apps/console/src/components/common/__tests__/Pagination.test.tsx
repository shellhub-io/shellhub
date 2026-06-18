import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import Pagination from "../Pagination";

describe("Pagination", () => {
  it("Prev and Next buttons have explicit type='button'", () => {
    render(<Pagination page={1} totalPages={3} onPageChange={() => {}} />);
    // must pass totalPages >= 2 so buttons render
    const buttons = screen.getAllByRole("button");
    buttons.forEach((btn) => {
      expect(btn).toHaveAttribute("type", "button");
    });
  });
});
