import { describe, it, expect, vi, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent } from "@testing-library/react";
import userEvent from "@testing-library/user-event";

// Mock the focus trap so it doesn't interfere with jsdom focus state
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

// jsdom doesn't implement showModal/close — stub them so they set/remove the
// `open` attribute, which is what React Testing Library uses to resolve the
// `dialog` role.
HTMLDialogElement.prototype.showModal = vi.fn(function (
  this: HTMLDialogElement,
) {
  this.setAttribute("open", "");
});
HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) {
  this.removeAttribute("open");
});

// CopyButton calls navigator.clipboard — stub it to avoid jsdom errors
Object.defineProperty(navigator, "clipboard", {
  value: { writeText: vi.fn().mockResolvedValue(undefined) },
  configurable: true,
});

import CreateNamespaceDialog from "../CreateNamespaceDialog";

afterEach(cleanup);

function renderDialog(open: boolean, onClose = vi.fn()) {
  return { onClose, ...render(<CreateNamespaceDialog open={open} onClose={onClose} />) };
}

describe("CreateNamespaceDialog", () => {
  describe("when open=false", () => {
    it("renders nothing", () => {
      renderDialog(false);
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  describe("when open=true", () => {
    it("renders the dialog", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    });

    it("displays the heading 'Create a Namespace'", () => {
      renderDialog(true);
      expect(
        screen.getByRole("heading", { name: "Create a Namespace" }),
      ).toBeInTheDocument();
    });
  });

  describe("CLI command", () => {
    it("displays the correct CLI command text", () => {
      renderDialog(true);
      expect(
        screen.getByText("./bin/cli namespace create <namespace> <owner>"),
      ).toBeInTheDocument();
    });
  });

  describe("naming rules", () => {
    it("displays the '3–30 characters' rule", () => {
      renderDialog(true);
      expect(screen.getByText("3–30 characters")).toBeInTheDocument();
    });

    it("displays the 'Lowercase letters, numbers, and hyphens only' rule", () => {
      renderDialog(true);
      expect(
        screen.getByText("Lowercase letters, numbers, and hyphens only"),
      ).toBeInTheDocument();
    });

    it("displays the 'Cannot begin or end with a hyphen' rule", () => {
      renderDialog(true);
      expect(
        screen.getByText("Cannot begin or end with a hyphen"),
      ).toBeInTheDocument();
    });

    it("renders all three naming rules inside the labelled list", () => {
      renderDialog(true);
      const list = screen.getByRole("list", { name: "Namespace naming rules" });
      expect(list.querySelectorAll("li")).toHaveLength(3);
    });
  });

  describe("closing the dialog", () => {
    it("calls onClose when the X button is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Close dialog" }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when the Close button in the footer is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Close" }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when the native cancel event fires (Escape key)", () => {
      const { onClose } = renderDialog(true);

      fireEvent(screen.getByRole("dialog"), new Event("cancel"));

      expect(onClose).toHaveBeenCalledOnce();
    });
  });

  describe("aria attributes", () => {
    it("dialog has aria-labelledby='create-ns-title'", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toHaveAttribute(
        "aria-labelledby",
        "create-ns-title",
      );
    });

    it("dialog has aria-describedby='create-ns-description'", () => {
      renderDialog(true);
      expect(screen.getByRole("dialog")).toHaveAttribute(
        "aria-describedby",
        "create-ns-description",
      );
    });

    it("heading has id='create-ns-title'", () => {
      renderDialog(true);
      expect(
        screen.getByRole("heading", { name: "Create a Namespace" }),
      ).toHaveAttribute("id", "create-ns-title");
    });

    it("description paragraph has id='create-ns-description'", () => {
      renderDialog(true);
      // The description text is the paragraph that introduces the CLI command
      const description = screen.getByText(
        /Community Edition uses the CLI to manage namespaces/i,
      );
      expect(description).toHaveAttribute("id", "create-ns-description");
    });
  });

  describe("documentation link", () => {
    it("renders a link to the Administration Guide", () => {
      renderDialog(true);
      const link = screen.getByRole("link", { name: /administration guide/i });
      expect(link).toHaveAttribute(
        "href",
        "https://docs.shellhub.io/self-hosted/administration",
      );
    });

    it("link opens in a new tab", () => {
      renderDialog(true);
      const link = screen.getByRole("link", { name: /administration guide/i });
      expect(link).toHaveAttribute("target", "_blank");
    });
  });
});
