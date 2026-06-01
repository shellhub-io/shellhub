import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import EmptyState from "@/components/common/EmptyState";

const features = [
  {
    icon: <svg data-testid="f0" />,
    title: "Direct Access",
    description: "Routes directly.",
  },
  {
    icon: <svg data-testid="f1" />,
    title: "Device-side TLS",
    description: "Handles TLS locally.",
  },
];

describe("EmptyState", () => {
  it("renders the overline, title, description, children and footnote", () => {
    render(
      <EmptyState
        icon={<svg data-testid="hero" />}
        overline="HTTP Tunneling"
        title="Web Endpoints"
        description="Tunnel HTTP traffic to your devices."
        footnote="No VPN required."
      >
        <button type="button">Create your first endpoint</button>
      </EmptyState>,
    );

    expect(screen.getByText("HTTP Tunneling")).toBeInTheDocument();
    expect(
      screen.getByRole("heading", { level: 1, name: "Web Endpoints" }),
    ).toBeInTheDocument();
    expect(
      screen.getByText("Tunnel HTTP traffic to your devices."),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create your first endpoint" }),
    ).toBeInTheDocument();
    expect(screen.getByText("No VPN required.")).toBeInTheDocument();
  });

  it("names the region via its heading (aria-labelledby -> h1)", () => {
    render(
      <EmptyState
        icon={<svg />}
        overline="HTTP Tunneling"
        title="Web Endpoints"
        description="desc"
      />,
    );

    // The <section> is an accessible region named by the <h1>.
    expect(
      screen.getByRole("region", { name: "Web Endpoints" }),
    ).toBeInTheDocument();
  });

  it("renders the features as a list when provided", () => {
    render(
      <EmptyState
        icon={<svg />}
        overline="o"
        title="t"
        description="d"
        features={features}
      />,
    );

    const list = screen.getByRole("list");
    expect(within(list).getAllByRole("listitem")).toHaveLength(2);
    expect(
      screen.getByRole("heading", { level: 2, name: "Direct Access" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Handles TLS locally.")).toBeInTheDocument();
  });

  it("omits the features list when none are provided", () => {
    render(
      <EmptyState icon={<svg />} overline="o" title="t" description="d" />,
    );
    expect(screen.queryByRole("list")).not.toBeInTheDocument();
  });

  it("applies the yellow accent to the overline", () => {
    render(
      <EmptyState
        icon={<svg />}
        overline="Vault Locked"
        title="Locked"
        description="d"
        accent="yellow"
      />,
    );
    expect(screen.getByText("Vault Locked")).toHaveClass(
      "text-accent-yellow/80",
    );
  });

  it("defaults to the primary accent", () => {
    render(
      <EmptyState
        icon={<svg />}
        overline="Networking"
        title="t"
        description="d"
      />,
    );
    expect(screen.getByText("Networking")).toHaveClass("text-primary/80");
  });
});
