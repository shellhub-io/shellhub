import { useInstallKeys } from "@/hooks/useInstallKeys";
import {
  resolveEnrollmentSource,
  type EnrollmentSource,
} from "@/pages/install-keys/helpers";

const INSTALL_KEYS_PER_PAGE = 100;

/**
 * Resolves a device's install key digest to the source that enrolled it, returning null when the
 * device carries no digest or the key is not among the namespace's first INSTALL_KEYS_PER_PAGE
 * keys — the lookup is client-side over a single page, so a larger namespace resolves its tail to
 * null rather than to a name.
 */
export function useEnrollmentSource(
  installKeyId: string | undefined,
): EnrollmentSource | null {
  const { installKeys } = useInstallKeys({ perPage: INSTALL_KEYS_PER_PAGE });

  return resolveEnrollmentSource(installKeyId, installKeys);
}
