import { afterEach, describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import TimelineCard from "../TimelineCard";

afterEach(() => vi.useRealTimers());

describe("TimelineCard", () => {
  it("shows each timestamp under its label, with relative time for Last Seen", () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date("2026-06-20T17:00:00Z"));
    render(
      <TimelineCard
        createdAt="2026-01-15T10:30:00Z"
        lastSeen="2026-06-20T14:00:00Z"
        statusUpdatedAt="2026-03-01T08:00:00Z"
      />,
    );

    expect(screen.getByRole("heading", { name: "Timeline" })).toBeInTheDocument();
    expect(screen.getByRole("definition", { name: "Created" })).toHaveTextContent("Jan 15, 2026, 10:30");
    expect(screen.getByRole("definition", { name: "Last Seen" })).toHaveTextContent("3 hours ago");
    expect(screen.getByRole("definition", { name: "Last Seen" })).toHaveTextContent("Jun 20, 2026, 14:00");
    expect(screen.getByRole("definition", { name: "Status Updated" })).toHaveTextContent("Mar 1, 2026, 08:00");
  });

  it("marks each empty timestamp as absent", () => {
    render(<TimelineCard createdAt="" lastSeen="" statusUpdatedAt="" />);

    expect(screen.getByRole("definition", { name: "Created" })).toHaveTextContent(/^—$/);
    expect(screen.getByRole("definition", { name: "Last Seen" })).toHaveTextContent(/^—+$/);
    expect(screen.getByRole("definition", { name: "Status Updated" })).toHaveTextContent(/^—$/);
  });
});
