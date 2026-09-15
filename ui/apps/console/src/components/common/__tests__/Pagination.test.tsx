import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import Pagination from "../Pagination";

describe("Pagination", () => {
  it("gives Prev and Next an explicit type='button'", () => {
    render(
      <Pagination
        page={1}
        totalPages={3}
        totalCount={30}
        onPageChange={() => {}}
      />,
    );
    screen.getAllByRole("button").forEach((btn) => {
      expect(btn).toHaveAttribute("type", "button");
    });
  });

  it.each([
    [0, 0],
    [1, undefined],
    [1, 0],
  ])(
    "renders nothing for totalPages=%s and totalCount=%s",
    (totalPages, totalCount) => {
      const { container } = render(
        <Pagination
          page={1}
          totalPages={totalPages}
          totalCount={totalCount}
          onPageChange={() => {}}
        />,
      );
      expect(container.firstChild).toBeNull();
    },
  );

  it.each([
    [1, 5, "5 items"],
    [1, 1, "1 item"],
    [0, 7, "7 items"],
  ])(
    "totalPages=%s with totalCount=%s shows '%s' and no navigation",
    (totalPages, totalCount, label) => {
      render(
        <Pagination
          page={1}
          totalPages={totalPages}
          totalCount={totalCount}
          onPageChange={() => {}}
        />,
      );
      expect(screen.getByText(label)).toBeInTheDocument();
      expect(screen.queryAllByRole("button")).toHaveLength(0);
    },
  );

  it("falls back to the page indicator as the left label when totalCount is omitted", () => {
    render(<Pagination page={1} totalPages={3} onPageChange={() => {}} />);
    expect(screen.queryAllByRole("button")).toHaveLength(2);
    expect(screen.getByText("1 / 3")).toBeInTheDocument();
    expect(screen.getByText("Page 1 of 3")).toBeInTheDocument();
    expect(screen.queryByText(/item/i)).not.toBeInTheDocument();
  });

  it("shows the count alongside the navigation when there are multiple pages", () => {
    render(
      <Pagination
        page={2}
        totalPages={3}
        totalCount={30}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByText("30 items")).toBeInTheDocument();
    expect(screen.queryAllByRole("button")).toHaveLength(2);
    expect(screen.getByText("2 / 3")).toBeInTheDocument();
  });

  it.each([
    [1, "Previous page", "Next page"],
    [3, "Next page", "Previous page"],
  ])("on page %s of 3, %s is disabled and %s is not", (page, off, on) => {
    render(<Pagination page={page} totalPages={3} onPageChange={() => {}} />);
    expect(screen.getByRole("button", { name: off })).toBeDisabled();
    expect(screen.getByRole("button", { name: on })).not.toBeDisabled();
  });

  it.each([
    ["Previous page", 1],
    ["Next page", 3],
  ])("%s calls onPageChange with %i", async (name, expected) => {
    const onPageChange = vi.fn();
    render(
      <Pagination
        page={2}
        totalPages={3}
        totalCount={30}
        onPageChange={onPageChange}
      />,
    );
    await userEvent.click(screen.getByRole("button", { name }));
    expect(onPageChange).toHaveBeenCalledOnce();
    expect(onPageChange).toHaveBeenCalledWith(expected);
  });

  it("wraps the controls in a labelled navigation landmark", () => {
    render(
      <Pagination
        page={1}
        totalPages={3}
        totalCount={30}
        onPageChange={() => {}}
      />,
    );
    expect(
      screen.getByRole("navigation", { name: "Pagination" }),
    ).toBeInTheDocument();
  });

  it("marks the page indicator with aria-current", () => {
    render(
      <Pagination
        page={2}
        totalPages={3}
        totalCount={30}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByText("2 / 3")).toHaveAttribute("aria-current", "page");
  });
});
