import { describe, it, expect, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { UseContainerActionsResult } from "@/hooks/useContainerActions";
import type { EntityBase } from "@/hooks/useActionDialogState";

// Stand-in for the real dialog: it surfaces the wiring the portal owns (open
// state, entity, action, and the onClose/onSuccess callbacks) through the DOM
// so tests can drive it the way a user would, instead of inspecting props.
vi.mock("../ContainerActionDialog", () => ({
  default: ({ open, container, action, onClose, onSuccess }: {
    open: boolean;
    container: EntityBase | null;
    action: string;
    onClose: () => void;
    onSuccess?: () => void;
  }) =>
    open ? (
      <div role="dialog" aria-label={`${action} ${container?.name}`}>
        <button type="button" onClick={onClose}>cancel</button>
        {onSuccess && <button type="button" onClick={onSuccess}>confirm</button>}
      </div>
    ) : null,
}));

vi.mock("@/components/billing/BillingWarning", () => ({
  default: () => <div data-testid="billing-warning" />,
}));

import ContainerActionsPortal from "../ContainerActionsPortal";

function renderPortal(overrides: Partial<UseContainerActionsResult> = {}) {
  const controller: UseContainerActionsResult = {
    operation: undefined,
    requestAction: vi.fn(),
    close: vi.fn(),
    billingWarningOpen: false,
    closeBillingWarning: vi.fn(),
    onBillingWarning: undefined,
    runSuccess: vi.fn(),
    billingEnabled: false,
    ...overrides,
  };
  render(<ContainerActionsPortal controller={controller} />);
  return controller;
}

describe("ContainerActionsPortal", () => {
  it("opens the dialog for the active operation", () => {
    renderPortal({ operation: { entity: { uid: "uid-1", name: "my-container" }, action: "accept" } });

    expect(screen.getByRole("dialog", { name: "accept my-container" })).toBeInTheDocument();
  });

  it("keeps the dialog closed when there is no operation", () => {
    renderPortal();

    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("runs the operation's action when the dialog confirms", async () => {
    const controller = renderPortal({
      operation: { entity: { uid: "uid-2", name: "c2" }, action: "remove" },
    });

    await userEvent.click(screen.getByRole("button", { name: "confirm" }));

    expect(controller.runSuccess).toHaveBeenCalledWith("remove");
  });

  it("closes when the dialog is cancelled", async () => {
    const controller = renderPortal({
      operation: { entity: { uid: "uid-3", name: "c3" }, action: "accept" },
    });

    await userEvent.click(screen.getByRole("button", { name: "cancel" }));

    expect(controller.close).toHaveBeenCalledOnce();
  });

  it("renders BillingWarning only when billing is enabled", () => {
    renderPortal({ billingEnabled: true });
    expect(screen.getByTestId("billing-warning")).toBeInTheDocument();
  });

  it("does not render BillingWarning when billing is disabled", () => {
    renderPortal();
    expect(screen.queryByTestId("billing-warning")).not.toBeInTheDocument();
  });
});
