import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { http, HttpResponse } from "msw";
import { server, jsonWithTotal } from "@/tests/msw";
import PublicKeys from "../index";
import { createTestWrapper } from "@/tests/wrapper";
import { mockPublicKey } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

vi.mock("../KeyDrawer", () => ({
  default: () => null,
}));

vi.mock("@/components/common/CopyButton", async () => ({
  default: (await import("@/tests/mocks")).MockCopyButton,
}));

vi.mock("@/components/common/ConfirmDialog", async () => ({
  default: (await import("@/tests/mocks")).MockConfirmDialog,
}));

vi.mock("@/hooks/useDebouncedValue", () => ({
  useDebouncedValue: <T,>(value: T) => value,
}));

let lastKeysUrl: URL | null;

function setKeys(
  keys: ReturnType<typeof mockPublicKey>[],
  total?: number,
) {
  server.use(
    http.get("*/api/sshkeys/public-keys", ({ request }) => {
      lastKeysUrl = new URL(request.url);
      return jsonWithTotal(keys, total ?? keys.length);
    }),
  );
}

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <PublicKeys />
    </MemoryRouter>,
    { wrapper: createTestWrapper() },
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  lastKeysUrl = null;
  useAuthStore.setState({ role: "owner" });
  setKeys([mockPublicKey()]);
  server.use(
    http.delete(
      "*/api/sshkeys/public-keys/:fingerprint",
      () => new HttpResponse(null, { status: 204 }),
    ),
  );
});

describe("PublicKeys — delete error handling", () => {
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderPage();
    await screen.findByText("my-key");
    await user.click(screen.getByRole("button", { name: /^delete/i }));
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete public key/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    server.use(
      http.delete("*/api/sshkeys/public-keys/:fingerprint", () =>
        HttpResponse.json(
          { message: "Fingerprint in use" },
          { status: 403 },
        ),
      ),
    );
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

  it("shows the status code as fallback when the server returns no message", async () => {
    server.use(
      http.delete("*/api/sshkeys/public-keys/:fingerprint", () =>
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
        screen.queryByRole("dialog", { name: /delete public key/i }),
      ).not.toBeInTheDocument(),
    );
  });
});

describe("PublicKeys — URL hydration", () => {
  it("passes page=3 when URL has ?page=3", async () => {
    renderPage(["/?page=3"]);
    await waitFor(() => {
      expect(lastKeysUrl).not.toBeNull();
      expect(lastKeysUrl!.searchParams.get("page")).toBe("3");
    });
  });

  it("passes page=1 when URL has no page param", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      expect(lastKeysUrl).not.toBeNull();
      expect(lastKeysUrl!.searchParams.get("page")).toBe("1");
    });
  });

  it("passes a filter containing the search term when URL has ?search=mykey", async () => {
    renderPage(["/?search=mykey"]);
    await waitFor(() => {
      expect(lastKeysUrl).not.toBeNull();
      const filter = lastKeysUrl!.searchParams.get("filter") ?? "";
      expect(atob(filter)).toContain("mykey");
    });
  });

  it("passes no filter when URL has no search param", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      expect(lastKeysUrl).not.toBeNull();
      expect(lastKeysUrl!.searchParams.get("filter")).toBeNull();
    });
  });
});

describe("PublicKeys — URL writes", () => {
  it("passes page=2 when the user navigates to page 2", async () => {
    const user = userEvent.setup();
    setKeys(
      Array.from({ length: 10 }, (_, i) =>
        mockPublicKey({ fingerprint: `fp-${i}`, name: `key-${i}` }),
      ),
      25,
    );
    renderPage();

    await screen.findByText("key-0");

    await user.click(screen.getByRole("button", { name: "Next page" }));

    await waitFor(() => {
      expect(lastKeysUrl!.searchParams.get("page")).toBe("2");
    });
  });

  it("resets page to 1 when the user types in the search field", async () => {
    const user = userEvent.setup();
    renderPage(["/?page=2"]);

    await screen.findByText("my-key");

    const searchInput = screen.getByPlaceholderText(
      /search by name or fingerprint/i,
    );
    await user.type(searchInput, "a");

    await waitFor(() => {
      expect(lastKeysUrl!.searchParams.get("page")).toBe("1");
    });
  });
});
