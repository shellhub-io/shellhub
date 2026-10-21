import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import TimelineCard from "../TimelineCard";

describe("TimelineCard", () => {
  it("shows em-dash for empty dates", () => {
    render(
      <TimelineCard createdAt="" lastSeen="" statusUpdatedAt="" />,
    );

    const dashes = screen.getAllByText("—");
    expect(dashes.length).toBeGreaterThanOrEqual(2);
  });
});
