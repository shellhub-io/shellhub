import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { HighlightCard } from "@/components";

describe("HighlightCard", () => {
  it("renders children inside a card with a gradient overlay", () => {
    render(
      <HighlightCard color="primary" data-testid="card">
        <p>hello</p>
      </HighlightCard>,
    );
    const card = screen.getByTestId("card");
    expect(screen.getByText("hello")).toBeInTheDocument();
    expect(card).toHaveClass("relative", "bg-card", "border", "rounded-xl");

    const overlay = card.firstChild as HTMLElement;
    expect(overlay).toHaveClass("absolute", "inset-0", "pointer-events-none");
    expect(overlay.className).toContain("bg-gradient-to-br");
  });

  it.each([
    ["primary", "border-primary/30", "shadow-[0_0_40px_rgba(102,122,204,0.1)]", "from-primary/[0.06]"],
    ["accent-cyan", "border-accent-cyan/30", "shadow-[0_0_40px_rgba(78,154,163,0.08)]", "from-accent-cyan/[0.06]"],
    ["accent-green", "border-accent-green/30", "shadow-[0_0_40px_rgba(130,165,104,0.08)]", "from-accent-green/[0.06]"],
    ["accent-blue", "border-accent-blue/30", "shadow-[0_0_40px_rgba(86,162,225,0.06)]", "from-accent-blue/[0.06]"],
    ["accent-yellow", "border-accent-yellow/30", "shadow-[0_0_40px_rgba(191,140,93,0.06)]", "from-accent-yellow/[0.06]"],
  ] as const)("maps %s to its border, shadow, and gradient", (color, border, shadow, gradient) => {
    render(
      <HighlightCard color={color} data-testid="card">
        content
      </HighlightCard>,
    );
    const card = screen.getByTestId("card");
    expect(card.className).toContain(border);
    expect(card.className).toContain(shadow);
    expect((card.firstChild as HTMLElement).className).toContain(gradient);
  });

  it("merges className and forwards HTML attributes", () => {
    render(
      <HighlightCard
        color="primary"
        className="p-8 flex flex-col"
        aria-label="feature card"
        data-testid="card"
      >
        content
      </HighlightCard>,
    );
    const card = screen.getByTestId("card");
    expect(card).toHaveClass("p-8", "flex", "flex-col", "bg-card");
    expect(card).toHaveAttribute("aria-label", "feature card");
  });
});
