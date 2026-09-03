import { useMutation } from "@tanstack/react-query";
import { createBillingPortalSession } from "@/client/api";

/**
 *
 */
export function useOpenBillingPortal() {
  return useMutation({
    mutationFn: async () => {
      const { data } = await createBillingPortalSession({ throwOnError: true });
      if (!data.url) throw new Error("Missing billing portal URL");
      window.open(data.url, "_blank", "noopener,noreferrer");
      return data.url;
    },
  });
}
