import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import DeviceStatusChip from "../DeviceStatusChip";

describe("DeviceStatusChip", () => {
  it.each([
    ["accepted", "Accepted"],
    ["pending", "Pending"],
    ["rejected", "Rejected"],
    ["removed", "Removed"],
    ["unused", "Unused"],
  ] as const)("renders '%s' status as '%s'", (status, label) => {
    render(<DeviceStatusChip status={status} />);
    expect(screen.getByText(label)).toBeInTheDocument();
  });
});
