import { describe, it, expect } from "vitest";
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { Navbar } from "@/components/Navbar";
import { routes } from "@/routes";

function renderNavbar() {
  return render(
    <MemoryRouter>
      <Navbar />
    </MemoryRouter>,
  );
}

function nav() {
  return screen.getByTestId("desktop-nav");
}

describe("Navbar", () => {
  it.each(["Features", "Enterprise", "Pricing"])(
    "renders %s as a top-level link",
    (label) => {
      renderNavbar();
      expect(
        within(nav()).getByRole("link", { name: label }),
      ).toBeInTheDocument();
    },
  );

  it.each(["Solutions", "Resources"])(
    "renders %s as a dropdown trigger",
    (label) => {
      renderNavbar();
      expect(
        within(nav()).getByRole("button", { name: new RegExp(label, "i") }),
      ).toBeInTheDocument();
    },
  );

  it.each([
    {
      dropdown: "Solutions",
      items: [
        "IoT & Embedded",
        "Edge Computing",
        "Container Management",
        "Remote Support",
        "DevOps & CI/CD",
      ],
    },
    {
      dropdown: "Resources",
      items: ["Documentation", "GitHub", "Changelog", "Forum"],
    },
  ])("shows $dropdown items when open", async ({ dropdown, items }) => {
    renderNavbar();

    await userEvent.click(
      within(nav()).getByRole("button", { name: new RegExp(dropdown, "i") }),
    );

    for (const label of items) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
  });

  it("every internal href resolves to a known route", async () => {
    const routePaths = new Set(routes.map((r) => r.path));
    const hrefs = new Set<string>();

    for (const dropdown of ["Solutions", "Resources"] as const) {
      renderNavbar();

      await userEvent.click(
        within(nav()).getByRole("button", { name: new RegExp(dropdown, "i") }),
      );

      for (const a of screen.getAllByRole("link")) {
        const href = a.getAttribute("href");
        if (href) hrefs.add(href);
      }

      cleanup();
    }

    const deadLinks = [...hrefs]
      .filter((href) => href.startsWith("/"))
      .filter((href) => !routePaths.has(href));

    expect(deadLinks, `Dead nav links: ${deadLinks.join(", ")}`).toHaveLength(
      0,
    );
  });
});
