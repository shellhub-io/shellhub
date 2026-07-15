import { isCloud } from "@/env";
import {
  useActionDialogState,
  type UseActionDialogStateOptions,
  type UseActionDialogStateResult,
} from "./useActionDialogState";

export type UseDeviceActionsResult = UseActionDialogStateResult;

/**
 * Thin wrapper over useActionDialogState for device list/detail actions.
 *
 * Billing warning defaults to the cloud flag; callers may override it
 * (e.g. ContainerDetails disables it to preserve its inline-error flow).
 */
export function useDeviceActions(
  options?: Partial<
    Pick<UseActionDialogStateOptions, "enableBillingWarning" | "onSuccess">
  >,
): UseDeviceActionsResult {
  return useActionDialogState({
    ...options,
    enableBillingWarning: options?.enableBillingWarning ?? isCloud(),
  });
}
