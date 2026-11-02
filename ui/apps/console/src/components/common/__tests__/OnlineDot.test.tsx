import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import OnlineDot from "../OnlineDot";

describe("OnlineDot", () => {
  it("updates the accessible status when the device goes offline", () => {
    const { rerender } = render(<OnlineDot online />);
    expect(screen.getByRole("img", { name: "Online" })).toBeInTheDocument();

    rerender(<OnlineDot online={false} />);

    expect(screen.getByRole("img", { name: "Offline" })).toBeInTheDocument();
    expect(screen.queryByRole("img", { name: "Online" })).not.toBeInTheDocument();
  });
});
