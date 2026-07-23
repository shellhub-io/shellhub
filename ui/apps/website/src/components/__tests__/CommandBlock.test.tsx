import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import { CommandBlock } from "@/components";

const COMMAND = "docker run -d -p 80:80 shellhubio/shellhub";

const writeText = vi.fn();

beforeEach(() => {
  Object.defineProperty(globalThis, "isSecureContext", {
    value: true,
    configurable: true,
  });
  writeText.mockReset();
  writeText.mockResolvedValue(undefined);
  Object.defineProperty(navigator, "clipboard", {
    value: { writeText },
    configurable: true,
  });
});

afterEach(cleanup);

describe("CommandBlock", () => {
  it("displays the command", () => {
    render(<CommandBlock command={COMMAND} />);
    expect(screen.getByText(COMMAND)).toBeInTheDocument();
  });

  it("renders a copy button", () => {
    render(<CommandBlock command={COMMAND} />);
    expect(screen.getByRole("button", { name: "Copy" })).toBeInTheDocument();
  });

  it("copies the command to the clipboard when the copy button is clicked", () => {
    render(<CommandBlock command={COMMAND} />);
    fireEvent.click(screen.getByRole("button", { name: "Copy" }));
    expect(writeText).toHaveBeenCalledWith(COMMAND);
  });

  it("forwards className to the root element", () => {
    const { container } = render(
      <CommandBlock command={COMMAND} className="my-custom-class" />,
    );
    expect(container.firstElementChild?.className).toContain("my-custom-class");
  });
});
