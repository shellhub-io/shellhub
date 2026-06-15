import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { Button } from "../primitives/Button";

// ---------------------------------------------------------------------------
// Variant classes
// ---------------------------------------------------------------------------
describe("Button — variants", () => {
  it("variant=primary renders primary classes", () => {
    render(
      <Button data-testid="btn" variant="primary">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-primary");
    expect(el.className).toContain("text-white");
  });

  it("variant=secondary renders secondary classes", () => {
    render(
      <Button data-testid="btn" variant="secondary">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-surface");
    expect(el.className).toContain("border");
    expect(el.className).toContain("border-border");
    expect(el.className).toContain("text-text-primary");
  });

  it("variant=ghost renders ghost classes", () => {
    render(
      <Button data-testid="btn" variant="ghost">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-transparent");
    expect(el.className).toContain("text-text-primary");
  });

  it("variant=destructive renders destructive classes", () => {
    render(
      <Button data-testid="btn" variant="destructive">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-accent-red");
    expect(el.className).toContain("text-white");
  });

  it("variant=dangerSoft renders soft danger classes", () => {
    render(
      <Button data-testid="btn" variant="dangerSoft">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-accent-red/10");
    expect(el.className).toContain("text-accent-red");
    expect(el.className).toContain("border");
    expect(el.className).toContain("border-accent-red/20");
  });

  it("variant=success renders success classes", () => {
    render(
      <Button data-testid="btn" variant="success">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-accent-green");
    expect(el.className).toContain("text-white");
  });

  it("variant=warning renders warning classes", () => {
    render(
      <Button data-testid="btn" variant="warning">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-accent-yellow");
    expect(el.className).toContain("text-background");
  });

  it("variant=outline renders outline classes", () => {
    render(
      <Button data-testid="btn" variant="outline">
        Click
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-surface");
    expect(el.className).toContain("border-border");
    expect(el.className).toContain("text-text-secondary");
  });
});

// ---------------------------------------------------------------------------
// Size — padding / text / radius
// ---------------------------------------------------------------------------
describe("Button — sizes", () => {
  it("size=sm has correct padding, text, radius", () => {
    render(
      <Button data-testid="btn" size="sm">
        X
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("px-3");
    expect(el.className).toContain("py-1.5");
    expect(el.className).toContain("text-xs");
    expect(el.className).toContain("rounded-md");
  });

  it("size=md (default) has correct padding, text, radius", () => {
    render(<Button data-testid="btn">X</Button>);
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("px-4");
    expect(el.className).toContain("py-2");
    expect(el.className).toContain("text-sm");
    expect(el.className).toContain("rounded-lg");
  });

  it("size=lg has correct padding, text, radius", () => {
    render(
      <Button data-testid="btn" size="lg">
        X
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("px-5");
    expect(el.className).toContain("py-2.5");
    expect(el.className).toContain("text-base");
    expect(el.className).toContain("rounded-lg");
  });
});

// ---------------------------------------------------------------------------
// Base classes — focus ring
// ---------------------------------------------------------------------------
describe("Button — base classes", () => {
  it("includes focus-visible:ring-2", () => {
    render(<Button data-testid="btn">X</Button>);
    expect(screen.getByTestId("btn").className).toContain(
      "focus-visible:ring-2",
    );
  });

  it("includes focus-visible:ring-offset-background", () => {
    render(<Button data-testid="btn">X</Button>);
    expect(screen.getByTestId("btn").className).toContain(
      "focus-visible:ring-offset-background",
    );
  });
});

// ---------------------------------------------------------------------------
// Native button type
// ---------------------------------------------------------------------------
describe("Button — native type attribute", () => {
  it("defaults to type=button", () => {
    render(<Button data-testid="btn">X</Button>);
    expect(screen.getByTestId("btn")).toHaveAttribute("type", "button");
  });

  it("type=submit passes through", () => {
    render(
      <Button data-testid="btn" type="submit">
        X
      </Button>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute("type", "submit");
  });
});

// ---------------------------------------------------------------------------
// glow=true
// ---------------------------------------------------------------------------
describe("Button — glow", () => {
  it("glow=true adds shadow and scale classes on the button", () => {
    render(
      <Button data-testid="btn" glow>
        X
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("shadow");
    expect(el.className).toContain("scale");
  });

  it("glow=true does NOT render a shimmer overlay span", () => {
    render(
      <Button data-testid="btn" glow>
        X
      </Button>,
    );
    // Shimmer overlay was removed — no translate-x span should exist
    const shimmer = screen
      .getByTestId("btn")
      .querySelector("[class*='translate-x']");
    expect(shimmer).toBeNull();
  });

  it("glow=false (default) does NOT add shadow/scale classes", () => {
    render(<Button data-testid="btn">X</Button>);
    const el = screen.getByTestId("btn");
    expect(el.className).not.toContain("shadow-primary");
    expect(el.className).not.toContain("scale-[1.02]");
  });
});

// ---------------------------------------------------------------------------
// loading
// ---------------------------------------------------------------------------
describe("Button — loading", () => {
  it("loading=true renders a Spinner inside the button", () => {
    render(
      <Button data-testid="btn" loading>
        Save
      </Button>,
    );
    // Spinner renders a <span> with animate-spin
    const spinner = screen
      .getByTestId("btn")
      .querySelector("[class*='animate-spin']");
    expect(spinner).not.toBeNull();
  });

  it("loading=true sets disabled attribute", () => {
    render(
      <Button data-testid="btn" loading>
        Save
      </Button>,
    );
    expect(screen.getByTestId("btn")).toBeDisabled();
  });

  it("loading=true sets aria-busy=true", () => {
    render(
      <Button data-testid="btn" loading>
        Save
      </Button>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute("aria-busy", "true");
  });

  it("loading=true with icon: spinner replaces the icon", () => {
    const Icon = () => <svg data-testid="icon" />;
    render(
      <Button data-testid="btn" loading icon={<Icon />}>
        Save
      </Button>,
    );
    // The icon should not be rendered when loading
    expect(screen.queryByTestId("icon")).toBeNull();
    // But spinner should be there
    const spinner = screen
      .getByTestId("btn")
      .querySelector("[class*='animate-spin']");
    expect(spinner).not.toBeNull();
  });
});

// ---------------------------------------------------------------------------
// as=a (polymorphic)
// ---------------------------------------------------------------------------
describe("Button — as=a (anchor)", () => {
  it("renders an <a> element", () => {
    render(
      <Button as="a" href="/foo" data-testid="btn">
        Link
      </Button>,
    );
    expect(screen.getByTestId("btn").tagName).toBe("A");
  });

  it("forwards href, target, rel", () => {
    render(
      <Button
        as="a"
        href="/foo"
        target="_blank"
        rel="noopener noreferrer"
        data-testid="btn"
      >
        Link
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el).toHaveAttribute("href", "/foo");
    expect(el).toHaveAttribute("target", "_blank");
    expect(el).toHaveAttribute("rel", "noopener noreferrer");
  });

  it("does not emit type attribute for anchor", () => {
    render(
      <Button as="a" href="/foo" data-testid="btn">
        Link
      </Button>,
    );
    expect(screen.getByTestId("btn")).not.toHaveAttribute("type");
  });

  it("does not emit disabled attribute for anchor", () => {
    render(
      <Button as="a" href="/foo" data-testid="btn" loading>
        Link
      </Button>,
    );
    expect(screen.getByTestId("btn")).not.toHaveAttribute("disabled");
  });

  it("when loading, anchor gets aria-disabled=true and pointer-events-none", () => {
    render(
      <Button as="a" href="/foo" data-testid="btn" loading>
        Link
      </Button>,
    );
    const el = screen.getByTestId("btn");
    expect(el).toHaveAttribute("aria-disabled", "true");
    expect(el.className).toContain("pointer-events-none");
  });
});

// ---------------------------------------------------------------------------
// className override via cn
// ---------------------------------------------------------------------------
describe("Button — className override", () => {
  it("merges className via cn (custom class is present)", () => {
    render(
      <Button data-testid="btn" className="my-custom-class">
        X
      </Button>,
    );
    expect(screen.getByTestId("btn").className).toContain("my-custom-class");
  });

  it("className is applied last so it can override built-in classes", () => {
    render(
      <Button data-testid="btn" className="rounded-full">
        X
      </Button>,
    );
    const cls = screen.getByTestId("btn").className;
    // rounded-full should win over default rounded-lg (tw-merge resolves border-radius group)
    expect(cls).toContain("rounded-full");
    expect(cls).not.toContain("rounded-lg");
  });
});

// ---------------------------------------------------------------------------
// icon / iconRight placement
// ---------------------------------------------------------------------------
describe("Button — icon / iconRight", () => {
  it("icon prop renders a leading icon before children", () => {
    render(
      <Button data-testid="btn" icon={<span data-testid="leading" />}>
        label
      </Button>,
    );
    const btn = screen.getByTestId("btn");
    const leading = screen.getByTestId("leading");
    // leading icon must appear before the text in DOM order
    expect(btn.contains(leading)).toBe(true);
    const children = Array.from(btn.childNodes);
    const leadingIdx = children.findIndex((n) => n.contains(leading));
    const textIdx = children.findIndex((n) => n.textContent?.includes("label"));
    expect(leadingIdx).toBeLessThan(textIdx);
  });

  it("iconRight prop renders a trailing icon after children", () => {
    render(
      <Button data-testid="btn" iconRight={<span data-testid="trailing" />}>
        label
      </Button>,
    );
    const btn = screen.getByTestId("btn");
    const trailing = screen.getByTestId("trailing");
    expect(btn.contains(trailing)).toBe(true);
    const children = Array.from(btn.childNodes);
    const trailingIdx = children.findIndex((n) => n.contains(trailing));
    const textIdx = children.findIndex((n) => n.textContent?.includes("label"));
    expect(trailingIdx).toBeGreaterThan(textIdx);
  });
});

// ---------------------------------------------------------------------------
// fullWidth
// ---------------------------------------------------------------------------
describe("Button — fullWidth", () => {
  it("fullWidth=true adds w-full class", () => {
    render(
      <Button data-testid="btn" fullWidth>
        X
      </Button>,
    );
    expect(screen.getByTestId("btn").className).toContain("w-full");
  });

  it("fullWidth=false (default) does NOT add w-full", () => {
    render(<Button data-testid="btn">X</Button>);
    expect(screen.getByTestId("btn").className).not.toContain("w-full");
  });
});
