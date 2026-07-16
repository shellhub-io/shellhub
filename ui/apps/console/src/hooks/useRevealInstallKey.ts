import { useQuery } from "@tanstack/react-query";
import { installKeyRevealOptions } from "../client/@tanstack/react-query.gen";

/**
 * Reveal a install key's plaintext on demand. The secret is never preloaded for
 * the list rows, and not even fetched when the dialog opens: the query only
 * fires once a key is targeted (`name` set) AND the user opts in (`enabled`), so
 * the plaintext is decrypted only on an explicit click. The result is dropped
 * from cache as soon as the dialog closes so the decrypted value doesn't linger.
 */
export function useRevealInstallKey(name: string | null, enabled = true) {
  const result = useQuery({
    ...installKeyRevealOptions({ path: { key: name ?? "" } }),
    enabled: !!name && enabled,
    gcTime: 0,
  });

  return {
    key: result.data?.key ?? "",
    isLoading: result.isLoading,
    error: result.error,
  };
}
