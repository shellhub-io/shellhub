import type { ReactNode } from "react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { FirewallRule } from "../../../hooks/useFirewallRules";

// ── Module mocks ──────────────────────────────────────────────────────────────

vi.mock("../../../hooks/useFirewallRules", () => ({
  useFirewallRules: vi.fn(),
}));

vi.mock("../../../hooks/useFirewallRuleMutations", () => ({
  useDeleteFirewallRule: vi.fn(),
}));

// RestrictedAction gates buttons on permissions — always allow in tests.
vi.mock("../../../hooks/useHasPermission", () => ({
  useHasPermission: () => true,
}));

// RuleDrawer is not relevant for these tests and pulls in a lot of deps.
vi.mock("../RuleDrawer", () => ({
  default: () => null,
}));

// Flatten ConfirmDialog to a plain div so we can exercise the page's logic
// without animations, portals, or BaseDialog internals. Matches the pattern
// used in other page tests (DeleteNamespaceDialog, KeyDeleteDialog).
vi.mock("../../../components/common/ConfirmDialog", () => ({
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

import { useFirewallRules } from "../../../hooks/useFirewallRules";
import { useDeleteFirewallRule } from "../../../hooks/useFirewallRuleMutations";
import FirewallRules from "../index";

// ── Helpers ───────────────────────────────────────────────────────────────────

function makeRule(overrides: Partial<FirewallRule> = {}): FirewallRule {
  return {
    id: "rule-1",
    tenant_id: "tenant-abc",
    priority: 42,
    action: "allow",
    active: true,
    source_ip: ".*",
    username: ".*",
    filter: { hostname: ".*" },
    ...overrides,
  };
}

const mockMutateAsync = vi.fn();

beforeEach(() => {
  vi.clearAllMocks();
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
    render(<FirewallRules />);
    await user.click(screen.getByRole("button", { name: "Delete" }));
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
    await user.click(screen.getByRole("button", { name: "Delete" }));
    dialog = await getDialog();
    expect(within(dialog).queryByText("Transient")).not.toBeInTheDocument();
  });
});
