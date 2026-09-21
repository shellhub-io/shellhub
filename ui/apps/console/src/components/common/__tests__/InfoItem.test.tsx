import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { ClipboardProvider } from "../ClipboardProvider";
import InfoItem from "../InfoItem";

function renderInfoItem(ui: React.ReactElement) {
  return render(<ClipboardProvider>{ui}</ClipboardProvider>);
}

describe("InfoItem", () => {
  describe("value mode", () => {
    it("renders label in a dt and value in a dd", () => {
      renderInfoItem(<InfoItem label="UID" value="abc-123" />);

      expect(screen.getByRole("term")).toHaveTextContent("UID");
      expect(screen.getByRole("definition")).toHaveTextContent("abc-123");
    });

    it("renders a CopyButton when copyable is true and value is truthy", () => {
      renderInfoItem(<InfoItem label="UID" value="abc-123" copyable />);

      expect(screen.getByRole("button", { name: /copy/i })).toBeInTheDocument();
    });

    it("does not render a CopyButton when value is empty", () => {
      renderInfoItem(<InfoItem label="UID" value="" copyable />);

      expect(
        screen.queryByRole("button", { name: /copy/i }),
      ).not.toBeInTheDocument();
    });

    it("truncates the display text but passes full value to CopyButton", () => {
      renderInfoItem(
        <InfoItem label="UID" value="abcdefgh-1234" copyable truncate={8} />,
      );

      expect(screen.getByText("abcdefgh")).toBeInTheDocument();
      expect(screen.queryByText("abcdefgh-1234")).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: /copy/i })).toBeInTheDocument();
    });

    it("renders an em-dash when no value is provided", () => {
      renderInfoItem(<InfoItem label="MAC" />);

      expect(screen.getByRole("definition")).toHaveTextContent("—");
    });
  });

  describe("children mode", () => {
    it("renders children inside dd", () => {
      renderInfoItem(
        <InfoItem label="Status">
          <span data-testid="custom">Active</span>
        </InfoItem>,
      );

      expect(screen.getByRole("term")).toHaveTextContent("Status");
      expect(screen.getByTestId("custom")).toBeInTheDocument();
    });

    it("ignores value/mono/copyable/truncate when children are provided", () => {
      renderInfoItem(
        <InfoItem label="Link" value="ignored" mono copyable truncate={4}>
          <a href="/test">Custom link</a>
        </InfoItem>,
      );

      expect(screen.queryByText("ignored")).not.toBeInTheDocument();
      expect(screen.queryByText("igno")).not.toBeInTheDocument();
      expect(
        screen.queryByRole("button", { name: /copy/i }),
      ).not.toBeInTheDocument();
      expect(
        screen.getByRole("link", { name: "Custom link" }),
      ).toBeInTheDocument();
    });
  });
});
