import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import PublicKeys from "../index";
import { mockSdkResponse, paginatedResponse } from "@/tests/sdk";
import { createTestWrapper } from "@/tests/wrapper";
import { mockPublicKey } from "@/tests/factories";
import { useAuthStore } from "@/stores/authStore";

const sdk = vi.hoisted(() =>
  mockSdkGen({
    getPublicKeys: vi.fn(),
    deletePublicKey: vi.fn(),
  }),
);

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
  sdk.getPublicKeys.mockResolvedValue(paginatedResponse([mockPublicKey()]));
  sdk.deletePublicKey.mockResolvedValue(mockSdkResponse(undefined));
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
    sdk.deletePublicKey.mockRejectedValue(new Error("Fingerprint in use"));
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
    sdk.deletePublicKey.mockRejectedValue({ status: 500 });
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
  it("passes page=3 to the SDK when URL has ?page=3", async () => {
    renderPage(["/?page=3"]);
    await waitFor(() => {
      expect(sdk.getPublicKeys).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 3 }),
        }),
      );
    });
  });

  it("passes page=1 to the SDK when URL has no page param", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      expect(sdk.getPublicKeys).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 1 }),
        }),
      );
    });
  });

  it("passes a filter containing the search term when URL has ?search=mykey", async () => {
    renderPage(["/?search=mykey"]);
    await waitFor(() => {
      const call = sdk.getPublicKeys.mock.calls.at(-1)?.[0] as {
        query?: { filter?: string };
      };
      const decoded = atob(call?.query?.filter ?? "");
      expect(decoded).toContain("mykey");
    });
  });

  it("passes no filter to the SDK when URL has no search param", async () => {
    renderPage(["/"]);
    await waitFor(() => {
      const call = sdk.getPublicKeys.mock.calls[0]?.[0] as {
        query?: { filter?: string };
      };
      expect(call?.query?.filter).toBeUndefined();
    });
  });
});

describe("PublicKeys — URL writes", () => {
  it("passes page=2 to the SDK when the user navigates to page 2", async () => {
    const user = userEvent.setup();
    sdk.getPublicKeys.mockResolvedValue(
      paginatedResponse(
        Array.from({ length: 10 }, (_, i) =>
          mockPublicKey({ fingerprint: `fp-${i}`, name: `key-${i}` }),
        ),
        25,
      ),
    );
    renderPage();

    await screen.findByText("key-0");

    await user.click(screen.getByRole("button", { name: "Next page" }));

    await waitFor(() => {
      expect(sdk.getPublicKeys).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 2 }),
        }),
      );
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
      expect(sdk.getPublicKeys).toHaveBeenCalledWith(
        expect.objectContaining({
          query: expect.objectContaining({ page: 1 }),
        }),
      );
    });
  });
});
