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
        linkLabel="View all"
        linkTo="/devices"
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
        linkLabel="View pending"
        onClick={() => undefined}
      />
    </MemoryRouter>,
  );
}

describe("StatCard", () => {
  it("title is visible", () => {
    renderLink();
    expect(screen.getByText("Online Devices")).toBeInTheDocument();
  });

  it("value is visible", () => {
    renderLink();
    expect(screen.getByText("42")).toBeInTheDocument();
  });

  it("link-variant renders a link to the correct destination", () => {
    renderLink();
    const link = screen.getByRole("link", { name: /view all/i });
    expect(link).toHaveAttribute("href", "/devices");
  });

  it("button-variant renders a button", () => {
    renderButton();
    expect(
      screen.getByRole("button", { name: /view pending/i }),
    ).toBeInTheDocument();
  });

  it("hides the decorative icon wrapper from assistive technology", () => {
    renderLink();
    // closest() is robust to any intermediate wrapper the icon slot might add.
    expect(screen.getByText("icon").closest("[aria-hidden]")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });
});
