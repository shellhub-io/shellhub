import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import ActiveBadge from "../ActiveBadge";

describe("ActiveBadge", () => {
  it.each([
    [true, "Active", "text-accent-green"],
    [false, "Inactive", "text-accent-yellow"],
  ])("active=%s reads '%s' in %s", (active, label, colour) => {
    render(<ActiveBadge active={active} />);
    const chip = screen.getByText(label).closest("span");
    expect(chip).not.toBeNull();
    expect(chip!.className).toContain(colour);
  });
});
