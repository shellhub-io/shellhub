import { useState, useCallback, useRef, useEffect } from "react";

/**
 * Shared entity shape used by both device and container action flows (and by
 * the `*ActionDialog` components). Both domains require only `uid` and `name`.
 */
export interface EntityBase {
  uid: string;
  name: string;
}

/**
 * The set of actions supported for both devices and containers.
 */
export type EntityAction = "accept" | "reject" | "remove";

export interface Operation {
  entity: EntityBase;
  action: EntityAction;
}

export interface UseActionDialogStateOptions {
  enableBillingWarning: boolean;
  onSuccess?: (action: EntityAction) => void;
}

export interface UseActionDialogStateResult {
  operation: Operation | undefined;
  requestAction: (entity: EntityBase, action: EntityAction) => void;
  close: () => void;
  billingWarningOpen: boolean;
  closeBillingWarning: () => void;
  onBillingWarning: (() => void) | undefined;
  /**
   * Invoke `onSuccess` with the action that was confirmed. Callers (portals/dialogs) must pass
   * the action explicitly — captured at `handleConfirm` time — rather than relying on the live
   * operation state, so that the correct action is reported even if the dialog was cancelled
   * (via `close()`) while the mutation was in-flight.
   */
  runSuccess: (action: EntityAction) => void;
  /** True when billing warning is supported for this controller (mirrors enableBillingWarning). */
  billingEnabled: boolean;
}

/**
 * Shared state core for action-triggered dialogs with optional billing warning gate.
 *
 * - `requestAction(entity, action)` — open the dialog for a specific entity/action pair.
 * - `close()` — clear the operation (dismiss the dialog).
 * - `billingWarningOpen` / `closeBillingWarning` — control the billing warning overlay.
 * - `onBillingWarning` — defined only when `enableBillingWarning` is truthy; call it to show
 *   the billing warning before the action proceeds. Callers (e.g. `useDeviceActions`) are
 *   responsible for passing `enableBillingWarning: !!getConfig().cloud`.
 * - `runSuccess(action)` — invoke `onSuccess` with the explicitly passed action. The caller
 *   must capture the action at confirm time and pass it here, so the correct action is reported
 *   even if the dialog was cancelled (via `close()`) while a mutation was in-flight. Does NOT
 *   close the dialog; the portal/dialog is responsible for calling `close()` after this.
 *
 * All returned functions are stable across re-renders (wrapped in `useCallback`).
 */
export function useActionDialogState({
  enableBillingWarning,
  onSuccess,
}: UseActionDialogStateOptions): UseActionDialogStateResult {
  const [operation, setOperation] = useState<Operation | undefined>(undefined);
  const [billingWarningOpen, setBillingWarningOpen] = useState(false);

  const requestAction = useCallback((entity: EntityBase, action: EntityAction) => {
    setOperation({ entity, action });
  }, []);

  const close = useCallback(() => {
    setOperation(undefined);
  }, []);

  const closeBillingWarning = useCallback(() => {
    setBillingWarningOpen(false);
  }, []);

  const openBillingWarning = useCallback(() => {
    setOperation(undefined);
    setBillingWarningOpen(true);
  }, []);

  const onBillingWarning = enableBillingWarning ? openBillingWarning : undefined;

  // Keep a ref to onSuccess so runSuccess always invokes the latest callback
  // without being in the useCallback dep array. This avoids stale-closure bugs
  // when callers pass inline lambdas and prevents runSuccess identity churn on
  // every requestAction() call (which would force re-renders in memoised consumers).
  const onSuccessRef = useRef(onSuccess);
  useEffect(() => {
    onSuccessRef.current = onSuccess;
  });

  // The caller (portal/dialog) passes the action it confirmed explicitly.
  // This ensures the correct action is forwarded even when the dialog was
  // cancelled (close() cleared operation state) while the mutation was in-flight.
  const runSuccess = useCallback((action: EntityAction) => {
    onSuccessRef.current?.(action);
  }, []);

  return {
    operation,
    requestAction,
    close,
    billingWarningOpen,
    closeBillingWarning,
    onBillingWarning,
    runSuccess,
    billingEnabled: enableBillingWarning,
  };
}
