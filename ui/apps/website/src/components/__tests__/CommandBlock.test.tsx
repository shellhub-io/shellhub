import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, fireEvent, cleanup } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { CommandBlock } from "@/components";

const COMMAND =
  "git clone https://github.com/shellhub-io/shellhub.git && cd shellhub && make keygen && make start";

const MULTI_LINE = [
  "git clone https://github.com/shellhub-io/shellhub.git",
  "cd shellhub",
  "make start",
];

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

  it("shows each line of a multi-line command on its own", () => {
    render(<CommandBlock command={MULTI_LINE.join("\n")} />);

    for (const line of MULTI_LINE)
      expect(screen.getByText(line)).toBeInTheDocument();
  });

  it("copies every line of a multi-line command", async () => {
    render(<CommandBlock command={MULTI_LINE.join("\n")} />);

    await userEvent.click(screen.getByRole("button", { name: "Copy" }));

    expect(writeText).toHaveBeenCalledWith(MULTI_LINE.join("\n"));
  });

  it("forwards className to the root element", () => {
    const { container } = render(
      <CommandBlock command={COMMAND} className="my-custom-class" />,
    );
    expect(container.firstElementChild?.className).toContain("my-custom-class");
  });
});
