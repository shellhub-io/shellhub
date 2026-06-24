import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import SettingsCard from "@/components/common/SettingsCard";

describe("SettingsCard", () => {
  it("renders the title", () => {
    render(<SettingsCard title="Account Settings">content</SettingsCard>);
    expect(screen.getByText("Account Settings")).toBeInTheDocument();
  });

  it("renders children", () => {
    render(
      <SettingsCard title="Account Settings">
        <span>child content</span>
      </SettingsCard>,
    );
    expect(screen.getByText("child content")).toBeInTheDocument();
  });

  it("applies danger classes when danger prop is true", () => {
    const { container } = render(
      <SettingsCard title="Danger Zone" danger>
        content
      </SettingsCard>,
    );
    const card = container.firstChild as HTMLElement;
    expect(card.className).toContain("border-accent-red/20");
    expect(card.className).toContain("border-l-accent-red/40");
    const header = card.querySelector("div");
    expect(header!.className).toContain("border-accent-red/10");
    const heading = card.querySelector("h3");
    expect(heading!.className).toContain("text-accent-red");
  });

  it("applies default classes when danger prop is false", () => {
    const { container } = render(
      <SettingsCard title="Normal Section">content</SettingsCard>,
    );
    const card = container.firstChild as HTMLElement;
    expect(card.className).toContain("border-border");
    const header = card.querySelector("div");
    expect(header!.className).toContain("border-border");
    const heading = card.querySelector("h3");
    expect(heading!.className).toContain("text-text-primary");
  });
});
