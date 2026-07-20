import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { CTABanner } from "@/components/marketing/CTABanner";

vi.mock("@shellhub/design-system/components", () => ({
  Reveal: ({
    children,
    className,
  }: {
    children: React.ReactNode;
    className?: string;
  }) => (
    <div data-testid="reveal" className={className}>
      {children}
    </div>
  ),
  ConnectionGrid: () => <div data-testid="connection-grid" />,
}));

function renderBanner(
  props?: Partial<React.ComponentProps<typeof CTABanner>>,
) {
  const defaults: React.ComponentProps<typeof CTABanner> = {
    eyebrow: "Ready?",
    title: "Get started today",
    subtitle: "Deploy in minutes.",
    primaryAction: { label: "Start", to: "/getting-started" },
    secondaryAction: { label: "Pricing", to: "/pricing" },
    ...props,
  };

  return render(
    <MemoryRouter>
      <CTABanner {...defaults} />
    </MemoryRouter>,
  );
}

describe("CTABanner", () => {
  it("renders heading, eyebrow, subtitle, Reveal, and ConnectionGrid", () => {
    renderBanner();
    expect(
      screen.getByRole("heading", { level: 2, name: "Get started today" }),
    ).toBeInTheDocument();
    expect(screen.getByText("Ready?")).toBeInTheDocument();
    expect(screen.getByText("Deploy in minutes.")).toBeInTheDocument();
    expect(screen.getByTestId("reveal")).toBeInTheDocument();
    expect(screen.getByTestId("connection-grid")).toBeInTheDocument();
  });

  it.each([
    { slot: "primaryAction", label: "Go", to: "/go" },
    { slot: "secondaryAction", label: "Alt", to: "/alt" },
  ] as const)("$slot with `to` renders an internal link", ({ slot, label, to }) => {
    renderBanner({ [slot]: { label, to } });
    const link = screen.getByRole("link", { name: label });
    expect(link).toHaveAttribute("href", to);
    expect(link).not.toHaveAttribute("target");
  });

  it("href without external omits target and rel", () => {
    renderBanner({
      primaryAction: { label: "Email", href: "mailto:sales@shellhub.io" },
    });
    const link = screen.getByRole("link", { name: "Email" });
    expect(link).toHaveAttribute("href", "mailto:sales@shellhub.io");
    expect(link).not.toHaveAttribute("target");
  });

  it("href with external sets target and rel", () => {
    renderBanner({
      secondaryAction: {
        label: "Docs",
        href: "https://docs.shellhub.io",
        external: true,
      },
    });
    const link = screen.getByRole("link", { name: "Docs" });
    expect(link).toHaveAttribute("href", "https://docs.shellhub.io");
    expect(link).toHaveAttribute("target", "_blank");
    expect(link).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("uses primary→accent-cyan gradient by default", () => {
    const { container } = renderBanner();
    const overlay = container.querySelector("[class*='bg-gradient-to-br']");
    expect(overlay!.className).toContain("from-primary/[0.06]");
    expect(overlay!.className).toContain("to-accent-cyan/[0.04]");
  });

  it("accepts custom gradient colors", () => {
    const { container } = renderBanner({
      gradient: { from: "accent-green", to: "primary" },
    });
    const overlay = container.querySelector("[class*='bg-gradient-to-br']");
    expect(overlay!.className).toContain("from-accent-green/[0.06]");
    expect(overlay!.className).toContain("to-primary/[0.04]");
  });

  it.each([
    { color: "green" as const, expected: "text-accent-green" },
    { color: undefined, expected: "text-primary" },
  ])("eyebrowColor=$color applies $expected", ({ color, expected }) => {
    renderBanner({ eyebrowColor: color });
    expect(screen.getByText("Ready?")).toHaveClass(expected);
  });
});
