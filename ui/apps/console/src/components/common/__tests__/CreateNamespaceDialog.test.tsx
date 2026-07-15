import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import {
  render,
  screen,
  cleanup,
  fireEvent,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./helpers/setup-dialog";

// Mock the focus trap so it doesn't interfere with jsdom focus state
vi.mock("@/hooks/useFocusTrap", () => ({
  useFocusTrap: vi.fn(),
}));
vi.mock("@/hooks/useNamespaceMutations", () => ({
  useCreateNamespace: vi.fn(),
}));

import { getConfig, defaultConfig } from "@/env";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";
import { ClipboardProvider } from "../ClipboardProvider";
import CreateNamespaceDialog from "../CreateNamespaceDialog";

const mockGetConfig = vi.mocked(getConfig);
const mockUseCreateNamespace = vi.mocked(useCreateNamespace);

beforeEach(() => {
  // Default to CE (no cloud/enterprise features)
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  mockUseCreateNamespace.mockReturnValue({
    mutateAsync: vi.fn(),
    isPending: false,
    error: null,
    reset: vi.fn(),
  } as unknown as ReturnType<typeof useCreateNamespace>);
});

afterEach(cleanup);

function renderDialog(open: boolean, onClose = vi.fn()) {
  return {
    onClose,
    ...render(
      <ClipboardProvider>
        <CreateNamespaceDialog open={open} onClose={onClose} />
      </ClipboardProvider>,
    ),
  };
}

describe("CreateNamespaceDialog (community)", () => {
  it("renders nothing — namespace creation is a premium feature", () => {
    // Default config is community; the selector shows the upsell instead of this dialog.
    renderDialog(true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});

describe("CreateNamespaceDialog (cloud/enterprise)", () => {
  beforeEach(() => {
    mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "enterprise" });
  });

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

  describe("closing the dialog", () => {
    it("calls onClose when the X button is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Close dialog" }));

      expect(onClose).toHaveBeenCalledOnce();
    });

    it("calls onClose when the Cancel button in the footer is clicked", async () => {
      const user = userEvent.setup();
      const { onClose } = renderDialog(true);

      await user.click(screen.getByRole("button", { name: "Cancel" }));

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

    it("heading id matches dialog's aria-labelledby", () => {
      renderDialog(true);
      const dialog = screen.getByRole("dialog");
      const labelId = dialog.getAttribute("aria-labelledby")!;
      expect(
        screen.getByRole("heading", { name: "Create a Namespace" }),
      ).toHaveAttribute("id", labelId);
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
    fireEvent.change(screen.getByPlaceholderText("my-namespace"), {
      target: { value: "my-ns" },
    });
    // When pending the button renders a spinner with no text; query by type
    expect(document.querySelector("button[type='submit']")).toBeDisabled();
  });

  it("shows a validation error when name is too short on submit", async () => {
    renderDialog(true);

    fireEvent.change(screen.getByPlaceholderText("my-namespace"), {
      target: { value: "ab" },
    });
    // Button is disabled for names < 3 chars; submit the form directly
    fireEvent.submit(
      screen.getByPlaceholderText("my-namespace").closest("form")!,
    );

    await waitFor(() =>
      expect(
        screen.getByText("Name must be at least 3 characters"),
      ).toBeInTheDocument(),
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
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockResolvedValue(undefined);
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

  it("shows 'A namespace with this name already exists.' on 409 and does NOT call onClose", async () => {
    const sdkError = { status: 409 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
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

    await waitFor(() =>
      expect(
        screen.getByText("A namespace with this name already exists."),
      ).toBeInTheDocument(),
    );
    expect(onClose).not.toHaveBeenCalled();
  });

  it("shows the limit/permission message on 403", async () => {
    const sdkError = { status: 403 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
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

    await waitFor(() =>
      expect(
        screen.getByText(
          "You have reached the namespace limit or do not have permission.",
        ),
      ).toBeInTheDocument(),
    );
  });

  it("shows the invalid-name message on 400", async () => {
    const sdkError = { status: 400 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
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

    await waitFor(() =>
      expect(
        screen.getByText("The namespace name is invalid."),
      ).toBeInTheDocument(),
    );
  });

  it("shows the generic fallback message on 500", async () => {
    const sdkError = { status: 500 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
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

    await waitFor(() =>
      expect(
        screen.getByText("An unexpected error occurred. Please try again."),
      ).toBeInTheDocument(),
    );
  });

  it("clears the error text from the DOM when the user types after a failed submission", async () => {
    const sdkError = { status: 409 };
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockRejectedValue(sdkError);
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

    await waitFor(() =>
      expect(
        screen.getByText("A namespace with this name already exists."),
      ).toBeInTheDocument(),
    );

    // Now type more — error must disappear
    await user.type(screen.getByPlaceholderText("my-namespace"), "x");

    expect(
      screen.queryByText("A namespace with this name already exists."),
    ).not.toBeInTheDocument();
  });

  it("calls onClose after successful creation", async () => {
    const mutateAsync = vi
      .fn<() => Promise<void>>()
      .mockResolvedValue(undefined);
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
