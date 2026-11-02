import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Callout } from "../primitives/Callout";

describe("Callout — semantic roles", () => {
  it("error variant uses role=alert and aria-live=assertive", () => {
    render(<Callout variant="error">Error</Callout>);
    const el = screen.getByRole("alert");
    expect(el).toHaveAttribute("aria-live", "assertive");
  });

  it("warning variant uses role=alert and aria-live=assertive", () => {
    render(<Callout variant="warning">Warning</Callout>);
    const el = screen.getByRole("alert");
    expect(el).toHaveAttribute("aria-live", "assertive");
  });

  it("success variant uses role=status and aria-live=polite", () => {
    render(<Callout variant="success">Success</Callout>);
    const el = screen.getByRole("status");
    expect(el).toHaveAttribute("aria-live", "polite");
  });

  it("info variant uses role=status and aria-live=polite", () => {
    render(<Callout variant="info">Info</Callout>);
    const el = screen.getByRole("status");
    expect(el).toHaveAttribute("aria-live", "polite");
  });
});

describe("Callout — content rendering", () => {
  it("renders plain string children", () => {
    render(<Callout variant="error">Something went wrong</Callout>);
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders ReactNode children with nested elements", () => {
    render(
      <Callout variant="error">
        <span>
          Error: <strong>field required</strong>
        </span>
      </Callout>,
    );
    expect(screen.getByText(/field required/)).toBeInTheDocument();
  });
});

describe("Callout — dismiss behavior", () => {
  it("does not render a dismiss button when onDismiss is not provided", () => {
    render(<Callout variant="error">Error</Callout>);
    expect(
      screen.queryByRole("button", { name: /dismiss/i }),
    ).not.toBeInTheDocument();
  });

  it("calls onDismiss when the dismiss button is clicked", async () => {
    const user = userEvent.setup();
    const onDismiss = vi.fn();
    render(
      <Callout variant="error" onDismiss={onDismiss}>
        Error
      </Callout>,
    );
    await user.click(screen.getByRole("button", { name: "Dismiss" }));
    expect(onDismiss).toHaveBeenCalledOnce();
  });

  it("dismiss button has type=button to avoid form submission", () => {
    render(
      <Callout variant="error" onDismiss={() => {}}>
        Error
      </Callout>,
    );
    const btn = screen.getByRole("button", { name: "Dismiss" });
    expect(btn).toHaveAttribute("type", "button");
  });
});

describe("Callout — className pass-through", () => {
  it("appends extra className to the root element", () => {
    const { container } = render(
      <Callout variant="error" className="mb-4">
        Error
      </Callout>,
    );
    expect((container.firstChild as HTMLElement).className).toContain("mb-4");
  });

  it("preserves built-in classes when className is provided", () => {
    const { container } = render(
      <Callout variant="error" className="mb-4">
        Error
      </Callout>,
    );
    const cls = (container.firstChild as HTMLElement).className;
    expect(cls).toContain("rounded-md");
  });
});

describe("Callout — semantic typeface", () => {
  it.each(["error", "success", "warning", "info"] as const)(
    "applies font-sans and not font-mono on the %s root",
    (variant) => {
      const { container } = render(<Callout variant={variant}>Body</Callout>);
      const cls = (container.firstChild as HTMLElement).className;
      expect(cls).toContain("font-sans");
      expect(cls).not.toContain("font-mono");
    },
  );
});

describe("Callout — feature variant", () => {
  it("renders role=note", () => {
    render(<Callout variant="feature">Body</Callout>);
    expect(screen.getByRole("note")).toBeInTheDocument();
  });

  it("renders the CTA link with the given href and label when action is provided", () => {
    render(
      <Callout
        variant="feature"
        action={{ href: "/x", label: "See all editions" }}
      >
        Body
      </Callout>,
    );
    const link = screen.getByRole("link", { name: "See all editions" });
    expect(link).toHaveAttribute("href", "/x");
  });

  it("renders no CTA link when action is omitted", () => {
    render(<Callout variant="feature">Body</Callout>);
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
  });

  it("does not render a dismiss button even when onDismiss is provided", () => {
    render(
      <Callout variant="feature" onDismiss={() => {}}>
        Body
      </Callout>,
    );
    expect(
      screen.queryByRole("button", { name: /dismiss/i }),
    ).not.toBeInTheDocument();
  });

  it("renders <strong> children", () => {
    render(
      <Callout variant="feature">
        <>
          Included with <strong>Cloud</strong>
        </>
      </Callout>,
    );
    expect(screen.getByText("Cloud").tagName).toBe("STRONG");
  });

  it("applies font-sans and not font-mono on the feature root", () => {
    const { container } = render(<Callout variant="feature">Body</Callout>);
    const cls = (container.firstChild as HTMLElement).className;
    expect(cls).toContain("font-sans");
    expect(cls).not.toContain("font-mono");
  });
});
