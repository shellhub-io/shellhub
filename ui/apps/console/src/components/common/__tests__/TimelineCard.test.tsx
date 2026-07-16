import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import TimelineCard from "../TimelineCard";

describe("TimelineCard", () => {
  const defaultProps = {
    createdAt: "2026-01-15T10:30:00Z",
    lastSeen: "2026-06-20T14:00:00Z",
    statusUpdatedAt: "2026-03-01T08:00:00Z",
  };

  it("renders the Timeline heading", () => {
    render(<TimelineCard {...defaultProps} />);

    expect(
      screen.getByRole("heading", { name: /timeline/i }),
    ).toBeInTheDocument();
  });

  it("renders the Created date", () => {
    render(<TimelineCard {...defaultProps} />);

    const terms = screen.getAllByRole("term");
    expect(terms.some((t) => t.textContent === "Created")).toBe(true);
  });

  it("renders Last Seen with both relative and absolute display", () => {
    render(<TimelineCard {...defaultProps} />);

    const terms = screen.getAllByRole("term");
    expect(terms.some((t) => t.textContent === "Last Seen")).toBe(true);

    const definitions = screen.getAllByRole("definition");
    const lastSeenDd = definitions[1];
    const spans = lastSeenDd.querySelectorAll("span");
    expect(spans.length).toBe(2);
  });

  it("renders Status Updated date", () => {
    render(<TimelineCard {...defaultProps} />);

    const terms = screen.getAllByRole("term");
    expect(terms.some((t) => t.textContent === "Status Updated")).toBe(true);
  });

  it("shows em-dash for empty dates", () => {
    render(
      <TimelineCard createdAt="" lastSeen="" statusUpdatedAt="" />,
    );

    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBeGreaterThanOrEqual(2);
  });
});
