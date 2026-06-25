import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach, type Mock } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, useLocation } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ApiKey } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

const mockApiKeysImpl = vi.fn<
  () => {
    apiKeys: ApiKey[];
    totalCount: number;
    isLoading: boolean;
    error: null;
  }
>();

/** Spy that captures args passed to the hook. Does NOT call the impl itself —
 *  the factory below calls the impl exactly once and returns its value. */
const mockUseApiKeys: Mock = vi.fn();

vi.mock("@/hooks/useApiKeys", () => ({
  useApiKeys: (...args: unknown[]) => {
    mockUseApiKeys(...args);
    return mockApiKeysImpl();
  },
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
        <button type="button" onClick={onClose}>Cancel</button>
        <button type="button" onClick={() => void onConfirm()}>{confirmLabel}</button>
      </div>
    );
  },
}));

// ── Imports (after mocks) ─────────────────────────────────────────────────────

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
  mockApiKeysImpl.mockReturnValue({
    apiKeys: [makeApiKey()],
    totalCount: 1,
    isLoading: false,
    error: null,
  });
  vi.mocked(useDeleteApiKey).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

// ── Helpers ───────────────────────────────────────────────────────────────────

/** Exposes the current search string from inside the MemoryRouter. */
function LocationProbe({
  onLocation,
}: {
  onLocation: (search: string) => void;
}) {
  const loc = useLocation();
  onLocation(loc.search);
  return null;
}

function renderTab(initialEntries: string[] = ["/"]) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  let lastSearch = "";

  const result = render(
    <MemoryRouter initialEntries={initialEntries}>
      <QueryClientProvider client={queryClient}>
        <ApiKeysTab />
        <LocationProbe onLocation={(s) => { lastSearch = s; }} />
      </QueryClientProvider>
    </MemoryRouter>,
  );

  return { ...result, getSearch: () => lastSearch };
}

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("ApiKeysTab — pagination count display", () => {
  it("does not pass totalCount to DataTable so count is shown only in the header (single page)", () => {
    renderTab();

    // The header renders "1 key" exactly once — the Pagination below the table
    // must NOT duplicate it (totalCount is intentionally not forwarded to DataTable).
    const countMatches = screen.getAllByText(/\b1 key\b/);
    expect(countMatches).toHaveLength(1);
  });

  it("renders Prev/Next navigation buttons when there are more than PER_PAGE keys", () => {
    // 25 total keys, 10 on the current page -> totalPages=3, page=1
    const keys = Array.from({ length: 10 }, (_, i) =>
      makeApiKey({ name: `key-${i}`, created_by: `user-${i}` }),
    );
    mockApiKeysImpl.mockReturnValue({
      apiKeys: keys,
      totalCount: 25,
      isLoading: false,
      error: null,
    });

    renderTab();

    // Prev/Next buttons must exist even though totalCount is not passed to DataTable
    expect(screen.getByRole("button", { name: /prev/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
    // Page indicator
    expect(screen.getByText("1 / 3")).toBeInTheDocument();
  });
});

describe("ApiKeysTab — sorting", () => {
  it("requests created_at/desc sort by default", () => {
    renderTab();
    expect(mockUseApiKeys).toHaveBeenCalledWith(
      expect.objectContaining({ sortBy: "created_at", orderBy: "desc" }),
    );
  });

  it("toggles sort when the Name header is clicked", async () => {
    const user = userEvent.setup();
    renderTab();

    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    let calls = mockUseApiKeys.mock.calls;
    let last = calls[calls.length - 1][0];
    expect(last).toMatchObject({ sortBy: "name", orderBy: "asc" });

    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    calls = mockUseApiKeys.mock.calls;
    last = calls[calls.length - 1][0];
    expect(last).toMatchObject({ sortBy: "name", orderBy: "desc" });
  });
});

describe("ApiKeysTab — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderTab();
    await user.click(screen.getByRole("button", { name: "Delete API key" }));
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

// ── URL sync (usePaginatedListState adoption, prefix "key") ──────────────────

describe("ApiKeysTab — URL sync with prefix 'key'", () => {
  it("hydrates page from ?key.page=3 — hook receives page 3", () => {
    renderTab(["/?key.page=3"]);
    expect(mockUseApiKeys).toHaveBeenCalledWith(
      expect.objectContaining({ page: 3 }),
    );
  });

  it("clicking Next writes key.page=2 to the URL (not bare page=2)", async () => {
    const user = userEvent.setup();
    mockApiKeysImpl.mockReturnValue({
      apiKeys: Array.from({ length: 10 }, (_, i) =>
        makeApiKey({ name: `key-${i}`, created_by: `user-${i}` }),
      ),
      totalCount: 25,
      isLoading: false,
      error: null,
    });
    const { getSearch } = renderTab();

    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() => {
      const sp = new URLSearchParams(getSearch());
      expect(sp.get("key.page")).toBe("2");
      expect(sp.get("page")).toBeNull(); // bare page must not appear
    });
  });

  it("does not consume a bare ?page=5 param as key.page — hook receives page 1", () => {
    renderTab(["/?page=5"]);
    // The hook must receive the default page (1), not the bare page=5
    expect(mockUseApiKeys).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 }),
    );
    // The bare page=5 must survive in the URL untouched
    // (LocationProbe captures it at render time)
  });
});
