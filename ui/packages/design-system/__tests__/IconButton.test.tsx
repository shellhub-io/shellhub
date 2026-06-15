import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { IconButton } from "../primitives/IconButton";

// ---------------------------------------------------------------------------
// aria-label forwarding
// ---------------------------------------------------------------------------
describe("IconButton — aria-label", () => {
  it("forwards aria-label to the underlying element", () => {
    render(
      <IconButton aria-label="Close dialog" data-testid="btn">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute(
      "aria-label",
      "Close dialog",
    );
  });
});

// ---------------------------------------------------------------------------
// Base classes
// ---------------------------------------------------------------------------
describe("IconButton — base classes", () => {
  it("includes inline-flex items-center justify-center shrink-0", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("inline-flex");
    expect(el.className).toContain("items-center");
    expect(el.className).toContain("justify-center");
    expect(el.className).toContain("shrink-0");
  });

  it("includes focus-visible:ring-2", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    expect(screen.getByTestId("btn").className).toContain(
      "focus-visible:ring-2",
    );
  });

  it("includes disabled:opacity-50", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    expect(screen.getByTestId("btn").className).toContain(
      "disabled:opacity-50",
    );
  });
});

// ---------------------------------------------------------------------------
// Size map
// ---------------------------------------------------------------------------
describe("IconButton — sizes", () => {
  it("size=sm adds p-1 rounded", () => {
    render(
      <IconButton data-testid="btn" size="sm">
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("p-1");
    expect(el.className).toContain("rounded");
  });

  it("size=sm does NOT add rounded-md or rounded-lg (just rounded)", () => {
    render(
      <IconButton data-testid="btn" size="sm">
        X
      </IconButton>,
    );
    const cls = screen.getByTestId("btn").className;
    // Should have `rounded` but not `rounded-md` or `rounded-lg` from the size token
    const hasRounded = /\brounded\b/.test(cls);
    const hasRoundedMd = /\brounded-md\b/.test(cls);
    const hasRoundedLg = /\brounded-lg\b/.test(cls);
    expect(hasRounded).toBe(true);
    expect(hasRoundedMd).toBe(false);
    expect(hasRoundedLg).toBe(false);
  });

  it("size=md (default) adds p-1.5 rounded-md", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("p-1.5");
    expect(el.className).toContain("rounded-md");
  });

  it("size=lg adds w-8 h-8 rounded-lg", () => {
    render(
      <IconButton data-testid="btn" size="lg">
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("w-8");
    expect(el.className).toContain("h-8");
    expect(el.className).toContain("rounded-lg");
  });
});

// ---------------------------------------------------------------------------
// Variant map
// ---------------------------------------------------------------------------
describe("IconButton — variants", () => {
  it("variant=ghost (default) adds ghost classes", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-transparent");
    expect(el.className).toContain("text-text-primary");
  });

  it("variant=primary adds primary classes", () => {
    render(
      <IconButton data-testid="btn" variant="primary">
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("bg-primary");
    expect(el.className).toContain("text-white");
  });

  it("variant=danger adds danger classes", () => {
    render(
      <IconButton data-testid="btn" variant="danger">
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    expect(el.className).toContain("text-accent-red");
  });
});

// ---------------------------------------------------------------------------
// Polymorphic — as=a forwards href
// ---------------------------------------------------------------------------
describe("IconButton — as=a (anchor)", () => {
  it("renders an <a> element", () => {
    render(
      <IconButton as="a" href="/foo" data-testid="btn">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn").tagName).toBe("A");
  });

  it("forwards href", () => {
    render(
      <IconButton as="a" href="/foo" data-testid="btn">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute("href", "/foo");
  });

  it("does not emit type attribute for anchor", () => {
    render(
      <IconButton as="a" href="/foo" data-testid="btn">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).not.toHaveAttribute("type");
  });

  it("does not emit disabled attribute for anchor", () => {
    render(
      <IconButton as="a" href="/foo" data-testid="btn" loading>
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).not.toHaveAttribute("disabled");
  });

  it("when loading, anchor gets aria-disabled=true and pointer-events-none", () => {
    render(
      <IconButton as="a" href="/foo" data-testid="btn" loading>
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    expect(el).toHaveAttribute("aria-disabled", "true");
    expect(el.className).toContain("pointer-events-none");
  });
});

// ---------------------------------------------------------------------------
// Polymorphic — as={Link} forwards to prop
// ---------------------------------------------------------------------------
describe("IconButton — as={Link} (React Router Link-like)", () => {
  it("renders a component passed as `as` and forwards to prop", () => {
    // Simulate a React Router Link component
    const FakeLink = ({
      to,
      children,
      ...rest
    }: {
      to: string;
      children?: React.ReactNode;
      [key: string]: unknown;
    }) => (
      <a href={to as string} {...rest}>
        {children}
      </a>
    );

    render(
      <IconButton as={FakeLink} to="/dashboard" data-testid="btn">
        X
      </IconButton>,
    );
    const el = screen.getByTestId("btn");
    // Should render as an anchor (from FakeLink) and forward the `to` as href
    expect(el.tagName).toBe("A");
    expect(el).toHaveAttribute("href", "/dashboard");
  });
});

// ---------------------------------------------------------------------------
// Native button type
// ---------------------------------------------------------------------------
describe("IconButton — native type attribute", () => {
  it("defaults to type=button", () => {
    render(<IconButton data-testid="btn">X</IconButton>);
    expect(screen.getByTestId("btn")).toHaveAttribute("type", "button");
  });

  it("type=submit passes through", () => {
    render(
      <IconButton data-testid="btn" type="submit">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute("type", "submit");
  });
});

// ---------------------------------------------------------------------------
// disabled handling
// ---------------------------------------------------------------------------
describe("IconButton — disabled", () => {
  it("disabled prop makes the button disabled", () => {
    render(
      <IconButton data-testid="btn" disabled>
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toBeDisabled();
  });
});

// ---------------------------------------------------------------------------
// disabled + loading interaction
// ---------------------------------------------------------------------------
describe("IconButton — disabled prop interaction", () => {
  it("loading=true with disabled=false still produces a disabled button", () => {
    render(
      <IconButton data-testid="btn" loading disabled={false} aria-label="x">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toBeDisabled();
  });
});

// ---------------------------------------------------------------------------
// loading
// ---------------------------------------------------------------------------
describe("IconButton — loading", () => {
  it("loading=true renders a Spinner inside the button", () => {
    render(
      <IconButton data-testid="btn" loading>
        X
      </IconButton>,
    );
    // Spinner renders a <span> with animate-spin
    const spinner = screen
      .getByTestId("btn")
      .querySelector("[class*='animate-spin']");
    expect(spinner).not.toBeNull();
  });

  it("loading=true sets disabled attribute on native button", () => {
    render(
      <IconButton data-testid="btn" loading>
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toBeDisabled();
  });

  it("loading=true sets aria-busy=true", () => {
    render(
      <IconButton data-testid="btn" loading>
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn")).toHaveAttribute("aria-busy", "true");
  });
});

// ---------------------------------------------------------------------------
// className override
// ---------------------------------------------------------------------------
describe("IconButton — className override", () => {
  it("merges custom className", () => {
    render(
      <IconButton data-testid="btn" className="my-custom-class">
        X
      </IconButton>,
    );
    expect(screen.getByTestId("btn").className).toContain("my-custom-class");
  });
});

// ---------------------------------------------------------------------------
// Export re-check from primitives/index
// ---------------------------------------------------------------------------
describe("IconButton — exported from primitives/index", () => {
  it("is exported from the primitives barrel", async () => {
    const mod = await import("../primitives/index");
    expect(mod.IconButton).toBeDefined();
  });
});
