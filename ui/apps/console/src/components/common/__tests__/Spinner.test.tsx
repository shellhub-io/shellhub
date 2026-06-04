import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import Spinner, {
  type SpinnerSize,
  type SpinnerTone,
} from "@/components/common/Spinner";

describe("Spinner", () => {
  it("renders with default md size and onSurface tone", () => {
    const { container } = render(<Spinner />);
    const el = container.firstChild as HTMLElement;
    expect(el.className).toContain("w-4 h-4");
    expect(el.className).toContain("border-primary/30 border-t-primary");
    expect(el.className).toContain("border-2");
    expect(el.className).toContain("rounded-full");
    expect(el.className).toContain("animate-spin");
  });

  it("is decorative by default (aria-hidden=true, no role)", () => {
    const { container } = render(<Spinner />);
    const el = container.firstChild as HTMLElement;
    expect(el).toHaveAttribute("aria-hidden", "true");
    expect(el).not.toHaveAttribute("role");
    expect(el).not.toHaveAttribute("aria-label");
  });

  it("becomes a live status region when aria-label is provided", () => {
    render(<Spinner aria-label="Loading users" />);
    const el = screen.getByRole("status");
    expect(el).toHaveAttribute("aria-label", "Loading users");
    expect(el).not.toHaveAttribute("aria-hidden");
  });

  it.each<[SpinnerSize, string]>([
    ["xs", "w-3 h-3"],
    ["sm", "w-3.5 h-3.5"],
    ["md", "w-4 h-4"],
    ["lg", "w-5 h-5"],
    ["xl", "w-6 h-6"],
    ["2xl", "w-10 h-10"],
  ])("emits the correct classes for size=%s", (size, expected) => {
    const { container } = render(<Spinner size={size} />);
    expect((container.firstChild as HTMLElement).className).toContain(expected);
  });

  it.each<[SpinnerTone, string]>([
    ["onPrimary", "border-white/30 border-t-white"],
    ["onSurface", "border-primary/30 border-t-primary"],
    ["subtle", "border-text-muted/30 border-t-text-muted"],
  ])("emits the correct classes for tone=%s", (tone, expected) => {
    const { container } = render(<Spinner tone={tone} />);
    expect((container.firstChild as HTMLElement).className).toContain(expected);
  });

  it("appends extra className after built-in classes", () => {
    const { container } = render(<Spinner className="mb-4" />);
    const cls = (container.firstChild as HTMLElement).className;
    expect(cls).toContain("mb-4");
    expect(cls.indexOf("mb-4")).toBeGreaterThan(cls.indexOf("animate-spin"));
  });
});
