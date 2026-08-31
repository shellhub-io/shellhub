import { useState, useCallback, useRef, useEffect } from "react";

/**
 * The least a thing needs for a confirmation dialog to name it: something to act on and
 * something to call it.
 */
export interface EntityBase {
  uid: string;
  name: string;
}

/**
 * What is being confirmed. The wording and the danger of the dialog follow from this.
 */
export type EntityOperation = "accept" | "reject" | "remove";

/**
 * A pending confirmation: which entity, and what is about to happen to it.
 */
export interface Action {
  entity: EntityBase;
  operation: EntityOperation;
}

interface UseActionDialogOptions {
  onSuccess?: (operation: EntityOperation) => void;
}

/**
 * What useActionDialog hands back. actionKey changes with every request, so a dialog keyed on it
 * remounts and cannot show state left from the previous confirmation.
 */
export interface UseActionDialogResult {
  action: Action | undefined;
  actionKey: string;
  requestAction: (entity: EntityBase, operation: EntityOperation) => void;
  close: () => void;
  handleSuccess: (operation: EntityOperation) => void;
}

/**
 * Tracks which row's confirmation dialog is open, so a table needs one dialog rather than one
 * per row. onSuccess fires after the action completes, for the caller's toast or refetch.
 */
export function useActionDialog(
  options?: UseActionDialogOptions,
): UseActionDialogResult {
  const [action, setAction] = useState<Action | undefined>(undefined);

  const requestAction = useCallback(
    (entity: EntityBase, operation: EntityOperation) => {
      setAction({ entity, operation });
    },
    [],
  );

  const close = useCallback(() => {
    setAction(undefined);
  }, []);

  const onSuccessRef = useRef(options?.onSuccess);
  useEffect(() => {
    onSuccessRef.current = options?.onSuccess;
  });

  const handleSuccess = useCallback((operation: EntityOperation) => {
    onSuccessRef.current?.(operation);
  }, []);

  return {
    action,
    actionKey: action
      ? `${action.operation}/${action.entity.uid}`
      : "closed",
    requestAction,
    close,
    handleSuccess,
  };
}
