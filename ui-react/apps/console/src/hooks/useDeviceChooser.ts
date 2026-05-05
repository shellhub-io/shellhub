import { useMutation, useQuery } from "@tanstack/react-query";
import {
  choiceDevicesMutation,
  getDevicesMostUsedOptions,
} from "../client/@tanstack/react-query.gen";
import { useInvalidateByIds } from "./useInvalidateQueries";
import { normalizeDevice, type NormalizedDevice } from "./useDevices";

export function useSuggestedDevices(enabled = true) {
  const result = useQuery({
    ...getDevicesMostUsedOptions(),
    enabled,
  });
  const devices: NormalizedDevice[] = (result.data ?? []).map(normalizeDevice);
  return {
    devices,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

export function useChoiceDevices() {
  const invalidate = useInvalidateByIds(
    "getDevices",
    "getDevice",
    "getStatusDevices",
  );
  return useMutation({
    ...choiceDevicesMutation(),
    onSuccess: invalidate,
  });
}
