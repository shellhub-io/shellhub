import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import DeviceStatusChip from "../DeviceStatusChip";

describe("DeviceStatusChip", () => {
  it('renders "Accepted" label for accepted status', () => {
    render(<DeviceStatusChip status="accepted" />);
    expect(screen.getByText("Accepted")).toBeInTheDocument();
  });

  it('renders "Pending" label for pending status', () => {
    render(<DeviceStatusChip status="pending" />);
    expect(screen.getByText("Pending")).toBeInTheDocument();
  });

  it('renders "Rejected" label for rejected status', () => {
    render(<DeviceStatusChip status="rejected" />);
    expect(screen.getByText("Rejected")).toBeInTheDocument();
  });

  it('renders "Removed" label for removed status', () => {
    render(<DeviceStatusChip status="removed" />);
    expect(screen.getByText("Removed")).toBeInTheDocument();
  });

  it('renders "Unused" label for unused status', () => {
    render(<DeviceStatusChip status="unused" />);
    expect(screen.getByText("Unused")).toBeInTheDocument();
  });
});
