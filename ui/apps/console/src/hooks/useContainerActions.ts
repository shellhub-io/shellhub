import { isCloud } from "@/env";
import {
  useActionDialogState,
  type UseActionDialogStateOptions,
  type UseActionDialogStateResult,
} from "./useActionDialogState";

export type UseContainerActionsResult = UseActionDialogStateResult;

/**
 * Thin wrapper over useActionDialogState for container list/detail actions.
 *
 * Billing warning defaults to the cloud flag; callers may override it
 * (e.g. ContainerDetails disables it to preserve its inline-error flow).
 */
export function useContainerActions(
  options?: Partial<
    Pick<UseActionDialogStateOptions, "enableBillingWarning" | "onSuccess">
  >,
): UseContainerActionsResult {
  return useActionDialogState({
    ...options,
    enableBillingWarning: options?.enableBillingWarning ?? isCloud(),
  });
}
