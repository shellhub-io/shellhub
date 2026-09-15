import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import ApiKeysTab from "../ApiKeysTab";
import type { ApiKey } from "@/client/model";
import { createTestWrapper } from "@/tests/wrapper";
import { LocationProbe } from "@/tests/LocationProbe";
import { useAuthStore } from "@/stores/authStore";

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

let lastApiKeysUrl: URL | null;

function setApiKeys(keys: ApiKey[], total?: number) {
  server.use(
    http.get("*/api/namespaces/api-key", ({ request }) => {
      lastApiKeysUrl = new URL(request.url);
      return jsonWithTotal(keys, total ?? keys.length);
    }),
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  lastApiKeysUrl = null;
  setApiKeys([mockApiKey()]);
  server.use(
    http.delete(
      "*/api/namespaces/api-key/:key",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
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
    setApiKeys(keys, 25);

    renderTab();
    await screen.findByText("key-0");

    expect(screen.getByRole("button", { name: /prev/i })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /next/i })).toBeInTheDocument();
    expect(screen.getByText("1 / 3")).toBeInTheDocument();
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
    server.use(
      http.delete("*/api/namespaces/api-key/:key", () =>
        HttpResponse.json(
          { message: "Key is protected" },
          { status: 403 },
        ),
      ),
    );
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("Key is protected")).toBeInTheDocument(),
    );
    expect(dialog).toBeInTheDocument();
  });

  it("shows the status code as fallback when the server returns no message", async () => {
    server.use(
      http.delete("*/api/namespaces/api-key/:key", () =>
        HttpResponse.json({}, { status: 500 }),
      ),
    );
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("500")).toBeInTheDocument(),
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
  it("does not consume a bare ?page=5 param as key.page — API receives page 1", async () => {
    renderTab(["/?page=5"]);
    await waitFor(() => {
      expect(lastApiKeysUrl).not.toBeNull();
      expect(lastApiKeysUrl!.searchParams.get("page")).toBe("1");
    });
  });
});
