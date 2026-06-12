import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import OnlineDot from "../OnlineDot";

vi.mock("@shellhub/design-system/primitives", () => ({
  StatusDot: ({
    online,
    className,
  }: {
    online?: boolean;
    className?: string;
  }) => (
    <span
      role="img"
      aria-label={online === false ? "Offline" : "Online"}
      className={`status-dot-mock ${className ?? ""}`}
      data-testid="status-dot"
    />
  ),
}));

describe("OnlineDot", () => {
  describe("online", () => {
    it("renders aria-label Online", () => {
      render(<OnlineDot online />);
      expect(screen.getByRole("img", { name: "Online" })).toBeInTheDocument();
    });

    it("has mx-auto class on outer span", () => {
      render(<OnlineDot online />);
      const el = screen.getByRole("img", { name: "Online" });
      expect(el.className).toContain("mx-auto");
    });

    it("delegates to StatusDot", () => {
      render(<OnlineDot online />);
      expect(screen.getByTestId("status-dot")).toBeInTheDocument();
    });
  });

  describe("offline", () => {
    it("renders aria-label Offline", () => {
      render(<OnlineDot online={false} />);
      expect(screen.getByRole("img", { name: "Offline" })).toBeInTheDocument();
    });

    it("has mx-auto class on the single span", () => {
      render(<OnlineDot online={false} />);
      const el = screen.getByRole("img", { name: "Offline" });
      expect(el.className).toContain("mx-auto");
    });

    it("delegates to StatusDot", () => {
      render(<OnlineDot online={false} />);
      expect(screen.getByTestId("status-dot")).toBeInTheDocument();
    });
  });
});
