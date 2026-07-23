import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import { SectionHeader } from "@/components";

vi.mock("@shellhub/design-system/components", () => ({
  Reveal: ({
    children,
    className,
  }: {
    children: React.ReactNode;
    className?: string;
  }) => (
    <div data-testid="reveal-wrapper" className={className}>
      {children}
    </div>
  ),
}));

describe("SectionHeader", () => {
  it("reveal=true wraps content in Reveal component", () => {
    render(<SectionHeader title="Hello" reveal={true} />);
    expect(screen.getByTestId("reveal-wrapper")).toBeInTheDocument();
  });

  it("reveal=false renders a plain div without Reveal", () => {
    render(<SectionHeader title="Hello" reveal={false} />);
    expect(screen.queryByTestId("reveal-wrapper")).toBeNull();
    // Should still render a wrapper div
    const h2 = screen.getByRole("heading", { level: 2 });
    expect(h2).toBeInTheDocument();
  });

  it("title renders as an <h2>", () => {
    render(<SectionHeader title="My Title" reveal={false} />);
    const h2 = screen.getByRole("heading", { level: 2 });
    expect(h2).toHaveTextContent("My Title");
  });

  it("title accepts ReactNode with inner <span>", () => {
    render(
      <SectionHeader
        title={
          <>
            <span data-testid="inner-span">Inner</span> text
          </>
        }
        reveal={false}
      />,
    );
    expect(screen.getByTestId("inner-span")).toBeInTheDocument();
    const h2 = screen.getByRole("heading", { level: 2 });
    expect(h2).toContainElement(screen.getByTestId("inner-span"));
  });

  it("size='sub' emits exact class text-[clamp(1.75rem,4vw,2.5rem)]", () => {
    render(<SectionHeader title="Sub" size="sub" reveal={false} />);
    const h2 = screen.getByRole("heading", { level: 2 });
    expect(h2.className).toContain("text-[clamp(1.75rem,4vw,2.5rem)]");
  });

  it("align='left' does not apply text-center or mx-auto to the wrapper", () => {
    const { container } = render(
      <SectionHeader title="Left" align="left" reveal={false} />,
    );
    const wrapper = container.firstChild as HTMLElement;
    expect(wrapper.className).not.toContain("text-center");
    expect(wrapper.className).not.toContain("mx-auto");
  });

  it("eyebrow has the complete required class set and does not use text-sm", () => {
    render(<SectionHeader title="T" eyebrow="Feature" reveal={false} />);
    const eyebrow = screen.getByText("Feature");
    expect(eyebrow).toHaveClass("text-2xs");
    expect(eyebrow).toHaveClass("font-mono");
    expect(eyebrow).toHaveClass("font-semibold");
    expect(eyebrow).toHaveClass("uppercase");
    expect(eyebrow).toHaveClass("tracking-label");
    expect(eyebrow).toHaveClass("mb-3");
    expect(eyebrow).toHaveClass("text-primary");
    expect(eyebrow).not.toHaveClass("text-sm");
  });

  it("h2 has tracking-[-0.03em]", () => {
    render(<SectionHeader title="My Title" reveal={false} />);
    const h2 = screen.getByRole("heading", { level: 2 });
    expect(h2.className).toContain("tracking-[-0.03em]");
  });

  it("className='mb-10' lands on the outer wrapper", () => {
    render(<SectionHeader title="T" className="mb-10" reveal={false} />);
    const wrapper =
      screen.queryByTestId("reveal-wrapper") ??
      document.querySelector(".mb-10");
    expect(wrapper).not.toBeNull();
    expect((wrapper as HTMLElement).className).toContain("mb-10");
  });

  it("subtitle has default text-sm max-w-lg mx-auto leading-relaxed and mt-4", () => {
    render(<SectionHeader title="T" subtitle="Sub text" reveal={false} />);
    const subtitle = screen.getByText("Sub text");
    expect(subtitle).toHaveClass("text-sm");
    expect(subtitle).toHaveClass("max-w-lg");
    expect(subtitle).toHaveClass("mx-auto");
    expect(subtitle).toHaveClass("leading-relaxed");
    expect(subtitle).toHaveClass("mt-4");
  });

  it("outer wrapper has mb-14 by default (reveal=true)", () => {
    render(<SectionHeader title="T" />);
    const wrapper = screen.getByTestId("reveal-wrapper");
    expect(wrapper.className).toContain("mb-14");
  });

  it("className='mb-10' overrides default mb-14 (reveal=false)", () => {
    render(<SectionHeader title="T" className="mb-10" reveal={false} />);
    const wrapper = document.querySelector(".mb-10") as HTMLElement;
    expect(wrapper).not.toBeNull();
    expect(wrapper.className).toContain("mb-10");
    expect(wrapper.className).not.toContain("mb-14");
  });

  it("align='left' subtitle drops max-w-lg and mx-auto", () => {
    render(
      <SectionHeader
        title="T"
        subtitle="Sub text"
        align="left"
        reveal={false}
      />,
    );
    const subtitle = screen.getByText("Sub text");
    expect(subtitle).toHaveClass("text-sm");
    expect(subtitle).toHaveClass("leading-relaxed");
    expect(subtitle).toHaveClass("mt-4");
    expect(subtitle).not.toHaveClass("max-w-lg");
    expect(subtitle).not.toHaveClass("mx-auto");
  });

  it("subtitleClassName='max-w-xl' overrides default max-w-lg", () => {
    render(
      <SectionHeader
        title="T"
        subtitle="Sub text"
        subtitleClassName="max-w-xl"
        reveal={false}
      />,
    );
    const subtitle = screen.getByText("Sub text");
    expect(subtitle).toHaveClass("max-w-xl");
    expect(subtitle).not.toHaveClass("max-w-lg");
  });

  describe("variant='cta'", () => {
    it("renders NO Reveal wrapper by default", () => {
      render(<SectionHeader variant="cta" title="CTA Title" />);
      expect(screen.queryByTestId("reveal-wrapper")).toBeNull();
    });

    it("h2 gets the cta clamp class", () => {
      render(<SectionHeader variant="cta" title="CTA Title" />);
      const h2 = screen.getByRole("heading", { level: 2 });
      expect(h2.className).toContain("text-[clamp(1.5rem,3vw,2.25rem)]");
    });

    it("content wrapper has mb-0 and NOT mb-14", () => {
      const { container } = render(
        <SectionHeader variant="cta" title="CTA Title" />,
      );
      const wrapper = container.firstChild as HTMLElement;
      expect(wrapper.className).toContain("mb-0");
      expect(wrapper.className).not.toContain("mb-14");
    });

    it("centered subtitle gets max-w-md, mx-auto, mb-8 and NOT max-w-lg", () => {
      render(
        <SectionHeader variant="cta" title="CTA Title" subtitle="CTA Sub" />,
      );
      const subtitle = screen.getByText("CTA Sub");
      expect(subtitle).toHaveClass("max-w-md");
      expect(subtitle).toHaveClass("mx-auto");
      expect(subtitle).toHaveClass("mb-8");
      expect(subtitle).not.toHaveClass("max-w-lg");
    });

    it("explicit reveal=true still renders the Reveal wrapper", () => {
      render(<SectionHeader variant="cta" reveal={true} title="CTA Title" />);
      expect(screen.getByTestId("reveal-wrapper")).toBeInTheDocument();
    });
  });
});
