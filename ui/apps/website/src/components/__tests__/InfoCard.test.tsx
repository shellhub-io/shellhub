import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { InfoCard, type InfoCardProps } from "@/components";
import { C } from "@shellhub/design-system/constants";

const color = C.primary;

function StarIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg data-testid="star-icon" {...props}>
      <path d="M0 0" />
    </svg>
  );
}

function renderCard(overrides: Partial<InfoCardProps> = {}, children?: React.ReactNode) {
  return render(
    <InfoCard
      icon={StarIcon}
      color={color}
      title="Title"
      description="Description"
      {...overrides}
    >
      {children}
    </InfoCard>,
  );
}

describe("InfoCard", () => {
  describe("vertical layout (default)", () => {
    it("renders icon with w-5 h-5 and color, plus title and description", () => {
      renderCard();

      const icon = screen.getByTestId("star-icon");
      expect(icon).toHaveClass("w-5", "h-5");
      expect(icon).toHaveStyle({ color });
      expect(screen.getByRole("heading", { name: "Title" })).toBeInTheDocument();
      expect(screen.getByText("Description")).toBeInTheDocument();
    });

    it("applies color to icon container background and border", () => {
      const { container } = renderCard();

      const iconContainer = container.querySelector(".w-10.h-10");
      expect(iconContainer).toHaveStyle({
        background: `${color}15`,
        borderColor: `${color}25`,
      });
    });

    it("renders icon container with mb-4 (no flex row)", () => {
      const { container } = renderCard();

      const iconContainer = container.querySelector(".w-10.h-10");
      expect(iconContainer).toHaveClass("mb-4");
      expect(iconContainer?.parentElement).not.toHaveClass("flex");
    });

    it("renders children after description", () => {
      renderCard({}, <span data-testid="extra">Extra</span>);
      expect(screen.getByTestId("extra")).toBeInTheDocument();
    });
  });

  describe("horizontal layout", () => {
    it("renders icon and text side by side", () => {
      const { container } = renderCard({ layout: "horizontal" });

      const flexRow = container.querySelector(".flex.items-start.gap-4");
      expect(flexRow).not.toBeNull();

      const iconContainer = container.querySelector(".w-10.h-10");
      expect(iconContainer).toHaveClass("shrink-0");
      expect(iconContainer).not.toHaveClass("mb-4");
    });
  });

  describe("dot layout", () => {
    it("renders a small dot instead of an icon container", () => {
      const { container } = renderCard({ layout: "dot" });

      const dot = container.querySelector(".w-2.h-2.rounded-full");
      expect(dot).not.toBeNull();
      expect(dot).toHaveStyle({ background: color });
      expect(container.querySelector(".w-10.h-10")).toBeNull();
    });

    it("renders children after description", () => {
      renderCard({ layout: "dot" }, <div data-testid="mockup">Mockup</div>);
      expect(screen.getByTestId("mockup")).toBeInTheDocument();
    });
  });

  describe("accessibility", () => {
    it("renders title as an h4 heading", () => {
      renderCard();
      expect(screen.getByRole("heading", { name: "Title" }).tagName).toBe("H4");
    });
  });
});
