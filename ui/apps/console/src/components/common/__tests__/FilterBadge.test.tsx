import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import FilterBadge from "../FilterBadge";

describe("FilterBadge", () => {
  describe("tags variant — uses Badge component with color=primary", () => {
    it("renders a chip for each tag", () => {
      render(
        <FilterBadge
          filter={{
            tags: [
              { name: "web", tenant_id: "", created_at: "", updated_at: "" },
              { name: "iot", tenant_id: "", created_at: "", updated_at: "" },
            ],
          }}
        />,
      );
      expect(screen.getByText("web")).toBeInTheDocument();
      expect(screen.getByText("iot")).toBeInTheDocument();
    });
  });

  describe("hostname variant — stays inline (font-mono, no font-weight — NOT a Badge)", () => {
    it("renders the hostname text", () => {
      render(<FilterBadge filter={{ hostname: "web-server-01" }} />);
      expect(screen.getByText("web-server-01")).toBeInTheDocument();
    });

    it("chip keeps font-mono (not font-medium) — non-lossless, must stay inline", () => {
      render(<FilterBadge filter={{ hostname: "web-server-01" }} />);
      const chip = screen.getByText("web-server-01").closest("span");
      expect(chip!.className).toContain("font-mono");
      // Badge default shape adds font-medium which would be wrong here
      expect(chip!.className).not.toContain("font-medium");
    });
  });

  describe("all-devices variant — stays inline (bg-hover-medium, no palette — NOT a Badge)", () => {
    it("renders 'All devices' when filter is empty", () => {
      render(<FilterBadge filter={{}} />);
      expect(screen.getByText("All devices")).toBeInTheDocument();
    });

    it("renders 'All devices' when hostname is '.*'", () => {
      render(<FilterBadge filter={{ hostname: ".*" }} />);
      expect(screen.getByText("All devices")).toBeInTheDocument();
    });

    it("chip uses bg-hover-medium (not a palette colour — must stay inline)", () => {
      render(<FilterBadge filter={{}} />);
      const chip = screen.getByText("All devices").closest("span");
      expect(chip!.className).toContain("bg-hover-medium");
    });
  });
});
