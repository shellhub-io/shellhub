import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import PageLoader from "@/components/common/PageLoader";

describe("PageLoader", () => {
  it("renders a single status live region announced by the spinner's aria-label", () => {
    render(<PageLoader label="Loading users" />);

    expect(
      screen.getByRole("status", { name: "Loading users" }),
    ).toBeInTheDocument();
    expect(screen.getAllByRole("status")).toHaveLength(1);
    expect(screen.queryByText("Loading users")).toBeNull();
  });

  it("does not put role=status on the wrapper in the default form", () => {
    const { container } = render(<PageLoader label="Loading users" />);
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.tagName).toBe("DIV");
    expect(wrapper).not.toHaveAttribute("role");
  });

  it("when showLabel is true, the wrapper is the live region and the spinner is decorative", () => {
    render(<PageLoader label="Loading settings..." showLabel />);

    const region = screen.getByRole("status", { name: "Loading settings..." });
    expect(region.tagName).toBe("DIV");
    expect(screen.getAllByRole("status")).toHaveLength(1);
    const spinner = region.querySelector(".animate-spin") as HTMLElement;
    expect(spinner).toHaveAttribute("aria-hidden", "true");
    expect(spinner).not.toHaveAttribute("role");
  });
});
