import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { FeatureListItem } from "@/components";

describe("FeatureListItem", () => {
  it("renders an <li> with items-center and children", () => {
    render(<FeatureListItem>My label</FeatureListItem>);
    const li = screen.getByRole("listitem");
    expect(li).toHaveClass("items-center");
    expect(screen.getByText("My label")).toBeInTheDocument();
  });

  it("color='green' gives icon text-accent-green", () => {
    render(<FeatureListItem color="green">Label</FeatureListItem>);
    const li = screen.getByRole("listitem");
    const svg = li.querySelector("svg");
    expect(svg).not.toBeNull();
    expect(svg).toHaveClass("text-accent-green");
  });

  it("color='muted' (default) gives icon text-text-muted", () => {
    render(<FeatureListItem>Label</FeatureListItem>);
    const li = screen.getByRole("listitem");
    const svg = li.querySelector("svg");
    expect(svg).not.toBeNull();
    expect(svg).toHaveClass("text-text-muted");
  });
});
