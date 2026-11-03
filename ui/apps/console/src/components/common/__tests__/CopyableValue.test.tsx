import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ClipboardProvider } from "../ClipboardProvider";
import CopyableValue from "../CopyableValue";

function renderCopyableValue(ui: React.ReactElement) {
  return render(<ClipboardProvider>{ui}</ClipboardProvider>);
}

describe("CopyableValue", () => {
  it("names the value by its label", () => {
    renderCopyableValue(<CopyableValue label="Your API Key" value="sk-123" />);

    expect(screen.getByLabelText("Your API Key")).toHaveTextContent("sk-123");
  });

  it("keeps two values apart when both are on screen", () => {
    renderCopyableValue(
      <>
        <CopyableValue label="First" value="one" />
        <CopyableValue label="Second" value="two" />
      </>,
    );

    expect(screen.getByLabelText("First")).toHaveTextContent("one");
    expect(screen.getByLabelText("Second")).toHaveTextContent("two");
  });
});
