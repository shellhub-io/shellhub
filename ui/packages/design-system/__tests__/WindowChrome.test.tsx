import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { WindowChrome } from "../primitives/WindowChrome";

// ---------------------------------------------------------------------------
// Terminal variant
// ---------------------------------------------------------------------------
describe("WindowChrome — terminal variant", () => {
  it("shows title text", () => {
    render(<WindowChrome variant="terminal" title="My Terminal" />);
    expect(screen.getByText("My Terminal")).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// Browser variant
// ---------------------------------------------------------------------------
describe("WindowChrome — browser variant", () => {
  it("shows the namespace url path", () => {
    render(<WindowChrome variant="browser" path="/devices/rpi-gateway" />);
    expect(
      screen.getByText("shellhub.io/devices/rpi-gateway"),
    ).toBeInTheDocument();
  });

  it("renders an svg lock icon", () => {
    const { container } = render(
      <WindowChrome variant="browser" path="/devices/rpi-gateway" />,
    );
    const svg = container.querySelector("svg");
    expect(svg).not.toBeNull();
  });
});

// ---------------------------------------------------------------------------
// Title bar slot
// ---------------------------------------------------------------------------
describe("WindowChrome — titleBarSlot", () => {
  it("renders titleBarSlot content", () => {
    render(
      <WindowChrome variant="terminal" titleBarSlot={<button>Copy</button>} />,
    );
    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// Children render inside body
// ---------------------------------------------------------------------------
describe("WindowChrome — children", () => {
  it("renders children inside the body", () => {
    render(
      <WindowChrome variant="terminal">
        <span data-testid="body-child">content</span>
      </WindowChrome>,
    );
    expect(screen.getByTestId("body-child")).toBeInTheDocument();
  });
});

// ---------------------------------------------------------------------------
// Caller-provided classes are forwarded
// ---------------------------------------------------------------------------
describe("WindowChrome — class forwarding", () => {
  it("applies bodyClassName to the body div", () => {
    const { container } = render(
      <WindowChrome variant="terminal" bodyClassName="custom-body">
        body
      </WindowChrome>,
    );
    expect(container.querySelector(".custom-body")).not.toBeNull();
  });

  it("applies className to the root element", () => {
    const { container } = render(
      <WindowChrome variant="terminal" className="my-root-class" />,
    );
    expect(container.firstElementChild!.className).toContain("my-root-class");
  });
});

// ---------------------------------------------------------------------------
// Export re-check from primitives/index
// ---------------------------------------------------------------------------
describe("WindowChrome — exported from primitives/index", () => {
  it("is exported as a callable component from the primitives barrel", async () => {
    const mod = await import("../primitives/index");
    expect(mod.WindowChrome).toBeDefined();
    expect(typeof mod.WindowChrome).toBe("function");
  });
});
