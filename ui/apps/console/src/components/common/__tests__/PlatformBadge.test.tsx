import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import PlatformBadge from "../PlatformBadge";

describe("PlatformBadge", () => {
  it.each([
    ["docker", "Docker", "accent-blue"],
    ["native", "Native", "accent-green"],
  ] as const)("the %s variant reads '%s' in %s", (platform, label, colour) => {
    render(<PlatformBadge platform={platform} />);
    const chip = screen.getByText(label).closest("span");
    expect(chip).not.toBeNull();
    expect(chip!.className).toContain(`bg-${colour}/10`);
    expect(chip!.className).toContain(`text-${colour}`);
  });
});
