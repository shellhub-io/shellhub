import { describe, it, expect } from "vitest";
import { render } from "@testing-library/react";
import { GlowOrbs } from "../components/GlowOrbs";

describe("GlowOrbs", () => {
  it.each([
    { name: "hero", element: <GlowOrbs preset="hero" /> },
    { name: "duo/primary", element: <GlowOrbs preset="duo" tone="primary" /> },
    {
      name: "section/primary",
      element: <GlowOrbs preset="section" tone="primary" />,
    },
    {
      name: "corner/primary",
      element: <GlowOrbs preset="corner" tone="primary" />,
    },
    {
      name: "ambient/brand",
      element: <GlowOrbs preset="ambient" tone="brand" />,
    },
  ])("renders an aria-hidden, roleless root div for $name", ({ element }) => {
    const { container } = render(element);
    const root = container.firstElementChild;

    expect(root).not.toBeNull();
    expect(root).toHaveAttribute("aria-hidden", "true");
    expect(root).not.toHaveAttribute("role");
  });
});
