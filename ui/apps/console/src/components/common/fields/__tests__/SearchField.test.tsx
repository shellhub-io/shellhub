import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import SearchField from "@/components/common/fields/SearchField";

type Overrides = Partial<React.ComponentProps<typeof SearchField>>;

function renderSearchField(overrides: Overrides = {}) {
  return render(
    <SearchField
      value=""
      onChange={() => {}}
      placeholder="Search..."
      aria-label="Search"
      {...overrides}
    />,
  );
}

describe("SearchField", () => {
  it("exposes aria-label as the accessible name via a visually hidden <label>", () => {
    renderSearchField({ "aria-label": "Search users by username" });

    expect(
      screen.getByRole("searchbox", { name: "Search users by username" }),
    ).toBeInTheDocument();
  });

  it("calls onChange with the new value as the user types", async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();

    renderSearchField({ onChange });

    await user.type(screen.getByRole("searchbox"), "abc");
    expect(onChange).toHaveBeenCalledTimes(3);
    expect(onChange).toHaveBeenNthCalledWith(1, "a");
    expect(onChange).toHaveBeenNthCalledWith(2, "b");
    expect(onChange).toHaveBeenNthCalledWith(3, "c");
  });

  it("renders the magnifying glass icon as decorative (aria-hidden)", () => {
    const { container } = renderSearchField();

    expect(container.querySelector("svg")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
  });

  it("generates unique ids when two are rendered together", () => {
    render(
      <>
        <SearchField
          value=""
          onChange={() => {}}
          placeholder="a"
          aria-label="First"
        />
        <SearchField
          value=""
          onChange={() => {}}
          placeholder="b"
          aria-label="Second"
        />
      </>,
    );

    const [first, second] = screen.getAllByRole("searchbox");
    expect(first.id).not.toBe(second.id);
    expect(first.id).toBeTruthy();
    expect(second.id).toBeTruthy();
  });
});
