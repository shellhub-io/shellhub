import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusDot } from "../primitives/StatusDot";

describe("StatusDot", () => {
  describe("online (default)", () => {
    it("has aria-label Online and role img", () => {
      render(<StatusDot data-testid="dot" />);
      const el = screen.getByRole("img", { name: "Online" });
      expect(el).toBeInTheDocument();
    });

    it("contains animate-ping child (halo)", () => {
      render(<StatusDot data-testid="dot" />);
      const outer = screen.getByRole("img", { name: "Online" });
      const halo = outer.querySelector(".animate-ping");
      expect(halo).not.toBeNull();
    });

    it("applies green glow shadow by default", () => {
      render(<StatusDot data-testid="dot" />);
      const outer = screen.getByRole("img", { name: "Online" });
      // The solid dot (second child) carries the glow
      const solidDot = outer.children[1] as HTMLElement;
      expect(solidDot.className).toContain(
        "shadow-[0_0_6px_rgba(130,165,104,0.4)]",
      );
    });

    it("applies primary glow shadow when color=primary", () => {
      render(<StatusDot color="primary" />);
      const outer = screen.getByRole("img", { name: "Online" });
      const solidDot = outer.children[1] as HTMLElement;
      expect(solidDot.className).toContain(
        "shadow-[0_0_6px_rgba(102,122,204,0.4)]",
      );
    });

    it("applies yellow glow shadow when color=yellow", () => {
      render(<StatusDot color="yellow" />);
      const outer = screen.getByRole("img", { name: "Online" });
      const solidDot = outer.children[1] as HTMLElement;
      expect(solidDot.className).toContain(
        "shadow-[0_0_6px_rgba(191,140,93,0.4)]",
      );
    });

    it("halo has opacity-40 and no glow", () => {
      render(<StatusDot />);
      const outer = screen.getByRole("img", { name: "Online" });
      const halo = outer.children[0] as HTMLElement;
      expect(halo.className).toContain("opacity-40");
      expect(halo.className).not.toContain("shadow-[");
    });

    it("forwards className to outer span", () => {
      render(<StatusDot className="mx-auto" />);
      const outer = screen.getByRole("img", { name: "Online" });
      expect(outer.className).toContain("mx-auto");
    });

    it("outer span has relative flex classes", () => {
      render(<StatusDot />);
      const outer = screen.getByRole("img", { name: "Online" });
      expect(outer.className).toContain("relative");
      expect(outer.className).toContain("flex");
    });

    it("size sm applies h-1.5 w-1.5", () => {
      render(<StatusDot size="sm" />);
      const outer = screen.getByRole("img", { name: "Online" });
      expect(outer.className).toContain("h-1.5");
      expect(outer.className).toContain("w-1.5");
    });

    it("size md applies h-2.5 w-2.5", () => {
      render(<StatusDot size="md" />);
      const outer = screen.getByRole("img", { name: "Online" });
      expect(outer.className).toContain("h-2.5");
      expect(outer.className).toContain("w-2.5");
    });
  });

  describe("offline", () => {
    it("has aria-label Offline and role img", () => {
      render(<StatusDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el).toBeInTheDocument();
    });

    it("has block class (required for mx-auto centering)", () => {
      render(<StatusDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("block");
    });

    it("has bg-text-muted/30 class", () => {
      render(<StatusDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("bg-text-muted/30");
    });

    it("does NOT contain animate-ping", () => {
      render(<StatusDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      const halo = el.querySelector(".animate-ping");
      expect(halo).toBeNull();
    });

    it("forwards className to the single span", () => {
      render(<StatusDot online={false} className="mx-auto" />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("mx-auto");
    });

    it("size sm applies h-1.5 w-1.5", () => {
      render(<StatusDot online={false} size="sm" />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("h-1.5");
      expect(el.className).toContain("w-1.5");
    });

    it("size md applies h-2.5 w-2.5", () => {
      render(<StatusDot online={false} size="md" />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("h-2.5");
      expect(el.className).toContain("w-2.5");
    });

    it("is a single span (no children)", () => {
      render(<StatusDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.children).toHaveLength(0);
    });
  });
});
