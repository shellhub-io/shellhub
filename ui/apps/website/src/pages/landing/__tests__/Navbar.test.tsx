import { describe, it, expect, vi } from "vitest";
import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { Navbar } from "@/pages/landing/Navbar";

function renderNavbar(mobileMenu: boolean, setMobileMenu = vi.fn()) {
  return render(
    <MemoryRouter>
      <Navbar
        navSolid={false}
        mobileMenu={mobileMenu}
        setMobileMenu={setMobileMenu}
      />
    </MemoryRouter>,
  );
}

describe("Navbar interactions", () => {
  it("hamburger toggle calls setMobileMenu(true) when mobileMenu is false", async () => {
    const setMobileMenu = vi.fn();
    renderNavbar(false, setMobileMenu);

    const toggle = screen.getByTestId("mobile-nav-toggle");
    await userEvent.click(toggle);

    expect(setMobileMenu).toHaveBeenCalledWith(true);
  });

  it("desktop dropdown flips aria-expanded on Product trigger", async () => {
    renderNavbar(false);

    const desktopNav = screen.getByTestId("desktop-nav");
    const productBtn = within(desktopNav).getByRole("button", {
      name: /product/i,
    });

    expect(productBtn).toHaveAttribute("aria-expanded", "false");

    await userEvent.click(productBtn);
    expect(productBtn).toHaveAttribute("aria-expanded", "true");

    await userEvent.click(productBtn);
    expect(productBtn).toHaveAttribute("aria-expanded", "false");
  });

  it("mobile accordion expands and collapses SSH Gateway link on Product trigger", async () => {
    renderNavbar(true);

    const mobileNav = screen.getByTestId("mobile-nav");
    const productBtn = within(mobileNav).getByRole("button", {
      name: /^product$/i,
    });

    // Before clicking: SSH Gateway link should not be visible
    expect(
      within(mobileNav).queryByText("SSH Gateway"),
    ).not.toBeInTheDocument();

    await userEvent.click(productBtn);
    expect(within(mobileNav).getByText("SSH Gateway")).toBeInTheDocument();

    await userEvent.click(productBtn);
    expect(
      within(mobileNav).queryByText("SSH Gateway"),
    ).not.toBeInTheDocument();
  });
});
