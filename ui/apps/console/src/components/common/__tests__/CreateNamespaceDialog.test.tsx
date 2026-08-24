import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import "./helpers/setup-dialog";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import { ClipboardProvider } from "../ClipboardProvider";
import CreateNamespaceDialog from "../CreateNamespaceDialog";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    createNamespace: vi.fn(),
    getNamespaceToken: vi.fn(),
    getNamespaces: vi.fn(),
  }),
);

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  sdk.getNamespaces.mockResolvedValue(paginatedResponse([]));
  sdk.createNamespace.mockResolvedValue(
    mockSdkResponse(mockNamespace({ name: "my-ns" })),
  );
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
});

function renderDialog(open: boolean, onClose = vi.fn()) {
  return {
    onClose,
    ...render(
      <ClipboardProvider>
        <CreateNamespaceDialog open={open} onClose={onClose} />
      </ClipboardProvider>,
      { wrapper: createTestWrapper() },
    ),
  };
}

describe("CreateNamespaceDialog (community)", () => {
  it("renders nothing — namespace creation is a premium feature", () => {
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

  it("Create button is disabled while mutation is pending", async () => {
    sdk.createNamespace.mockReturnValue(new Promise(() => {}));
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => {
      expect(document.querySelector("button[type='submit']")).toBeDisabled();
    });
  });

  it("shows a validation error when name is too short on submit", async () => {
    renderDialog(true);
    fireEvent.change(screen.getByPlaceholderText("my-namespace"), {
      target: { value: "ab" },
    });
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

  it("calls createNamespace with the namespace name on valid submission", async () => {
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() =>
      expect(sdk.createNamespace).toHaveBeenCalledWith(
        expect.objectContaining({
          body: { name: "my-ns" },
          throwOnError: true,
        }),
      ),
    );
  });

  it("forces lowercase on input", async () => {
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "MyNS");
    expect(screen.getByPlaceholderText("my-namespace")).toHaveValue("myns");
  });

  it("shows 'A namespace with this name already exists.' on 409 and does NOT call onClose", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 409 });
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("shows the limit/permission message on 403", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 403 });
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText(
        "You have reached the namespace limit or do not have permission.",
      ),
    ).toBeInTheDocument();
  });

  it("shows the invalid-name message on 400", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 400 });
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("The namespace name is invalid."),
    ).toBeInTheDocument();
  });

  it("shows the generic fallback message on 500", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 500 });
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText(
        "An unexpected error occurred. Please try again.",
      ),
    ).toBeInTheDocument();
  });

  it("clears the error text when the user types after a failed submission", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 409 });
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();

    await user.type(screen.getByPlaceholderText("my-namespace"), "x");
    expect(
      screen.queryByText("A namespace with this name already exists."),
    ).not.toBeInTheDocument();
  });

  it("calls onClose after successful creation", async () => {
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
  });
});
