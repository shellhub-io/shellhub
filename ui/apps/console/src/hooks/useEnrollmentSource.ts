import { useCallback } from "react";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import {
  resolveEnrollmentSource,
  type EnrollmentSource,
} from "@/pages/install-keys/helpers";

const INSTALL_KEYS_PER_PAGE = 100;

/**
 * A resolver from a device's install key digest to the source that enrolled it, for a caller
 * resolving more than one device: it subscribes to the namespace's keys once, where calling
 * useEnrollmentSource per row would subscribe once per row.
 *
 * The lookup is client-side over a single page of keys, so a namespace with more than
 * INSTALL_KEYS_PER_PAGE of them resolves its tail to null rather than to a name.
 */
export function useEnrollmentSourceResolver(): (
  installKeyId: string | undefined,
) => EnrollmentSource | null {
  const { installKeys } = useInstallKeys({ perPage: INSTALL_KEYS_PER_PAGE });

  return useCallback(
    (installKeyId: string | undefined) =>
      resolveEnrollmentSource(installKeyId, installKeys),
    [installKeys],
  );
}

/** The same resolution for a single device. Returns null when the key is not resolvable. */
export function useEnrollmentSource(
  installKeyId: string | undefined,
): EnrollmentSource | null {
  return useEnrollmentSourceResolver()(installKeyId);
}
