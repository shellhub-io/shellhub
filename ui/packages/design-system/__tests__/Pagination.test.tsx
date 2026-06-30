import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Pagination } from "../components/Pagination";

describe("Pagination", () => {
  it("Prev and Next buttons have explicit type='button'", () => {
    render(
      <Pagination page={1} totalPages={3} totalCount={30} onPageChange={() => {}} />,
    );
    const buttons = screen.getAllByRole("button");
    buttons.forEach((btn) => {
      expect(btn).toHaveAttribute("type", "button");
    });
  });

  it("shows count for single-page non-empty list and no navigation buttons", () => {
    const { container } = render(
      <Pagination page={1} totalPages={1} totalCount={5} onPageChange={() => {}} />,
    );
    expect(container.firstChild).not.toBeNull();
    expect(screen.getByText("5 items")).toBeInTheDocument();
    expect(screen.queryAllByRole("button")).toHaveLength(0);
  });

  it("uses singular label when totalCount is 1", () => {
    const { container } = render(
      <Pagination page={1} totalPages={1} totalCount={1} onPageChange={() => {}} />,
    );
    expect(container.firstChild).not.toBeNull();
    expect(screen.getByText("1 item")).toBeInTheDocument();
  });

  it("renders null for an empty list (totalCount 0, totalPages 0)", () => {
    // Empty lists defer to each page's own empty-state, so Pagination renders nothing
    // rather than a lonely "0 items" beside it.
    const { container } = render(
      <Pagination page={1} totalPages={0} totalCount={0} onPageChange={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("shows count but no navigation when totalPages is 0 yet totalCount is provided and non-zero", () => {
    const { container } = render(
      <Pagination page={1} totalPages={0} totalCount={7} onPageChange={() => {}} />,
    );
    expect(container.firstChild).not.toBeNull();
    expect(screen.getByText("7 items")).toBeInTheDocument();
    expect(screen.queryAllByRole("button")).toHaveLength(0);
  });

  it("renders null when totalCount is omitted and totalPages is 1", () => {
    const { container } = render(
      <Pagination page={1} totalPages={1} onPageChange={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  it("renders Prev/Next navigation when totalPages > 1 and totalCount is omitted", () => {
    render(
      <Pagination page={1} totalPages={3} onPageChange={() => {}} />,
    );
    expect(screen.queryAllByRole("button")).toHaveLength(2);
    expect(screen.getByText("1 / 3")).toBeInTheDocument();
    // No item-count label when totalCount is not provided; Page X of Y fallback renders instead
    expect(screen.queryByText(/item/i)).not.toBeInTheDocument();
    expect(screen.getByText("Page 1 of 3")).toBeInTheDocument();
  });

  it("Prev is disabled on the first page when totalCount is omitted", () => {
    render(
      <Pagination page={1} totalPages={3} onPageChange={() => {}} />,
    );
    expect(screen.getByText("Prev")).toBeDisabled();
    expect(screen.getByText("Next")).not.toBeDisabled();
  });

  it("Next is disabled on the last page when totalCount is omitted", () => {
    render(
      <Pagination page={3} totalPages={3} onPageChange={() => {}} />,
    );
    expect(screen.getByText("Next")).toBeDisabled();
    expect(screen.getByText("Prev")).not.toBeDisabled();
  });

  it("shows count, two navigation buttons, and page indicator with multiple pages", () => {
    render(
      <Pagination page={2} totalPages={3} totalCount={30} onPageChange={() => {}} />,
    );
    expect(screen.getByText("30 items")).toBeInTheDocument();
    expect(screen.queryAllByRole("button")).toHaveLength(2);
    expect(screen.getByText("2 / 3")).toBeInTheDocument();
  });

  it("Prev button calls onPageChange with page - 1", async () => {
    const onPageChange = vi.fn();
    render(
      <Pagination page={2} totalPages={3} totalCount={30} onPageChange={onPageChange} />,
    );
    await userEvent.click(screen.getByText("Prev"));
    expect(onPageChange).toHaveBeenCalledOnce();
    expect(onPageChange).toHaveBeenCalledWith(1);
  });

  it("Next button calls onPageChange with page + 1", async () => {
    const onPageChange = vi.fn();
    render(
      <Pagination page={2} totalPages={3} totalCount={30} onPageChange={onPageChange} />,
    );
    await userEvent.click(screen.getByText("Next"));
    expect(onPageChange).toHaveBeenCalledOnce();
    expect(onPageChange).toHaveBeenCalledWith(3);
  });

  it("renders null when totalCount is 0 and totalPages is 1", () => {
    // A single empty page has nothing to navigate and no positive count to show,
    // so it defers to the page's own empty-state.
    const { container } = render(
      <Pagination page={1} totalPages={1} totalCount={0} onPageChange={() => {}} />,
    );
    expect(container.firstChild).toBeNull();
  });

  describe("accessibility", () => {
    it("wraps the controls in a labelled navigation landmark", () => {
      render(
        <Pagination page={1} totalPages={3} totalCount={30} onPageChange={() => {}} />,
      );
      expect(
        screen.getByRole("navigation", { name: "Pagination" }),
      ).toBeInTheDocument();
    });

    it("exposes the Prev/Next buttons via aria-label, not just text", () => {
      render(
        <Pagination page={2} totalPages={3} totalCount={30} onPageChange={() => {}} />,
      );
      expect(
        screen.getByRole("button", { name: "Previous page" }),
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Next page" }),
      ).toBeInTheDocument();
    });

    it("marks the page indicator with aria-current", () => {
      render(
        <Pagination page={2} totalPages={3} totalCount={30} onPageChange={() => {}} />,
      );
      expect(screen.getByText("2 / 3")).toHaveAttribute("aria-current", "page");
    });
  });
});
