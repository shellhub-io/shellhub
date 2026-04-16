import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ApiKey } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useApiKeys", () => ({
  useApiKeys: vi.fn(),
}));

vi.mock("@/hooks/useApiKeyMutations", () => ({
  useDeleteApiKey: vi.fn(),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("../GenerateKeyDrawer", () => ({
  default: () => null,
}));

vi.mock("../EditKeyDrawer", () => ({
  default: () => null,
}));

vi.mock("@/components/common/ConfirmDialog", () => ({
  default: ({
    open,
    onClose,
    onConfirm,
    title,
    description,
    confirmLabel = "Confirm",
    children,
  }: {
    open: boolean;
    onClose: () => void;
    onConfirm: () => Promise<void> | void;
    title: string;
    description: ReactNode;
    confirmLabel?: string;
    children?: ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog" aria-label={title}>
        <h2>{title}</h2>
        <div>{description}</div>
        {children}
        <button onClick={onClose}>Cancel</button>
        <button onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useApiKeys } from "@/hooks/useApiKeys";
import { useDeleteApiKey } from "@/hooks/useApiKeyMutations";
import ApiKeysTab from "../ApiKeysTab";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeApiKey(overrides: Partial<ApiKey> = {}): ApiKey {
  return {
    tenant_id: "tenant-abc",
    created_by: "user-xyz",
    role: "administrator",
    name: "prod-key",
    expires_in: Math.floor(Date.now() / 1000) + 3600 * 24 * 30,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

const mockMutateAsync = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(useApiKeys).mockReturnValue({
    apiKeys: [makeApiKey()],
    totalCount: 1,
    isLoading: false,
    error: null,
  });
  vi.mocked(useDeleteApiKey).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("ApiKeysTab — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    render(<ApiKeysTab />);
    await user.click(screen.getByRole("button", { name: "Delete" }));
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete api key/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    mockMutateAsync.mockRejectedValue(new Error("Key is protected"));
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("Key is protected")).toBeInTheDocument(),
    );
    expect(dialog).toBeInTheDocument();
  });

  it("shows a generic fallback message when the rejection is not an Error", async () => {
    mockMutateAsync.mockRejectedValue("boom");
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText(/failed to delete api key/i),
      ).toBeInTheDocument(),
    );
  });

  it("closes the dialog and does not show an error on successful deletion", async () => {
    mockMutateAsync.mockResolvedValue(undefined);
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: /delete api key/i }),
      ).not.toBeInTheDocument(),
    );
  });
});
