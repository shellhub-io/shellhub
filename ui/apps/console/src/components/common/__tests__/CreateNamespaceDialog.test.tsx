import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import { createTestWrapper } from "@/tests/wrapper";
import { mockNamespace, mockUserAuth } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import { ClipboardProvider } from "../ClipboardProvider";
import CreateNamespaceDialog from "../CreateNamespaceDialog";

const mockGetConfig = vi.mocked(getConfig);

beforeEach(() => {
  vi.clearAllMocks();
  mockGetConfig.mockReturnValue({ ...defaultConfig });
  server.use(
    http.get("*/api/namespaces", () => jsonWithTotal([])),
    http.post("*/api/namespaces", () =>
      HttpResponse.json(mockNamespace({ name: "my-ns" })),
    ),
    http.get("*/api/auth/token/:tenant", () =>
      HttpResponse.json(mockUserAuth({ token: "jwt-token" })),
    ),
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

  it("renders nothing when open=false", () => {
    renderDialog(false);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("labels the dialog with its heading", () => {
    renderDialog(true);
    const labelId = screen
      .getByRole("dialog")
      .getAttribute("aria-labelledby");
    expect(
      screen.getByRole("heading", { name: "Create a Namespace" }),
    ).toHaveAttribute("id", labelId);
  });

  it("links to the Administration Guide in a new tab", () => {
    renderDialog(true);
    const link = screen.getByRole("link", { name: /administration guide/i });
    expect(link).toHaveAttribute(
      "href",
      "https://docs.shellhub.io/self-hosted/administration",
    );
    expect(link).toHaveAttribute("target", "_blank");
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

  it("Create button is disabled when name is fewer than 3 characters", () => {
    renderDialog(true);
    expect(screen.getByRole("button", { name: "Create" })).toBeDisabled();
  });

  it("Create button is disabled while mutation is pending", async () => {
    server.use(http.post("*/api/namespaces", () => new Promise(() => {})));
    const user = userEvent.setup();
    renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => {
      expect(document.querySelector("button[type='submit']")).toBeDisabled();
    });
  });

  it("keeps the dialog open and shows the error when creation fails", async () => {
    server.use(
      http.post("*/api/namespaces", () =>
        HttpResponse.json({}, { status: 409 }),
      ),
    );
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);

    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));

    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
    expect(onClose).not.toHaveBeenCalled();
  });

  it("calls onClose after successful creation", async () => {
    const user = userEvent.setup();
    const { onClose } = renderDialog(true);
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    await waitFor(() => expect(onClose).toHaveBeenCalledOnce());
  });
});
