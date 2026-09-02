import { useQuery } from "@tanstack/react-query";
import { getDevicesMostUsedOptions } from "@/client/api";
import { normalizeDeviceTags, type TaggedDevice } from "@/utils/deviceTags";

/**
 *
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
