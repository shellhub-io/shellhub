import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import InstanceApiKeys from "../InstanceApiKeys";
import type { InstanceApiKey } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { ClipboardProvider } from "@/components/common/ClipboardProvider";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    listInstanceApiKeys: vi.fn(),
    createInstanceApiKey: vi.fn(),
    deleteInstanceApiKey: vi.fn(),
  }),
);

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

function mockInstanceApiKey(
  overrides: Partial<InstanceApiKey> = {},
): InstanceApiKey {
  return {
    name: "billing-export",
    created_by: "3dd0d1f8-8246-4519-b11a-a3dd33717f65",
    created_at: "2026-05-01T00:00:00Z",
    updated_at: "2026-05-01T00:00:00Z",
    expires_at: "2036-05-31T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  sdk.listInstanceApiKeys.mockResolvedValue(
    paginatedResponse([mockInstanceApiKey()]),
  );
  sdk.deleteInstanceApiKey.mockResolvedValue(mockSdkResponse(undefined));
});

function renderPage() {
  return render(
    <ClipboardProvider>
      <InstanceApiKeys />
    </ClipboardProvider>,
    {
      wrapper: createTestWrapper({
        initialEntries: ["/admin/instance-api-keys"],
      }),
    },
  );
}

describe("InstanceApiKeys", () => {
  it("lists the instance keys the API returns", async () => {
    renderPage();

    expect(await screen.findByText("billing-export")).toBeInTheDocument();
  });

  it("shows the plaintext key once after creating one", async () => {
    const user = userEvent.setup();
    sdk.createInstanceApiKey.mockResolvedValue(
      mockSdkResponse({
        ...mockInstanceApiKey({ name: "license-sync" }),
        id: "sh_admin_cdfd3cb0-c44e-4e54-b931-6d57713ad159",
      }),
    );

    renderPage();

    await user.click(
      await screen.findByRole("button", { name: /generate key/i }),
    );
    await user.type(await screen.findByLabelText(/name/i), "license-sync");
    await user.click(screen.getByRole("radio", { name: /30 days/i }));
    await user.click(screen.getByRole("button", { name: /^generate$/i }));

    expect(
      await screen.findByText("sh_admin_cdfd3cb0-c44e-4e54-b931-6d57713ad159"),
    ).toBeInTheDocument();
  });

  it("opens with the shortest expiry already selected", async () => {
    const user = userEvent.setup();

    renderPage();

    await user.click(
      await screen.findByRole("button", { name: /generate key/i }),
    );

    const expiry = screen.getByRole("radiogroup", { name: /expiration/i });
    expect(within(expiry).getByRole("radio", { name: /30 days/i })).toBeChecked();

    await user.type(await screen.findByLabelText(/name/i), "license-sync");

    await waitFor(() =>
      expect(screen.getByRole("button", { name: /^generate$/i })).toBeEnabled(),
    );
  });

  it("offers no never-expires option", async () => {
    const user = userEvent.setup();

    renderPage();

    await user.click(
      await screen.findByRole("button", { name: /generate key/i }),
    );

    const expiry = screen.getByRole("radiogroup", { name: /expiration/i });
    expect(within(expiry).queryByRole("radio", { name: /never/i })).toBeNull();
  });

  it("revokes a key", async () => {
    const user = userEvent.setup();

    renderPage();

    await user.click(
      await screen.findByRole("button", { name: /revoke instance api key/i }),
    );
    await user.click(screen.getByRole("button", { name: /^revoke$/i }));

    await waitFor(() => {
      expect(sdk.deleteInstanceApiKey).toHaveBeenCalledWith(
        expect.objectContaining({ path: { name: "billing-export" } }),
      );
    });
  });
});
