import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import UserStatusChip from "../UserStatusChip";

describe("UserStatusChip", () => {
  it("renders 'Confirmed' for confirmed status", () => {
    render(<UserStatusChip status="confirmed" />);
    expect(screen.getByText("Confirmed")).toBeInTheDocument();
  });

  it("renders 'Not Confirmed' for not-confirmed status", () => {
    render(<UserStatusChip status="not-confirmed" />);
    expect(screen.getByText("Not Confirmed")).toBeInTheDocument();
  });
});
