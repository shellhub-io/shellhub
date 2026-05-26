import { describe, it, expect } from "vitest";
import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";
import Breadcrumb from "../Breadcrumb";

function renderBreadcrumb(ui: React.ReactNode) {
  return render(<MemoryRouter>{ui}</MemoryRouter>);
}

describe("Breadcrumb", () => {
  it("renders a navigation landmark named 'Breadcrumb'", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: "host-a" }]}
      />,
    );
    expect(
      screen.getByRole("navigation", { name: "Breadcrumb" }),
    ).toBeInTheDocument();
  });

  it("renders one list item per breadcrumb entry inside an ordered list", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[
          { label: "Announcements", to: "/admin/announcements" },
          { label: "Title", to: "/admin/announcements/1" },
          { label: "Edit" },
        ]}
      />,
    );
    const nav = screen.getByRole("navigation", { name: "Breadcrumb" });
    const list = within(nav).getByRole("list");
    expect(list.tagName).toBe("OL");
    expect(within(list).getAllByRole("listitem")).toHaveLength(3);
  });

  it("renders items with `to` as links and the final item as a non-link span", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: "host-a" }]}
      />,
    );
    const parent = screen.getByRole("link", { name: "Devices" });
    expect(parent).toHaveAttribute("href", "/devices");
    expect(
      screen.queryByRole("link", { name: "host-a" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("host-a").tagName).toBe("SPAN");
  });

  it("marks only the final crumb with aria-current='page'", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[
          { label: "A", to: "/a" },
          { label: "B", to: "/a/b" },
          { label: "C" },
        ]}
      />,
    );
    expect(screen.getByText("C")).toHaveAttribute("aria-current", "page");
    expect(screen.getByText("A")).not.toHaveAttribute("aria-current");
    expect(screen.getByText("B")).not.toHaveAttribute("aria-current");
  });

  it("renders a final item as a span even when `to` is provided", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[
          { label: "Devices", to: "/devices" },
          { label: "host-a", to: "/devices/host-a" },
        ]}
      />,
    );
    expect(
      screen.queryByRole("link", { name: "host-a" }),
    ).not.toBeInTheDocument();
    expect(screen.getByText("host-a")).toHaveAttribute("aria-current", "page");
  });

  it("renders a chevron between each item but not before the first", () => {
    const { container } = renderBreadcrumb(
      <Breadcrumb
        items={[
          { label: "A", to: "/a" },
          { label: "B", to: "/a/b" },
          { label: "C" },
        ]}
      />,
    );
    const chevrons = container.querySelectorAll("svg[aria-hidden='true']");
    expect(chevrons).toHaveLength(2);
  });

  it("forwards `title` to the rendered link or span for tooltip support", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[
          { label: "Devices", to: "/devices", title: "All devices" },
          { label: "host-a", title: "host-a.example.com" },
        ]}
      />,
    );
    expect(screen.getByRole("link", { name: "Devices" })).toHaveAttribute(
      "title",
      "All devices",
    );
    expect(screen.getByText("host-a")).toHaveAttribute(
      "title",
      "host-a.example.com",
    );
  });

  it("falls back to the string label as the title when none is provided", () => {
    renderBreadcrumb(
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: "host-a" }]}
      />,
    );
    expect(screen.getByRole("link", { name: "Devices" })).toHaveAttribute(
      "title",
      "Devices",
    );
    expect(screen.getByText("host-a")).toHaveAttribute("title", "host-a");
  });
});
