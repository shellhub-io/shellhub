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
import { decodeB64url } from "@/tests/decodeB64url";

vi.mock("../KeyModal", () => ({
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

function setKeys(
  keys: ReturnType<typeof mockPublicKey>[],
  total?: number,
) {
  server.use(
    http.get("*/api/sshkeys/public-keys", () =>
      jsonWithTotal(keys, total ?? keys.length),
    ),
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

    await user.click(within(dialog).getByRole("button", { name: /^delete key$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText("Failed to delete public key."),
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

    await user.click(within(dialog).getByRole("button", { name: /^delete key$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText("Failed to delete public key."),
      ).toBeInTheDocument(),
    );
  });

  it("closes the dialog and does not show an error on successful deletion", async () => {
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete key$/i }));

    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: /delete public key/i }),
      ).not.toBeInTheDocument(),
    );
  });
});

describe("PublicKeys — search", () => {
  it("matches the search against the key's name or fingerprint", async () => {
    let filter: string | null = null;
    server.use(
      http.get("*/api/sshkeys/public-keys", ({ request }) => {
        filter = new URL(request.url).searchParams.get("filter");
        return jsonWithTotal([]);
      }),
    );
    renderPage(["/?search=ab12"]);

    await waitFor(() => expect(filter).not.toBeNull());
    expect(decodeB64url(filter!)).toEqual([
      { type: "operator", params: { name: "or" } },
      { type: "property", params: { name: "name", operator: "contains", value: "ab12" } },
      { type: "operator", params: { name: "or" } },
      {
        type: "property",
        params: { name: "fingerprint", operator: "contains", value: "ab12" },
      },
    ]);
  });
});

describe("PublicKeys — list error", () => {
  it("does not show the first-key splash when the list fails to load", async () => {
    server.use(
      http.get("*/api/sshkeys/public-keys", () => HttpResponse.json({}, { status: 500 })),
    );
    renderPage();

    await screen.findByRole("alert");
    expect(screen.queryByText("Add your first key")).not.toBeInTheDocument();
  });
});

describe("PublicKeys — empty page", () => {
  it("keeps the table when a page comes back empty but the namespace has keys", async () => {
    setKeys([], 11);
    renderPage(["/?page=2"]);

    expect(await screen.findByText("No public keys found")).toBeInTheDocument();
    expect(screen.queryByText("Add your first key")).not.toBeInTheDocument();
  });
});
