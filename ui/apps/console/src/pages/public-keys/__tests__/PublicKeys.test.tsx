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
