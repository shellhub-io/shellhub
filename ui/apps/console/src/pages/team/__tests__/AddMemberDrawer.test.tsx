import { describe, it, expect, vi, beforeEach } from "vitest";
import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import AddMemberDrawer from "../AddMemberDrawer";

/* ------------------------------------------------------------------ */
/* Mocks                                                               */
/* ------------------------------------------------------------------ */

const mockAddMemberMutateAsync = vi.fn();

vi.mock("@/hooks/useMemberMutations", () => ({
  useAddMember: () => ({
    mutateAsync: mockAddMemberMutateAsync,
    isPending: false,
  }),
}));

vi.mock("@/components/common/Drawer", () => ({
  default: ({
    open,
    onClose,
    title,
    children,
    footer,
  }: {
    open: boolean;
    onClose: () => void;
    title: string;
    children: React.ReactNode;
    footer?: React.ReactNode;
  }) => {
    if (!open) return null;
    return (
      <div role="dialog" aria-label={title}>
        <h2>{title}</h2>
        <button type="button" onClick={onClose}>Close</button>
        {children}
        {footer ?? null}
      </div>
    );
  },
}));

/* ------------------------------------------------------------------ */
/* Setup                                                               */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
  mockAddMemberMutateAsync.mockResolvedValue({});
});

/* ================================================================== */
/* Tests                                                               */
/* ================================================================== */

describe("AddMemberDrawer", () => {
  it("blocks invalid emails and shows an inline validation error", async () => {
    const user = userEvent.setup();
    render(<AddMemberDrawer open onClose={vi.fn()} tenantId="t1" />);

    const input = screen.getByPlaceholderText(/user@example.com/i);
    await user.type(input, "foo");
    expect(screen.getByRole("button", { name: /add member/i })).toBeDisabled();

    const form = input.closest("form");
    expect(form).not.toBeNull();
    fireEvent.submit(form as HTMLFormElement);

    expect(mockAddMemberMutateAsync).not.toHaveBeenCalled();
    expect(
      screen.getByText(/enter a valid email address/i),
    ).toBeInTheDocument();
    expect(input).toHaveAttribute("aria-invalid", "true");
  });

  it("submits trimmed valid email addresses", async () => {
    const user = userEvent.setup();
    render(<AddMemberDrawer open onClose={vi.fn()} tenantId="t1" />);

    fireEvent.change(screen.getByPlaceholderText(/user@example.com/i), {
      target: { value: "  alice@example.com  " },
    });
    await user.click(screen.getByRole("button", { name: /add member/i }));

    await waitFor(() =>
      expect(mockAddMemberMutateAsync).toHaveBeenCalledWith({
        path: { tenant: "t1" },
        body: { email: "alice@example.com", role: "operator" },
      }),
    );
  });
});
