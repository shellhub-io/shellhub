import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Alert, { type AlertVariant } from "@/components/common/Alert";

describe("Alert", () => {
  describe("variant rendering", () => {
    it.each<[AlertVariant, string]>([
      ["error", "bg-accent-red/8"],
      ["success", "bg-accent-green/8"],
      ["warning", "bg-accent-yellow/8"],
      ["info", "bg-accent-blue/8"],
    ])("variant=%s applies the correct background class", (variant, expected) => {
      const { container } = render(
        <Alert variant={variant}>message</Alert>,
      );
      expect((container.firstChild as HTMLElement).className).toContain(expected);
    });

    it.each<[AlertVariant, string]>([
      ["error", "border-accent-red/20"],
      ["success", "border-accent-green/20"],
      ["warning", "border-accent-yellow/20"],
      ["info", "border-accent-blue/20"],
    ])("variant=%s applies the correct border class", (variant, expected) => {
      const { container } = render(
        <Alert variant={variant}>message</Alert>,
      );
      expect((container.firstChild as HTMLElement).className).toContain(expected);
    });

    it.each<[AlertVariant, string]>([
      ["error", "text-accent-red"],
      ["success", "text-accent-green"],
      ["warning", "text-accent-yellow"],
      ["info", "text-accent-blue"],
    ])("variant=%s applies the correct text class", (variant, expected) => {
      const { container } = render(
        <Alert variant={variant}>message</Alert>,
      );
      expect((container.firstChild as HTMLElement).className).toContain(expected);
    });
  });

  describe("semantic roles", () => {
    it("error variant uses role=alert and aria-live=assertive", () => {
      const { container } = render(<Alert variant="error">Error</Alert>);
      const el = container.firstChild as HTMLElement;
      expect(el).toHaveAttribute("role", "alert");
      expect(el).toHaveAttribute("aria-live", "assertive");
    });

    it("warning variant uses role=alert and aria-live=assertive", () => {
      const { container } = render(<Alert variant="warning">Warning</Alert>);
      const el = container.firstChild as HTMLElement;
      expect(el).toHaveAttribute("role", "alert");
      expect(el).toHaveAttribute("aria-live", "assertive");
    });

    it("success variant uses role=status and aria-live=polite", () => {
      const { container } = render(<Alert variant="success">Success</Alert>);
      const el = container.firstChild as HTMLElement;
      expect(el).toHaveAttribute("role", "status");
      expect(el).toHaveAttribute("aria-live", "polite");
    });

    it("info variant uses role=status and aria-live=polite", () => {
      const { container } = render(<Alert variant="info">Info</Alert>);
      const el = container.firstChild as HTMLElement;
      expect(el).toHaveAttribute("role", "status");
      expect(el).toHaveAttribute("aria-live", "polite");
    });
  });

  describe("content rendering", () => {
    it("renders plain string children", () => {
      render(<Alert variant="error">Something went wrong</Alert>);
      expect(screen.getByText("Something went wrong")).toBeInTheDocument();
    });

    it("renders ReactNode children with nested elements", () => {
      render(
        <Alert variant="error">
          <span>Error: <strong>field required</strong></span>
        </Alert>,
      );
      expect(screen.getByText(/field required/)).toBeInTheDocument();
    });

    it("is findable via its role", () => {
      render(<Alert variant="error">Error message</Alert>);
      expect(screen.getByRole("alert")).toBeInTheDocument();
    });

    it("is findable via role=status for success", () => {
      render(<Alert variant="success">Saved successfully</Alert>);
      expect(screen.getByRole("status")).toBeInTheDocument();
    });
  });

  describe("icon rendering", () => {
    it("renders an icon SVG for every variant", () => {
      const { container } = render(<Alert variant="error">msg</Alert>);
      expect(container.querySelector("svg")).not.toBeNull();
    });

    it.each<AlertVariant>(["error", "success", "warning", "info"])(
      "renders exactly one icon for variant=%s",
      (variant) => {
        const { container } = render(<Alert variant={variant}>msg</Alert>);
        expect(container.querySelectorAll("svg")).toHaveLength(1);
      },
    );
  });

  describe("dismiss behavior", () => {
    it("does not render a dismiss button when onDismiss is not provided", () => {
      render(<Alert variant="error">Error</Alert>);
      expect(
        screen.queryByRole("button", { name: /dismiss/i }),
      ).not.toBeInTheDocument();
    });

    it("renders a dismiss button when onDismiss is provided", () => {
      render(
        <Alert variant="error" onDismiss={() => {}}>
          Error
        </Alert>,
      );
      expect(
        screen.getByRole("button", { name: "Dismiss alert" }),
      ).toBeInTheDocument();
    });

    it("calls onDismiss when the dismiss button is clicked", async () => {
      const user = userEvent.setup();
      const onDismiss = vi.fn();
      render(
        <Alert variant="error" onDismiss={onDismiss}>
          Error
        </Alert>,
      );
      await user.click(screen.getByRole("button", { name: "Dismiss alert" }));
      expect(onDismiss).toHaveBeenCalledOnce();
    });

    it("dismiss button has type=button to avoid form submission", () => {
      render(
        <Alert variant="error" onDismiss={() => {}}>
          Error
        </Alert>,
      );
      const btn = screen.getByRole("button", { name: "Dismiss alert" });
      expect(btn).toHaveAttribute("type", "button");
    });
  });

  describe("className pass-through", () => {
    it("appends extra className to the root element", () => {
      const { container } = render(
        <Alert variant="error" className="mb-4">
          Error
        </Alert>,
      );
      expect((container.firstChild as HTMLElement).className).toContain("mb-4");
    });

    it("preserves built-in classes when className is provided", () => {
      const { container } = render(
        <Alert variant="error" className="mb-4">
          Error
        </Alert>,
      );
      const cls = (container.firstChild as HTMLElement).className;
      expect(cls).toContain("rounded-md");
      expect(cls).toContain("font-mono");
    });

    it("renders without errors when className is undefined", () => {
      expect(() =>
        render(<Alert variant="success">OK</Alert>),
      ).not.toThrow();
    });
  });

  describe("base styling", () => {
    it("always applies the compact inline layout classes", () => {
      const { container } = render(<Alert variant="info">msg</Alert>);
      const cls = (container.firstChild as HTMLElement).className;
      expect(cls).toContain("flex items-center gap-2");
      expect(cls).toContain("px-3.5 py-2.5");
      expect(cls).toContain("rounded-md");
      expect(cls).toContain("text-xs font-mono");
      expect(cls).toContain("animate-slide-down");
    });
  });
});
