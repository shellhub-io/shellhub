import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ApiKeysTab from "../ApiKeysTab";
import type { ApiKey } from "@/client";
import { createTestWrapper } from "@/tests/wrapper";
import { LocationProbe } from "@/tests/LocationProbe";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { useAuthStore } from "@/stores/authStore";

const mockApiKeyList = vi.hoisted(() => vi.fn());
const mockApiKeyDelete = vi.hoisted(() => vi.fn());

vi.mock("@/client/sdk.gen", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@/client/sdk.gen")>();
  return {
    ...actual,
    apiKeyList: mockApiKeyList,
    apiKeyDelete: mockApiKeyDelete,
  };
});

vi.mock("../GenerateKeyDrawer", () => ({
  default: () => null,
}));

vi.mock("../EditKeyDrawer", () => ({
  default: () => null,
}));

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

function mockApiKey(overrides: Partial<ApiKey> = {}): ApiKey {
  return {
    tenant_id: "tenant-456",
    created_by: "user-123",
    role: "administrator",
    name: "prod-key",
    expires_in: Math.floor(Date.now() / 1000) + 3600 * 24 * 30,
    created_at: "2024-01-01T00:00:00Z",
    updated_at: "2024-01-01T00:00:00Z",
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  mockApiKeyList.mockResolvedValue(paginatedResponse([mockApiKey()]));
  mockApiKeyDelete.mockResolvedValue(mockSdkResponse(undefined));
  useAuthStore.setState({ role: "owner" });
});

function renderTab(initialEntries: string[] = ["/"]) {
  let lastSearch = "";

  const result = render(
    <>
      <ApiKeysTab />
      <LocationProbe
        onLocation={(s) => {
          lastSearch = s;
        }}
      />
    </>,
    { wrapper: createTestWrapper({ initialEntries }) },
  );

  return { ...result, getSearch: () => lastSearch };
}

describe("ApiKeysTab — pagination count display", () => {
  it("does not pass totalCount to DataTable so count is shown only in the header (single page)", async () => {
    renderTab();
    await screen.findByText("prod-key");
    const countMatches = screen.getAllByText(/\b1 key\b/);
    expect(countMatches).toHaveLength(1);
  });

  it("renders Prev/Next navigation buttons when there are more than PER_PAGE keys", async () => {
    const keys = Array.from({ length: 10 }, (_, i) =>
      mockApiKey({ name: `key-${i}`, created_by: `user-${i}` }),
    );
    mockApiKeyList.mockResolvedValue(paginatedResponse(keys, 25));

    renderTab();
    await screen.findByText("key-0");

    expect(screen.getByRole("button", { name: /prev/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
    expect(screen.getByText("1 / 3")).toBeInTheDocument();
  });
});

describe("ApiKeysTab — sorting", () => {
  it("requests created_at/desc sort by default", async () => {
    renderTab();
    await screen.findByText("prod-key");
    expect(mockApiKeyList).toHaveBeenCalledWith(
      expect.objectContaining({
        query: expect.objectContaining({
          sort_by: "created_at",
          order_by: "desc",
        }),
      }),
    );
  });

  it("toggles sort when the Name header is clicked", async () => {
    const user = userEvent.setup();
    renderTab();
    await screen.findByText("prod-key");

    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    await waitFor(() => {
      expect(mockApiKeyList).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({
            sort_by: "name",
            order_by: "asc",
          }),
        }),
      );
    });

    await user.click(screen.getByRole("button", { name: "Sort by Name" }));
    await waitFor(() => {
      expect(mockApiKeyList).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({
            sort_by: "name",
            order_by: "desc",
          }),
        }),
      );
    });
  });
});

describe("ApiKeysTab — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderTab();
    await screen.findByText("prod-key");
    await user.click(screen.getByRole("button", { name: "Delete API key" }));
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete api key/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    mockApiKeyDelete.mockRejectedValue(new Error("Key is protected"));
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("Key is protected")).toBeInTheDocument(),
    );
    expect(dialog).toBeInTheDocument();
  });

  it("shows a generic fallback message when the rejection is not an Error", async () => {
    mockApiKeyDelete.mockRejectedValue("boom");
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

describe("ApiKeysTab — URL sync with prefix 'key'", () => {
  it("hydrates page from ?key.page=3 — SDK receives page 3", async () => {
    renderTab(["/?key.page=3"]);
    await waitFor(() => {
      expect(mockApiKeyList).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 3 }),
        }),
      );
    });
  });

  it("clicking Next writes key.page=2 to the URL (not bare page=2)", async () => {
    const user = userEvent.setup();
    mockApiKeyList.mockResolvedValue(
      paginatedResponse(
        Array.from({ length: 10 }, (_, i) =>
          mockApiKey({ name: `key-${i}`, created_by: `user-${i}` }),
        ),
        25,
      ),
    );
    const { getSearch } = renderTab();

    await screen.findByText("key-0");

    await user.click(screen.getByRole("button", { name: /next/i }));

    await waitFor(() => {
      const sp = new URLSearchParams(getSearch());
      expect(sp.get("key.page")).toBe("2");
      expect(sp.get("page")).toBeNull();
    });
  });

  it("does not consume a bare ?page=5 param as key.page — SDK receives page 1", async () => {
    renderTab(["/?page=5"]);
    await waitFor(() => {
      expect(mockApiKeyList).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 1 }),
        }),
      );
    });
  });
});
