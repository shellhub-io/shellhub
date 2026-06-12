import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Card } from "../primitives/Card";

describe("Card", () => {
  it("renders children", () => {
    render(<Card>hello card</Card>);
    expect(screen.getByText("hello card")).toBeInTheDocument();
  });

  it('as="a" renders an <a> element and forwards href', () => {
    render(
      <Card as="a" href="https://example.com">
        link card
      </Card>,
    );
    const el = screen.getByText("link card");
    expect(el.tagName.toLowerCase()).toBe("a");
    expect(el).toHaveAttribute("href", "https://example.com");
  });

  it("hover prop adds hover classes", () => {
    render(<Card hover data-testid="card" />);
    const el = screen.getByTestId("card");
    expect(el.className).toContain("transition-all");
    expect(el.className).toContain("duration-300");
    expect(el.className).toContain("hover:border-border-light");
  });

  it('className="rounded-lg" override removes rounded-xl via cn merge', () => {
    render(<Card className="rounded-lg" data-testid="card" />);
    const el = screen.getByTestId("card");
    expect(el.className).toContain("rounded-lg");
    expect(el.className).not.toContain("rounded-xl");
  });
});
