import { useMutation, useQuery } from "@tanstack/react-query";
import {
  choiceDevicesMutation,
  getDevicesMostUsedOptions,
} from "../client";
import { useInvalidateByIds } from "./useInvalidateQueries";
import { normalizeDeviceTags, type TaggedDevice } from "@/utils/deviceTags";

/**
 * The devices used most, offered first when a namespace over its limit has to choose which to
 * keep. Tags are normalized, so callers get plain strings.
 */
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

/**
 * Commits the chosen devices. Everything not chosen loses its place, so this refreshes the
 * device queries and the counts together.
 */
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
