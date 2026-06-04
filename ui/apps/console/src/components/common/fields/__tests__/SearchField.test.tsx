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
  it("renders the placeholder and current value", () => {
    renderSearchField({
      value: "ubuntu",
      placeholder: "Search devices...",
      "aria-label": "Search devices",
    });

    const input = screen.getByRole("searchbox", { name: "Search devices" });
    expect(input).toHaveAttribute("placeholder", "Search devices...");
    expect(input).toHaveValue("ubuntu");
  });

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

  it("constrains width with max-w-sm by default", () => {
    const { container } = renderSearchField();

    expect((container.firstChild as HTMLElement).className).toContain(
      "max-w-sm",
    );
  });

  it("omits the contained width when full is set", () => {
    const { container } = renderSearchField({ full: true });

    expect((container.firstChild as HTMLElement).className).not.toContain(
      "max-w-sm",
    );
  });

  it("appends custom className to the wrapper", () => {
    const { container } = renderSearchField({ className: "ml-auto mb-5" });

    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).toContain("ml-auto");
    expect(wrapper.className).toContain("mb-5");
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

  it("uses the provided id to bind the label and input", () => {
    renderSearchField({ id: "my-search", "aria-label": "My search" });

    const input = screen.getByRole("searchbox", { name: "My search" });
    expect(input).toHaveAttribute("id", "my-search");
  });
});
