import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it } from "vitest";
import SidebarShell from "../SidebarShell";
import type { ComponentProps } from "react";

type SidebarShellProps = ComponentProps<typeof SidebarShell>;

function renderSidebarShell(overrides: Partial<SidebarShellProps> = {}) {
  const props: SidebarShellProps = {
    expanded: true,
    ariaLabel: "Test navigation",
    logoHref: "/dashboard",
    account: <span>Account</span>,
    children: <span>Navigation item</span>,
    ...overrides,
  };

  return render(
    <MemoryRouter>
      <SidebarShell {...props} />
    </MemoryRouter>,
  );
}

describe("SidebarShell", () => {
  it("names the logo link and keeps the mark out of the accessibility tree", () => {
    renderSidebarShell();

    expect(screen.getByRole("link", { name: "ShellHub" })).toBeInTheDocument();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });
});
