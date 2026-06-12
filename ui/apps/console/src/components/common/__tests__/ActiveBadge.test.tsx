import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import ActiveBadge from "../ActiveBadge";

describe("ActiveBadge", () => {
  describe("active=true", () => {
    it("renders the text 'Active'", () => {
      render(<ActiveBadge active />);
      expect(screen.getByText("Active")).toBeInTheDocument();
    });

    it("chip carries green colour class (text-accent-green)", () => {
      render(<ActiveBadge active />);
      const chip = screen.getByText("Active").closest("span");
      expect(chip).not.toBeNull();
      expect(chip!.className).toContain("text-accent-green");
    });
  });

  describe("active=false", () => {
    it("renders the text 'Inactive'", () => {
      render(<ActiveBadge active={false} />);
      expect(screen.getByText("Inactive")).toBeInTheDocument();
    });

    it("chip carries yellow colour class (text-accent-yellow)", () => {
      render(<ActiveBadge active={false} />);
      const chip = screen.getByText("Inactive").closest("span");
      expect(chip).not.toBeNull();
      expect(chip!.className).toContain("text-accent-yellow");
    });
  });
});
