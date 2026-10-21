import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import FilterBadge from "../FilterBadge";

function tag(name: string) {
  return { name, tenant_id: "", created_at: "", updated_at: "" };
}

describe("FilterBadge", () => {
  it("renders a chip for each tag", () => {
    render(<FilterBadge filter={{ tags: [tag("web"), tag("iot")] }} />);
    expect(screen.getByText("web")).toBeInTheDocument();
    expect(screen.getByText("iot")).toBeInTheDocument();
  });

  it.each([[{}], [{ hostname: ".*" }]])(
    "renders 'All devices' for filter %j",
    (filter) => {
      render(<FilterBadge filter={filter} />);
      const chip = screen.getByText("All devices").closest("span");
      expect(chip!.className).toContain("bg-hover-medium");
    },
  );
});
