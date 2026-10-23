import { useMutation, useQuery } from "@tanstack/react-query";
import { configureTelemetryMutation, getTelemetryOptions } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * The instance's telemetry: the administrator's setting, whether the environment forces it off
 * (SHELLHUB_TELEMETRY=false, or a server with no telemetry client), and the events the next
 * heartbeat sends. status is undefined until it loads.
 */
export function useAdminTelemetry() {
  const result = useQuery({
    ...getTelemetryOptions(),
    refetchOnWindowFocus: false,
  });

  const status = result.data && {
    enabled: result.data.enabled,
    disabledByEnvironment: result.data.disabled_by_environment,
    preview: result.data.preview,
  };

  return {
    status,
    isLoading: result.isLoading,
    isError: result.isError,
  };
}

/**
 * Writes the instance's telemetry setting, refreshing the setting and its preview.
 */
export function useAdminTelemetryMutations() {
  const invalidate = useInvalidateByIds("getTelemetry");
  const configure = useMutation({
    ...configureTelemetryMutation(),
    onSuccess: invalidate,
  });

  return {
    setEnabled: (enabled: boolean) => configure.mutate({ body: { enabled } }),
    isPending: configure.isPending,
    isError: configure.isError,
  };
}
