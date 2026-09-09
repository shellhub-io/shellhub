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
    />
  ),
}));

describe("OnlineDot", () => {
  it.each([
    [true, "Online"],
    [false, "Offline"],
  ])("online=%s renders aria-label '%s'", (online, label) => {
    render(<OnlineDot online={online} />);
    expect(screen.getByRole("img", { name: label })).toBeInTheDocument();
  });
});
