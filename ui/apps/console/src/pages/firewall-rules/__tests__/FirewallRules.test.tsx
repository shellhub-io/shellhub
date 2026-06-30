import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router-dom";
import { FirewallRulesResponse } from "@/client";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("@/hooks/useFirewallRules", () => ({
  useFirewallRules: vi.fn(),
}));

vi.mock("@/hooks/useFirewallRuleMutations", () => ({
  useDeleteFirewallRule: vi.fn(),
}));

// RestrictedAction gates buttons on permissions — always allow in tests.
vi.mock("@/hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

// RuleDrawer is not relevant for these tests and pulls in a lot of deps.
vi.mock("../RuleDrawer", () => ({
  default: () => null,
}));

// Flatten ConfirmDialog to a plain div so we can exercise the page's logic
// without animations, portals, or BaseDialog internals.
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

// Capture DataTable props on every render so we can assert pagination suppression.
const capturedDataTableProps: Record<string, unknown>[] = [];
vi.mock("@shellhub/design-system/components", async (importOriginal) => {
  const actual =
    await importOriginal<typeof import("@shellhub/design-system/components")>();
  return {
    ...actual,
    DataTable: (props: Record<string, unknown>) => {
      capturedDataTableProps.push({ ...props });
      return actual.DataTable(props as unknown as Parameters<typeof actual.DataTable>[0]);
    },
  };
});

// ── Imports (after mocks) ─────────────────────────────────────────────────────

import { useFirewallRules } from "@/hooks/useFirewallRules";
import { useDeleteFirewallRule } from "@/hooks/useFirewallRuleMutations";
import FirewallRules from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeRule(
  overrides: Partial<FirewallRulesResponse> = {},
): FirewallRulesResponse {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 42,
    action: "allow",
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*", tags: [] },
    ...overrides,
  };
}

const mockMutateAsync = vi.fn();

function renderPage(initialEntries: string[] = ["/"]) {
  return render(
    <MemoryRouter initialEntries={initialEntries}>
      <FirewallRules />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  vi.clearAllMocks();
  capturedDataTableProps.length = 0;
  vi.mocked(useFirewallRules).mockReturnValue({
    rules: [makeRule()],
    totalCount: 1,
    isLoading: false,
    error: null,
  });
  vi.mocked(useDeleteFirewallRule).mockReturnValue({
    mutateAsync: mockMutateAsync,
  } as never);
});

// ── Tests ─────────────────────────────────────────────────────────────────────

describe("FirewallRules — delete error handling", () => {
  // The row-action Delete button has no visible text; its accessible name
  // comes from its `title` attribute. The dialog Delete button has visible
  // text "Delete". Before the dialog opens there is only one match; after
  // the dialog opens we scope dialog queries with `within(dialog)`.
  async function openDeleteDialog() {
    const user = userEvent.setup();
    renderPage();
    await user.click(
      screen.getByRole("button", { name: /^delete firewall rule/i }),
    );
    return user;
  }

  async function getDialog() {
    return screen.findByRole("dialog", { name: /delete firewall rule/i });
  }

  it("shows the mutation error message inside the dialog when deletion fails", async () => {
    mockMutateAsync.mockRejectedValue(new Error("Permission denied"));
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(within(dialog).getByText("Permission denied")).toBeInTheDocument(),
    );
    // Dialog stays open with the error visible.
    expect(dialog).toBeInTheDocument();
  });

  it("shows a generic fallback message when the rejection is not an Error", async () => {
    mockMutateAsync.mockRejectedValue("boom");
    const user = await openDeleteDialog();
    const dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));

    await waitFor(() =>
      expect(
        within(dialog).getByText(/failed to delete firewall rule/i),
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
        screen.queryByRole("dialog", { name: /delete firewall rule/i }),
      ).not.toBeInTheDocument(),
    );
    expect(
      screen.queryByText(/failed to delete firewall rule/i),
    ).not.toBeInTheDocument();
  });

  it("clears any previous error when the dialog is cancelled and reopened", async () => {
    mockMutateAsync.mockRejectedValueOnce(new Error("Transient"));
    const user = await openDeleteDialog();
    let dialog = await getDialog();

    await user.click(within(dialog).getByRole("button", { name: /^delete$/i }));
    await within(dialog).findByText("Transient");

    await user.click(within(dialog).getByRole("button", { name: /cancel/i }));
    await waitFor(() =>
      expect(
        screen.queryByRole("dialog", { name: /delete firewall rule/i }),
      ).not.toBeInTheDocument(),
    );

    // Reopen — the stale error text should be gone before the next attempt.
    await user.click(
      screen.getByRole("button", { name: /^delete firewall rule/i }),
    );
    dialog = await getDialog();
    expect(within(dialog).queryByText("Transient")).not.toBeInTheDocument();
  });
});

