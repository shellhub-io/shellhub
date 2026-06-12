import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";

import PageHeader from "@/components/common/PageHeader";

describe("PageHeader", () => {
  it("renders the overline, title and icon", () => {
    render(
      <PageHeader
        icon={<svg data-testid="icon" />}
        overline="Devices"
        title="All Devices"
      />,
    );
    expect(screen.getByText("Devices")).toBeInTheDocument();
    expect(screen.getByText("All Devices")).toBeInTheDocument();
    expect(screen.getByTestId("icon")).toBeInTheDocument();
  });

  it("renders description when provided", () => {
    render(
      <PageHeader
        icon={<svg />}
        overline="Devices"
        title="All Devices"
        description="Manage your devices."
      />,
    );
    expect(screen.getByText("Manage your devices.")).toBeInTheDocument();
  });

  it("renders children (action area) when provided", () => {
    render(
      <PageHeader icon={<svg />} overline="Devices" title="All Devices">
        <button type="button">Add Device</button>
      </PageHeader>,
    );
    expect(
      screen.getByRole("button", { name: "Add Device" }),
    ).toBeInTheDocument();
  });

  it("renders correct color classes when iconColor is set", () => {
    const { container } = render(
      <PageHeader
        icon={<svg />}
        overline="Sessions"
        title="All Sessions"
        iconColor="cyan"
      />,
    );
    const badge = container.querySelector(".w-12.h-12");
    expect(badge).not.toBeNull();
    expect(badge!.className).toContain("bg-accent-cyan/10");
    expect(badge!.className).toContain("border-accent-cyan/20");
    expect(badge!.className).toContain("text-accent-cyan");
  });
});
