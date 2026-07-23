import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { SiteHeader } from "@/components/SiteHeader";

function renderSiteHeader() {
  return render(
    <MemoryRouter>
      <SiteHeader />
    </MemoryRouter>,
  );
}

describe("SiteHeader", () => {
  it("opening one dropdown closes the other", async () => {
    renderSiteHeader();
    const nav = screen.getByTestId("desktop-nav");
    const solutions = within(nav).getByRole("button", {
      name: /solutions/i,
    });
    const resources = within(nav).getByRole("button", {
      name: /resources/i,
    });

    await userEvent.click(solutions);
    expect(solutions).toHaveAttribute("aria-expanded", "true");

    await userEvent.click(resources);
    expect(resources).toHaveAttribute("aria-expanded", "true");
    expect(solutions).toHaveAttribute("aria-expanded", "false");
  });

  it("Escape closes an open dropdown", async () => {
    renderSiteHeader();
    const nav = screen.getByTestId("desktop-nav");
    const solutions = within(nav).getByRole("button", {
      name: /solutions/i,
    });

    await userEvent.click(solutions);
    expect(solutions).toHaveAttribute("aria-expanded", "true");

    await userEvent.keyboard("{Escape}");
    expect(solutions).toHaveAttribute("aria-expanded", "false");
  });

  it("hamburger opens and closes the mobile nav", async () => {
    renderSiteHeader();
    const toggle = screen.getByTestId("mobile-nav-toggle");

    expect(screen.queryByTestId("mobile-nav")).not.toBeInTheDocument();

    await userEvent.click(toggle);
    expect(screen.getByTestId("mobile-nav")).toBeInTheDocument();

    await userEvent.click(toggle);
    expect(screen.queryByTestId("mobile-nav")).not.toBeInTheDocument();
  });

  it("mobile accordion expands and collapses", async () => {
    renderSiteHeader();
    await userEvent.click(screen.getByTestId("mobile-nav-toggle"));

    const mobileNav = screen.getByTestId("mobile-nav");
    const solutionsBtn = within(mobileNav).getByRole("button", {
      name: /^solutions$/i,
    });

    expect(
      within(mobileNav).queryByText("IoT & Embedded"),
    ).not.toBeInTheDocument();

    await userEvent.click(solutionsBtn);
    expect(within(mobileNav).getByText("IoT & Embedded")).toBeInTheDocument();

    await userEvent.click(solutionsBtn);
    expect(
      within(mobileNav).queryByText("IoT & Embedded"),
    ).not.toBeInTheDocument();
  });
});
