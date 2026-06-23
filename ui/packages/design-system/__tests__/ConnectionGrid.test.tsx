import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import { ConnectionGrid } from "../components/ConnectionGrid";

describe("ConnectionGrid", () => {
  it("root .connection-grid div has aria-hidden=\"true\"", () => {
    const { container: c } = render(<ConnectionGrid />);
    const root = c.querySelector(".connection-grid");
    expect(root).not.toBeNull();
    expect(root).toHaveAttribute("aria-hidden", "true");
  });

  it("root .connection-grid div has no role attribute", () => {
    const { container: c } = render(<ConnectionGrid />);
    const root = c.querySelector(".connection-grid");
    expect(root).not.toBeNull();
    expect(root).not.toHaveAttribute("role");
  });
});