// ── usePaginatedListState adoption ───────────────────────────────────────────

describe("FirewallRules — URL hydration", () => {
  it("hydrates search from URL on mount", () => {
    vi.mocked(useFirewallRules).mockReturnValue({
      rules: [makeRule({ action: "allow" })],
      totalCount: 1,
      isLoading: false,
      error: null,
    });
    renderPage(["/?search=allow"]);
    expect(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
    ).toHaveValue("allow");
  });

  it("hydrates page from URL and passes it to the hook", () => {
    renderPage(["/?page=3"]);
    expect(vi.mocked(useFirewallRules)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 3 }),
    );
  });

  it("calls useFirewallRules with page=1 when URL has no params", () => {
    renderPage(["/"]);
    expect(vi.mocked(useFirewallRules)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 1 }),
    );
  });

  it("setSearch resets page to 1 in the URL", async () => {
    const user = userEvent.setup();
    renderPage(["/?page=3"]);

    // Confirm page=3 was hydrated
    expect(vi.mocked(useFirewallRules)).toHaveBeenCalledWith(
      expect.objectContaining({ page: 3 }),
    );

    await user.type(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
      "allow",
    );

    await waitFor(() => {
      const calls = vi.mocked(useFirewallRules).mock.calls;
      const lastCall = calls.at(-1)![0];
      expect(lastCall).toBeDefined();
      expect(lastCall?.page).toBe(1);
    });
  });
});

describe("FirewallRules — pagination suppressed while searching", () => {
  beforeEach(() => {
    vi.mocked(useFirewallRules).mockReturnValue({
      rules: [makeRule({ id: "r1", action: "allow", priority: 1 })],
      totalCount: 1,
      isLoading: false,
      error: null,
    });
  });

  it("passes page/totalPages/onPageChange to DataTable when search is empty", () => {
    renderPage();
    const last = capturedDataTableProps.at(-1);
    expect(last).toBeDefined();
    expect(last).toHaveProperty("page");
    expect(last).toHaveProperty("totalPages");
    expect(last).toHaveProperty("onPageChange");
  });

  it("omits page/totalPages/onPageChange from DataTable while search is non-empty", async () => {
    const user = userEvent.setup();
    renderPage();

    await user.type(
      screen.getByRole("searchbox", {
        name: "Search firewall rules by action, priority, IP, or username",
      }),
      "allow",
    );

    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).toBeDefined();
      expect(last).not.toHaveProperty("page");
      expect(last).not.toHaveProperty("totalPages");
      expect(last).not.toHaveProperty("onPageChange");
    });
  });

  it("re-enables pagination props after search is cleared", async () => {
    const user = userEvent.setup();
    renderPage();

    const searchbox = screen.getByRole("searchbox", {
      name: "Search firewall rules by action, priority, IP, or username",
    });

    await user.type(searchbox, "allow");

    // Confirm pagination is suppressed
    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).not.toHaveProperty("page");
    });

    // Clear the search
    await user.clear(searchbox);

    // Pagination should be restored
    await waitFor(() => {
      const last = capturedDataTableProps.at(-1);
      expect(last).toBeDefined();
      expect(last).toHaveProperty("page");
      expect(last).toHaveProperty("totalPages");
      expect(last).toHaveProperty("onPageChange");
    });
  });
});
