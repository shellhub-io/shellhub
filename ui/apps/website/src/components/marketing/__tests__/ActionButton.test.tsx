import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { ActionButton, ActionButtonGroup } from "@/components/marketing";

function renderButton(
  props?: Partial<React.ComponentProps<typeof ActionButton>>,
) {
  const defaults: React.ComponentProps<typeof ActionButton> = {
    action: { label: "Get Started", to: "/getting-started" },
    ...props,
  };

  return render(
    <MemoryRouter>
      <ActionButton {...defaults} />
    </MemoryRouter>,
  );
}

describe("ActionButton", () => {
  describe("routing", () => {
    it("renders an internal link when action.to is set", () => {
      renderButton({ action: { label: "Go", to: "/go" } });
      const link = screen.getByRole("link", { name: "Go" });
      expect(link).toHaveAttribute("href", "/go");
      expect(link).not.toHaveAttribute("target");
    });

    it("renders an external link when action.href is set with external", () => {
      renderButton({
        action: {
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

    it("renders an anchor link without target when href has no external flag", () => {
      renderButton({
        action: { label: "Arch", href: "#architecture" },
      });
      const link = screen.getByRole("link", { name: "Arch" });
      expect(link).toHaveAttribute("href", "#architecture");
      expect(link).not.toHaveAttribute("target");
    });
  });

  describe("variant-driven defaults", () => {
    it("primary variant applies glow and ArrowRight icon by default", () => {
      const { container } = renderButton();
      const link = screen.getByRole("link");
      expect(link.className).toContain("shadow-primary/30");
      expect(
        container.querySelector("[aria-hidden='true']"),
      ).toBeInTheDocument();
    });

    it("outline variant has no glow and no icon by default", () => {
      const { container } = renderButton({ variant: "outline" });
      const link = screen.getByRole("link");
      expect(link.className).not.toContain("shadow-primary/30");
      expect(
        container.querySelector("[aria-hidden='true']"),
      ).not.toBeInTheDocument();
    });
  });

  describe("prop overrides", () => {
    it("size defaults to xl", () => {
      renderButton();
      const link = screen.getByRole("link");
      expect(link.className).toContain("py-3.5");
    });

    it("accepts a size prop", () => {
      renderButton({ size: "md" });
      const link = screen.getByRole("link");
      expect(link.className).toContain("py-2");
      expect(link.className).not.toContain("py-3.5");
    });

    it("glow can be explicitly disabled on primary", () => {
      renderButton({ glow: false });
      const link = screen.getByRole("link");
      expect(link.className).not.toContain("shadow-primary/30");
    });

    it("iconRight can be overridden", () => {
      renderButton({
        iconRight: <span data-testid="custom-icon" />,
      });
      expect(screen.getByTestId("custom-icon")).toBeInTheDocument();
    });

    it("iconRight={null} suppresses the default arrow", () => {
      const { container } = renderButton({
        iconRight: null,
      });
      expect(
        container.querySelector("[aria-hidden='true']"),
      ).not.toBeInTheDocument();
    });

    it("accepts a left icon prop", () => {
      renderButton({
        variant: "outline",
        icon: <span data-testid="github-icon" />,
      });
      expect(screen.getByTestId("github-icon")).toBeInTheDocument();
    });

    it("accepts fullWidth prop", () => {
      renderButton({ fullWidth: true });
      const link = screen.getByRole("link");
      expect(link).toHaveClass("w-full");
    });
  });
});

function renderGroup(
  props?: Partial<React.ComponentProps<typeof ActionButtonGroup>>,
) {
  const defaults: React.ComponentProps<typeof ActionButtonGroup> = {
    primaryAction: { label: "Get Started", to: "/getting-started" },
    secondaryAction: { label: "View Pricing", to: "/pricing" },
    ...props,
  };

  return render(
    <MemoryRouter>
      <ActionButtonGroup {...defaults} />
    </MemoryRouter>,
  );
}

describe("ActionButtonGroup", () => {
  it("renders both primary and secondary actions", () => {
    renderGroup();
    expect(
      screen.getByRole("link", { name: "Get Started" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "View Pricing" }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("link")).toHaveLength(2);
  });

  it("applies responsive flex layout", () => {
    const { container } = renderGroup();
    const wrapper = container.firstElementChild;
    expect(wrapper).toHaveClass(
      "flex",
      "flex-col",
      "sm:flex-row",
      "items-center",
      "justify-center",
      "gap-3",
    );
  });

  it("passes size to both buttons", () => {
    renderGroup({
      primaryAction: { label: "Start", to: "/start" },
      secondaryAction: { label: "Pricing", to: "/pricing" },
      size: "md",
    });
    const links = screen.getAllByRole("link");
    for (const link of links) {
      expect(link.className).toContain("py-2");
    }
  });
});
