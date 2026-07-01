import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import { describe, expect, it, vi } from "vitest";
import SidebarShell from "../SidebarShell";
import type { ComponentProps } from "react";

type SidebarShellProps = ComponentProps<typeof SidebarShell>;

function renderSidebarShell(overrides: Partial<SidebarShellProps> = {}) {
  const props: SidebarShellProps = {
    expanded: true,
    pinned: false,
    onToggle: vi.fn(),
    ariaLabel: "Test navigation",
    footerLabel: "Console",
    logoHref: "/dashboard",
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
  it("renders a pin control by default", () => {
    renderSidebarShell();

    const button = screen.getByRole("button", { name: "Pin sidebar" });
    expect(button).toHaveAttribute("title", "Pin sidebar open");
  });

  it("labels the pin control as unpin when pinned", () => {
    renderSidebarShell({ pinned: true });

    const button = screen.getByRole("button", { name: "Unpin sidebar" });
    expect(button).toHaveAttribute("title", "Unpin sidebar");
  });

  it("allows the toggle label to match a non-pin action", () => {
    renderSidebarShell({ toggleLabel: "Close sidebar" });

    const button = screen.getByRole("button", { name: "Close sidebar" });
    expect(button).toHaveAttribute("title", "Close sidebar");
  });

  it("names the logo link and hides both marks when expanded", () => {
    renderSidebarShell({ expanded: true });

    // The link carries the accessible name; both marks stay in the DOM (opacity
    // crossfade) but are aria-hidden, so "ShellHub" is announced exactly once.
    expect(screen.getByRole("link", { name: "ShellHub" })).toBeInTheDocument();
    // The wordmark has role="img" — aria-hidden keeps it out of the a11y tree.
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByTestId("sidebar-cloud-icon")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });

  it("names the logo link and hides both marks when collapsed", () => {
    renderSidebarShell({ expanded: false });

    expect(screen.getByRole("link", { name: "ShellHub" })).toBeInTheDocument();
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(screen.getByTestId("sidebar-cloud-icon")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });
});
