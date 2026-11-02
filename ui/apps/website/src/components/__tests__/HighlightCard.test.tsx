import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { HighlightCard } from "@/components";

describe("HighlightCard", () => {
  it("renders content and forwards HTML attributes", () => {
    render(
      <HighlightCard color="primary" role="region" aria-label="Feature" className="p-8">
        Card content
      </HighlightCard>,
    );

    const card = screen.getByRole("region", { name: "Feature" });
    expect(card).toHaveTextContent("Card content");
    expect(card).toHaveClass("bg-card", "p-8");
  });
});
