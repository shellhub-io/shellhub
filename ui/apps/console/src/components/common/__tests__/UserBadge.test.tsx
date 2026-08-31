import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import UserBadge from "../UserBadge";

describe("UserBadge", () => {
  it("shows the name over the email and a monogram", () => {
    render(<UserBadge name="John Doe" email="john@example.com" />);
    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.getByText("john@example.com")).toBeInTheDocument();
    expect(screen.getByText("JD")).toBeInTheDocument();
  });

  it("falls back to the email as the primary label when there is no name", () => {
    render(<UserBadge email="svc@example.com" />);
    expect(screen.getAllByText("svc@example.com")).toHaveLength(1);
  });

  it("short mode hides the email", () => {
    render(<UserBadge name="John Doe" email="john@example.com" short />);
    expect(screen.getByText("John Doe")).toBeInTheDocument();
    expect(screen.queryByText("john@example.com")).not.toBeInTheDocument();
  });

  it("lets a secondary node take the email's line", () => {
    render(
      <UserBadge
        name="CI runner"
        email="ci@example.com"
        secondary={<span>Service account</span>}
      />,
    );
    expect(screen.getByText("Service account")).toBeInTheDocument();
    expect(screen.queryByText("ci@example.com")).not.toBeInTheDocument();
  });
});
