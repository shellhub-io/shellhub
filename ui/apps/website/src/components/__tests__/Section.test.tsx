import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Section } from "@/components";

describe("Section", () => {
  it("spreads id and className onto the outer element only", () => {
    render(
      <Section id="my-section" className="custom-outer" data-testid="outer">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).toHaveAttribute("id", "my-section");
    expect(outer).toHaveClass("custom-outer");
  });

  it("renders inner container with containerClassName", () => {
    render(
      <Section data-testid="outer" containerClassName="custom-inner">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    // containerClassName should NOT be on the outer element
    expect(outer).not.toHaveClass("custom-inner");
    // inner container should carry the class
    const inner = outer.firstChild as HTMLElement;
    expect(inner).toHaveClass("custom-inner");
  });

  it("containerClassName='max-w-3xl' overrides max-w-7xl via cn-merge", () => {
    render(
      <Section data-testid="outer" containerClassName="max-w-3xl">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    const inner = outer.firstChild as HTMLElement;
    expect(inner).toHaveClass("max-w-3xl");
    expect(inner).not.toHaveClass("max-w-7xl");
  });

  it("container={false} removes inner wrapper", () => {
    render(
      <Section data-testid="outer" container={false}>
        <span>child</span>
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    // direct child should be the span, not an inner div
    expect(outer.firstChild).toHaveTextContent("child");
    expect(outer.querySelector("div")).toBeNull();
  });

  it("container={false} makes containerClassName inert (not applied anywhere)", () => {
    render(
      <Section
        data-testid="outer"
        container={false}
        containerClassName="should-not-appear"
      >
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer.innerHTML).not.toContain("should-not-appear");
  });

  it("container={false} makes centered inert (text-center not applied)", () => {
    render(
      <Section data-testid="outer" container={false} centered>
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).not.toHaveClass("text-center");
    expect(outer.innerHTML).not.toContain("text-center");
  });

  it("bordered={false} removes border-t class from outer element", () => {
    render(
      <Section data-testid="outer" bordered={false}>
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).not.toHaveClass("border-t");
  });

  it("bordered defaults to true and applies border-t border-border", () => {
    render(<Section data-testid="outer">content</Section>);
    const outer = screen.getByTestId("outer");
    expect(outer).toHaveClass("border-t");
    expect(outer).toHaveClass("border-border");
  });

  it("background='surface' adds bg-surface to outer element", () => {
    render(
      <Section data-testid="outer" background="surface">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).toHaveClass("bg-surface");
  });

  it("background='none' does not add bg-surface", () => {
    render(
      <Section data-testid="outer" background="none">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).not.toHaveClass("bg-surface");
  });

  it("polymorphic as='div' renders a <div> element", () => {
    render(
      <Section as="div" data-testid="outer">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer.tagName.toLowerCase()).toBe("div");
  });

  it("renders as <section> by default", () => {
    render(<Section data-testid="outer">content</Section>);
    const outer = screen.getByTestId("outer");
    expect(outer.tagName.toLowerCase()).toBe("section");
  });

  it("centered adds text-center to the inner container", () => {
    render(
      <Section data-testid="outer" centered>
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    const inner = outer.firstChild as HTMLElement;
    expect(inner).toHaveClass("text-center");
  });

  it("padding='lg' applies py-24 (default)", () => {
    render(<Section data-testid="outer">content</Section>);
    const outer = screen.getByTestId("outer");
    expect(outer).toHaveClass("py-24");
  });

  it("padding='md' applies py-12", () => {
    render(
      <Section data-testid="outer" padding="md">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).toHaveClass("py-12");
  });

  it("padding='none' applies no py- class", () => {
    render(
      <Section data-testid="outer" padding="none">
        content
      </Section>,
    );
    const outer = screen.getByTestId("outer");
    expect(outer).not.toHaveClass("py-24");
    expect(outer).not.toHaveClass("py-12");
  });
});
