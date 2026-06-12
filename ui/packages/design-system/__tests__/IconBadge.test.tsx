import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { IconBadge } from "../primitives/IconBadge";

describe("IconBadge", () => {
  describe("defaults", () => {
    it("renders with default primary/md classes", () => {
      render(<IconBadge data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-primary/10");
      expect(el.className).toContain("border-primary/20");
      expect(el.className).toContain("text-primary");
      expect(el.className).toContain("w-10");
      expect(el.className).toContain("h-10");
    });
  });

  describe("colors", () => {
    it("primary uses bg-primary/10 border-primary/20 text-primary", () => {
      render(<IconBadge color="primary" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-primary/10");
      expect(el.className).toContain("border-primary/20");
      expect(el.className).toContain("text-primary");
    });

    it("green uses bg-accent-green/10 border-accent-green/20 text-accent-green", () => {
      render(<IconBadge color="green" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-green/10");
      expect(el.className).toContain("border-accent-green/20");
      expect(el.className).toContain("text-accent-green");
    });

    it("red uses bg-accent-red/10 border-accent-red/20 text-accent-red", () => {
      render(<IconBadge color="red" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-red/10");
      expect(el.className).toContain("border-accent-red/20");
      expect(el.className).toContain("text-accent-red");
    });

    it("yellow uses bg-accent-yellow/10 border-accent-yellow/20 text-accent-yellow", () => {
      render(<IconBadge color="yellow" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-yellow/10");
      expect(el.className).toContain("border-accent-yellow/20");
      expect(el.className).toContain("text-accent-yellow");
    });

    it("blue uses bg-accent-blue/10 border-accent-blue/20 text-accent-blue", () => {
      render(<IconBadge color="blue" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-blue/10");
      expect(el.className).toContain("border-accent-blue/20");
      expect(el.className).toContain("text-accent-blue");
    });

    it("cyan uses bg-accent-cyan/10 border-accent-cyan/20 text-accent-cyan", () => {
      render(<IconBadge color="cyan" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-accent-cyan/10");
      expect(el.className).toContain("border-accent-cyan/20");
      expect(el.className).toContain("text-accent-cyan");
    });

    it("neutral uses bg-white/[0.04] border-border text-text-secondary", () => {
      render(<IconBadge color="neutral" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("bg-white/[0.04]");
      expect(el.className).toContain("border-border");
      expect(el.className).toContain("text-text-secondary");
    });
  });

  describe("sizes", () => {
    it("sm produces w-8 h-8 rounded-lg", () => {
      render(<IconBadge size="sm" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("w-8");
      expect(el.className).toContain("h-8");
      expect(el.className).toContain("rounded-lg");
    });

    it("md produces w-10 h-10 rounded-lg", () => {
      render(<IconBadge size="md" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("w-10");
      expect(el.className).toContain("h-10");
      expect(el.className).toContain("rounded-lg");
    });

    it("lg produces w-12 h-12 rounded-lg", () => {
      render(<IconBadge size="lg" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("w-12");
      expect(el.className).toContain("h-12");
      expect(el.className).toContain("rounded-lg");
    });
  });

  describe("base classes", () => {
    it("always includes flex items-center justify-center border", () => {
      render(<IconBadge data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("flex");
      expect(el.className).toContain("items-center");
      expect(el.className).toContain("justify-center");
      expect(el.className).toContain("border");
      expect(el.className).toContain("shrink-0");
    });
  });

  describe("children", () => {
    it("renders children", () => {
      render(<IconBadge>icon-content</IconBadge>);
      expect(screen.getByText("icon-content")).toBeInTheDocument();
    });
  });

  describe("className forwarding", () => {
    it("forwards custom className", () => {
      render(<IconBadge className="extra-class" data-testid="badge" />);
      const el = screen.getByTestId("badge");
      expect(el.className).toContain("extra-class");
    });
  });
});
