import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import PlatformBadge from "../PlatformBadge";

describe("PlatformBadge", () => {
  describe("docker variant — Badge color=blue", () => {
    it("renders the text 'Docker'", () => {
      render(<PlatformBadge platform="docker" />);
      expect(screen.getByText("Docker")).toBeInTheDocument();
    });

    it("chip carries blue colour classes (bg-accent-blue/10 text-accent-blue)", () => {
      render(<PlatformBadge platform="docker" />);
      const chip = screen.getByText("Docker").closest("span");
      expect(chip).not.toBeNull();
      expect(chip!.className).toContain("bg-accent-blue/10");
      expect(chip!.className).toContain("text-accent-blue");
    });
  });

  describe("native variant — Badge color=green", () => {
    it("renders the text 'Native'", () => {
      render(<PlatformBadge platform="native" />);
      expect(screen.getByText("Native")).toBeInTheDocument();
    });

    it("chip carries green colour classes (bg-accent-green/10 text-accent-green)", () => {
      render(<PlatformBadge platform="native" />);
      const chip = screen.getByText("Native").closest("span");
      expect(chip).not.toBeNull();
      expect(chip!.className).toContain("bg-accent-green/10");
      expect(chip!.className).toContain("text-accent-green");
    });
  });
});
