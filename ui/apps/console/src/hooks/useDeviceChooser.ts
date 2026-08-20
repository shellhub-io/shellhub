import { useMutation, useQuery } from "@tanstack/react-query";
import {
  choiceDevicesMutation,
  getDevicesMostUsedOptions,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";
import { normalizeDeviceTags, type TaggedDevice } from "@/utils/deviceTags";

export function useSuggestedDevices(enabled = true) {
  const result = useQuery({
    ...getDevicesMostUsedOptions(),
    enabled,
  });
  const devices: TaggedDevice[] = (result.data ?? []).map(normalizeDeviceTags);
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
