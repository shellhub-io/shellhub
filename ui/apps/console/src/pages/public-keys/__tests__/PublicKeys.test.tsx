import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { PublicKeyResponse as PublicKey } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/usePublicKeys", () => ({
  usePublicKeys: vi.fn(),
}));

vi.mock("@/hooks/usePublicKeyMutations", () => ({
  useDeletePublicKey: vi.fn(),
}));

vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

vi.mock("../KeyDrawer", () => ({
  default: () => null,
}));

// CopyButton reads from ClipboardProvider which we don't wire up in tests.
vi.mock("@/components/common/CopyButton", () => ({
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
        <button type="button" onClick={onClose}>Cancel</button>
        <button type="button" onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

// Return the debounced value immediately so tests don't need fake timers.
vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { usePublicKeys } from "@/hooks/usePublicKeys";
import { useDeletePublicKey } from "@/hooks/usePublicKeyMutations";
import PublicKeys from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeKey(overrides: Partial<PublicKey> = {}): PublicKey {
  return {
    data: "ssh-rsa AAAA...",
    fingerprint: "aa:bb:cc:dd",
    created_at: "2024-01-01T00:00:00Z",
    tenant_id: "tenant-abc",
    name: "my-key",
    filter: { hostname: ".*", tags: [] },
    username: ".*",
    ...overrides,
  };
}

const mockMutateAsync = vi.fn();

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <PublicKeys />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  vi.mocked(usePublicKeys).mockReturnValue({
    publicKeys: [makeKey()],
    totalCount: 1,
    isLoading: false,
    error: null,
  });
  vi.mocked(useDeletePublicKey).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("PublicKeys — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderPage();
    await user.click(screen.getByRole("button", { name: /^delete/i }));
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete public key/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    mockMutateAsync.mockRejectedValue(new Error("Fingerprint in use"));
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText("Fingerprint in use"),
      ).toBeInTheDocument(),
    );
    expect(dialog).toBeInTheDocument();
  });

  it("shows a generic fallback message when the rejection is not an Error", async () => {
    mockMutateAsync.mockRejectedValue({ status: 500 });
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText(/failed to delete public key/i),
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
        screen.queryByRole("dialog", { name: /delete public key/i }),
      ).not.toBeInTheDocument(),
    );
  });
});

// ── URL hydration (usePaginatedListState adoption) ────────────────────────────

describe("PublicKeys — URL hydration", () => {
  it("calls usePublicKeys with page hydrated from ?page=3", () => {
    renderPage(["/?page=3"]);
    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 3 }),
    );
  });

  it("calls usePublicKeys with page=1 when URL has no page param", () => {
    renderPage(["/"]);
    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 }),
    );
  });

  it("calls usePublicKeys with search hydrated from ?search=mykey", () => {
    renderPage(["/?search=mykey"]);
    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ search: "mykey" }),
    );
  });

  it("calls usePublicKeys with empty search when URL has no search param", () => {
    renderPage(["/"]);
    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ search: "" }),
    );
  });
});

// ── URL writes (usePaginatedListState adoption) ───────────────────────────────

describe("PublicKeys — URL writes", () => {
  it("writes ?page=2 to the URL when the user navigates to page 2", async () => {
    const user = userEvent.setup();
    vi.mocked(usePublicKeys).mockReturnValue({
      publicKeys: Array.from({ length: 10 }, (_, i) =>
        makeKey({ fingerprint: `fp-${i}`, name: `key-${i}` }),
      ),
      totalCount: 25,
      isLoading: false,
      error: null,
    });
    renderPage();

    await user.click(screen.getByRole("button", { name: "Next page" }));

    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 2 }),
    );
  });

  it("resets page to 1 when the user types in the search field", async () => {
    const user = userEvent.setup();
    // Start on page 2
    renderPage(["/?page=2"]);

    const searchInput = screen.getByPlaceholderText(
      /search by name or fingerprint/i,
    );
    await user.type(searchInput, "a");

    // After typing, hook must be called with page=1
    expect(vi.mocked(usePublicKeys)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 }),
    );
  });
});
