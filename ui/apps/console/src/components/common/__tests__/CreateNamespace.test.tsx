import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { mockNamespace } from "@/tests/factories";
import { getConfig, defaultConfig } from "@/env";
import CreateNamespace from "../CreateNamespace";

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
  mockGetConfig.mockReturnValue({ ...defaultConfig, edition: "cloud" });
  sdk.getNamespaces.mockResolvedValue(paginatedResponse([]));
  sdk.createNamespace.mockResolvedValue(
    mockSdkResponse(mockNamespace({ name: "my-ns" })),
  );
  sdk.getNamespaceToken.mockResolvedValue(
    mockSdkResponse({ token: "jwt-token", role: "owner" }),
  );
});

function renderComponent() {
  return render(<CreateNamespace />, { wrapper: createTestWrapper() });
}

describe("CreateNamespace — CloudForm", () => {
  it("calls createNamespace with the typed name on valid submit", async () => {
    const user = userEvent.setup();
    renderComponent();
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

  it("shows 'A namespace with this name already exists.' on 409", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 409 });
    const user = userEvent.setup();
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("A namespace with this name already exists."),
    ).toBeInTheDocument();
  });

  it("shows the limit/permission message on 403", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 403 });
    const user = userEvent.setup();
    renderComponent();
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
    renderComponent();
    await user.type(screen.getByPlaceholderText("my-namespace"), "my-ns");
    await user.click(screen.getByRole("button", { name: "Create" }));
    expect(
      await screen.findByText("The namespace name is invalid."),
    ).toBeInTheDocument();
  });

  it("shows the generic fallback message on 500", async () => {
    sdk.createNamespace.mockRejectedValue({ status: 500 });
    const user = userEvent.setup();
    renderComponent();
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
    renderComponent();
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
});
