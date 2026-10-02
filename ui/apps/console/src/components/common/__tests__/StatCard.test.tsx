import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

import StatCard from "../StatCard";

function renderLink() {
  return render(
    <MemoryRouter>
      <StatCard
        icon={<span>icon</span>}
        title="Online Devices"
        value={42}
        action={{ label: "View all", to: "/devices" }}
      />
    </MemoryRouter>,
  );
}

function renderButton() {
  return render(
    <MemoryRouter>
      <StatCard
        icon={<span>icon</span>}
        title="Pending Devices"
        value={5}
        action={{ label: "View pending", onClick: () => undefined }}
      />
    </MemoryRouter>,
  );
}

describe("StatCard", () => {
  it("shows the title and the value", () => {
    renderLink();
    expect(screen.getByText("Online Devices")).toBeInTheDocument();
    expect(screen.getByText("42")).toBeInTheDocument();
  });

  it("a link action renders a link to the correct destination", () => {
    renderLink();
    const link = screen.getByRole("link", { name: /view all/i });
    expect(link).toHaveAttribute("href", "/devices");
  });

  it("a click action renders a button", () => {
    renderButton();
    expect(
      screen.getByRole("button", { name: /view pending/i }),
    ).toBeInTheDocument();
  });

  it("renders neither a link nor a button without an action", () => {
    render(
      <StatCard icon={<span>icon</span>} title="Active Sessions" value={3} />,
    );
    expect(screen.getByText("3")).toBeInTheDocument();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("hides the decorative icon wrapper from assistive technology", () => {
    renderLink();
    expect(screen.getByText("icon").closest("[aria-hidden]")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });
});
