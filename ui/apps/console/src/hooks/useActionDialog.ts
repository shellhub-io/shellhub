import { useState, useCallback, useRef, useEffect } from "react";

export interface EntityBase {
  uid: string;
  name: string;
}

export type EntityOperation = "accept" | "reject" | "remove";

export interface Action {
  entity: EntityBase;
  operation: EntityOperation;
}

interface UseActionDialogOptions {
  onSuccess?: (operation: EntityOperation) => void;
}

export interface UseActionDialogResult {
  action: Action | undefined;
  actionKey: string;
  requestAction: (entity: EntityBase, operation: EntityOperation) => void;
  close: () => void;
  handleSuccess: (operation: EntityOperation) => void;
}

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
