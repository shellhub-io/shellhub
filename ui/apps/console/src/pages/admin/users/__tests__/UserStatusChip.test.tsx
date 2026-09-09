import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import UserStatusChip from "../UserStatusChip";

describe("UserStatusChip", () => {
  it.each([
    ["confirmed", "Confirmed"],
    ["not-confirmed", "Not Confirmed"],
  ] as const)("renders '%s' status as '%s'", (status, label) => {
    render(<UserStatusChip status={status} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});
