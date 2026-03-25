import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { render, screen, cleanup, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./helpers/setup-dialog";

// Mock the focus trap so it doesn't interfere with jsdom focus state
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));

vi.mock("@/env", () => ({
  getConfig: vi.fn(),
}));

vi.mock("@/hooks/useNamespaceMutations", () => ({
  useCreateNamespace: vi.fn(),
}));

// CopyButton calls navigator.clipboard — stub it to avoid jsdom errors
Object.defineProperty(navigator, "clipboard", {
  value: { writeText: vi.fn().mockResolvedValue(undefined) },
  configurable: true,
});

import { getConfig } from "@/env";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";
import { ClipboardProvider } from "../ClipboardProvider";
import CreateNamespaceDialog from "../CreateNamespaceDialog";

const mockGetConfig = vi.mocked(getConfig);
const mockUseCreateNamespace = vi.mocked(useCreateNamespace);

beforeEach(() => {
  // Default to CE (no cloud/enterprise features)
  mockGetConfig.mockReturnValue({ cloud: false, enterprise: false, version: "", onboardingUrl: "", announcements: false });
  mockUseCreateNamespace.mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
    error: null,
    reset: vi.fn(),
  } as unknown as ReturnType<typeof useCreateNamespace>);
});

afterEach(cleanup);

function renderDialog(open: boolean, onClose = vi.fn()) {
  return { onClose, ...render(<ClipboardProvider><CreateNamespaceDialog open={open} onClose={onClose} /></ClipboardProvider>) };
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
    it("dialog aria-labelledby points to the heading element", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby");
      expect(labelId).toBeTruthy();
      expect(document.getElementById(labelId!)).toHaveTextContent(
        "Create a Namespace",
      );
    });

    it("dialog aria-describedby points to the description paragraph", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const descId = dialog.getAttribute("aria-describedby");
      expect(descId).toBeTruthy();
      expect(document.getElementById(descId!)).toHaveTextContent(
        /Community Edition uses the CLI to manage namespaces/i,
      );
    });

    it("heading id matches dialog's aria-labelledby", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby")!;
      expect(
        screen.getByRole("heading", { name: "Create a Namespace" }),
      ).toHaveAttribute("id", labelId);
    });

    it("description paragraph id matches dialog's aria-describedby", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const descId = dialog.getAttribute("aria-describedby")!;
      const description = screen.getByText(
        /Community Edition uses the CLI to manage namespaces/i,
      );
      expect(description).toHaveAttribute("id", descId);
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

describe("CreateNamespaceDialog (cloud/enterprise)", () => {
  beforeEach(() => {
    mockGetConfig.mockReturnValue({ cloud: false, enterprise: true, version: "", onboardingUrl: "", announcements: false });
  });

  it("renders the creation form instead of CLI instructions", () => {
    renderDialog(true);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
    expect(screen.queryByText(/Community Edition uses the CLI/i)).not.toBeInTheDocument();
  });

  it("renders the name input and Create button", () => {
    renderDialog(true);
    expect(screen.getByPlaceholderText("my-namespace")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create" })).toBeInTheDocument();
  });

  it("Create button is disabled when name is fewer than 3 characters", () => {
    renderDialog(true);
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
  });

  it("Create button is disabled while mutation is pending", () => {
    mockUseCreateNamespace.mockReturnValue({
      mutateAsync: vi.fn(),
      isPending: true,
      error: null,
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateNamespace>);

    renderDialog(true);
    fireEvent.change(screen.getByPlaceholderText("my-namespace"), { target: { value: "my-ns" } });
    // When pending the button renders a spinner with no text; query by type
    expect(document.querySelector("button[type='submit']")).toBeDisabled();
  });

  it("shows a validation error when name is too short on submit", async () => {
    renderDialog(true);

    fireEvent.change(screen.getByPlaceholderText("my-namespace"), { target: { value: "ab" } });
    // Button is disabled for names < 3 chars; submit the form directly
    fireEvent.submit(screen.getByPlaceholderText("my-namespace").closest("form")!);

    await waitFor(() =>
      expect(screen.getByText("Name must be at least 3 characters")).toBeInTheDocument(),
    );
  });

  it("shows a validation error for names with invalid characters", async () => {
    const user = userEvent.setup();
    renderDialog(true);

    const input = screen.getByPlaceholderText("my-namespace");
    await user.type(input, "-badname");
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(
      screen.getByText(/Only lowercase letters, numbers, and hyphens/i),
    ).toBeInTheDocument();
  });

  it("calls mutateAsync with the namespace name on valid submission", async () => {
    const mutateAsync = vi.fn<() => Promise<void>>().mockResolvedValue(undefined);
    mockUseCreateNamespace.mockReturnValue({
      mutateAsync,
      isPending: false,
      error: null,
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateNamespace>);

    const user = userEvent.setup();
    renderDialog(true);

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(mutateAsync).toHaveBeenCalledWith("my-ns"));
  });

  it("forces lowercase on input", async () => {
    const user = userEvent.setup();
    renderDialog(true);

    await user.type(screen.getByPlaceholderText("my-namespace"), "MyNS");
    expect(screen.getByPlaceholderText("my-namespace")).toHaveValue("myns");
  });

  it("shows the mutation error message when creation fails", async () => {
    const mutateAsync = vi.fn<() => Promise<void>>().mockRejectedValue(new Error("name already taken"));
    mockUseCreateNamespace.mockReturnValue({
      mutateAsync,
      isPending: false,
      error: new Error("name already taken"),
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateNamespace>);

    const user = userEvent.setup();
    renderDialog(true);

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() =>
      expect(screen.getByText("name already taken")).toBeInTheDocument(),
    );
  });

  it("clears the mutation error when the user types in the input", async () => {
    const mutateAsync = vi.fn<() => Promise<void>>().mockRejectedValue(new Error("name already taken"));
    const reset = vi.fn();
    mockUseCreateNamespace.mockReturnValue({
      mutateAsync,
      isPending: false,
      error: new Error("name already taken"),
      reset,
    } as unknown as ReturnType<typeof useCreateNamespace>);

    const user = userEvent.setup();
    renderDialog(true);

    // Type to trigger onChange
    await user.type(screen.getByPlaceholderText("my-namespace"), "a");

    expect(reset).toHaveBeenCalled();
  });

  it("calls onClose after successful creation", async () => {
    const mutateAsync = vi.fn<() => Promise<void>>().mockResolvedValue(undefined);
    mockUseCreateNamespace.mockReturnValue({
      mutateAsync,
      isPending: false,
      error: null,
      reset: vi.fn(),
    } as unknown as ReturnType<typeof useCreateNamespace>);

    const user = userEvent.setup();
    const { onClose } = renderDialog(true);

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
  });
});

describe("CreateNamespaceDialog (cloud: true, enterprise: false)", () => {
  beforeEach(() => {
    mockGetConfig.mockReturnValue({ cloud: true, enterprise: false, version: "", onboardingUrl: "", announcements: false });
  });

  it("renders the creation form (cloud branch of isCloud)", () => {
    renderDialog(true);
    expect(screen.getByRole("textbox")).toBeInTheDocument();
    expect(screen.queryByText(/Community Edition uses the CLI/i)).not.toBeInTheDocument();
  });
});
