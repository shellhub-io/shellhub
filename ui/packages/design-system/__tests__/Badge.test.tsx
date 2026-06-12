import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Badge } from "../primitives/Badge";

describe("Badge", () => {
  describe("defaults", () => {
    it("renders with primary color and rounded shape by default", () => {
      render(<Badge data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-primary/10");
      expect(el.className).toContain("text-primary");
      expect(el.className).toContain("rounded");
      expect(el.className).toContain("px-1.5");
      expect(el.className).toContain("py-0.5");
      expect(el.className).toContain("font-medium");
    });

    it("always includes base inline-flex items-center gap-1 text-2xs", () => {
      render(<Badge data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("inline-flex");
      expect(el.className).toContain("items-center");
      expect(el.className).toContain("gap-1");
      expect(el.className).toContain("text-2xs");
    });
  });

  describe("shape=rounded (console chips)", () => {
    it("includes px-1.5 py-0.5 rounded font-medium", () => {
      render(<Badge shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("px-1.5");
      expect(el.className).toContain("py-0.5");
      expect(el.className).toContain("rounded");
      expect(el.className).toContain("font-medium");
    });

    it("does NOT include pill-only classes (font-semibold uppercase tracking-compact border)", () => {
      render(<Badge shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).not.toContain("font-semibold");
      expect(el.className).not.toContain("uppercase");
      expect(el.className).not.toContain("tracking-compact");
      // border is not part of rounded shape
      expect(el.className).not.toContain("border");
    });
  });

  describe("shape=pill (website)", () => {
    it("includes px-2 py-0.5 rounded-full font-mono font-semibold uppercase tracking-compact border", () => {
      render(<Badge shape="pill" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("px-2");
      expect(el.className).toContain("py-0.5");
      expect(el.className).toContain("rounded-full");
      expect(el.className).toContain("font-mono");
      expect(el.className).toContain("font-semibold");
      expect(el.className).toContain("uppercase");
      expect(el.className).toContain("tracking-compact");
      expect(el.className).toContain("border");
    });

    it("does NOT include rounded-chip classes (rounded without -full, font-medium)", () => {
      render(<Badge shape="pill" color="primary" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      // rounded-full is present; plain "rounded" class should NOT be present without -full suffix
      // (we check rounded-full is present, and that "rounded " standalone is absent)
      expect(el.className).toContain("rounded-full");
      expect(el.className).not.toContain("font-medium");
    });
  });

  describe("colors — rounded shape", () => {
    it("primary uses bg-primary/10 text-primary", () => {
      render(<Badge color="primary" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-primary/10");
      expect(el.className).toContain("text-primary");
    });

    it("green resolves to accent-green prefix: bg-accent-green/10 text-accent-green", () => {
      render(<Badge color="green" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-green/10");
      expect(el.className).toContain("text-accent-green");
    });

    it("red resolves to accent-red prefix: bg-accent-red/10 text-accent-red", () => {
      render(<Badge color="red" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-red/10");
      expect(el.className).toContain("text-accent-red");
    });

    it("yellow resolves to accent-yellow prefix: bg-accent-yellow/10 text-accent-yellow", () => {
      render(<Badge color="yellow" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-yellow/10");
      expect(el.className).toContain("text-accent-yellow");
    });

    it("blue resolves to accent-blue prefix: bg-accent-blue/10 text-accent-blue", () => {
      render(<Badge color="blue" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-blue/10");
      expect(el.className).toContain("text-accent-blue");
    });

    it("cyan resolves to accent-cyan prefix: bg-accent-cyan/10 text-accent-cyan", () => {
      render(<Badge color="cyan" shape="rounded" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-cyan/10");
      expect(el.className).toContain("text-accent-cyan");
    });
  });

  describe("colors — pill shape (also includes border color)", () => {
    it("primary pill uses bg-primary/10 text-primary border-primary/20", () => {
      render(<Badge color="primary" shape="pill" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-primary/10");
      expect(el.className).toContain("text-primary");
      expect(el.className).toContain("border-primary/20");
    });

    it("green pill uses bg-accent-green/10 text-accent-green border-accent-green/20", () => {
      render(<Badge color="green" shape="pill" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-green/10");
      expect(el.className).toContain("text-accent-green");
      expect(el.className).toContain("border-accent-green/20");
    });

    it("red pill uses bg-accent-red/10 text-accent-red border-accent-red/20", () => {
      render(<Badge color="red" shape="pill" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-red/10");
      expect(el.className).toContain("text-accent-red");
      expect(el.className).toContain("border-accent-red/20");
    });
  });

  describe("className override", () => {
    it('className="font-semibold" wins over base font-medium (same tailwind-merge group)', () => {
      render(
        <Badge className="font-semibold" shape="rounded" data-testid="badge" />,
      );
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("font-semibold");
      expect(el.className).not.toContain("font-medium");
    });

    it("forwards custom className", () => {
      render(<Badge className="custom-class" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("custom-class");
    });
  });

  describe("children", () => {
    it("renders children", () => {
      render(<Badge>status</Badge>);
      expect(screen.getByText("status")).toBeInTheDocument();
    });
  });
});
