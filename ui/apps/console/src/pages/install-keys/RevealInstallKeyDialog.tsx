import { useId, useState } from "react";
import {
  ExclamationCircleIcon,
  ExclamationTriangleIcon,
  EyeIcon,
  LockClosedIcon,
} from "@heroicons/react/24/outline";
import { Button, Card, Spinner } from "@shellhub/design-system/primitives";
import { useRevealInstallKey } from "@/hooks/useRevealInstallKey";
import { type InstallKey } from "@/client";
import { installKeyDisplayName } from "./helpers";
import CopyButton from "@/components/common/CopyButton";
import BaseDialog from "@/components/common/BaseDialog";
import RestrictedAction from "@/components/common/RestrictedAction";
import { LABEL } from "@/utils/styles";

/**
 * Shows an install key's identity and secret. The fingerprint (non-secret, stable) is shown first;
 * the secret sits below behind a deliberate reveal — it fetches nothing until clicked (see
 * useRevealInstallKey) and is dropped from cache on close so it doesn't linger. The legacy/system key
 * has no secret, so it shows only its fingerprint.
 */
export default function RevealInstallKeyDialog({
  installKey,
  onClose,
}: {
  installKey: InstallKey | null;
  onClose: () => void;
}) {
  const autoId = useId();
  const titleId = `reveal-install-key-title-${autoId}`;
  const fingerprintLabelId = `reveal-install-key-fingerprint-${autoId}`;
  const secretLabelId = `reveal-install-key-secret-${autoId}`;

  const name = installKey?.name ?? null;
  const displayName = installKey ? installKeyDisplayName(installKey) : name;
  const hasSecret = !!installKey?.key_hint;
  const fingerprint = installKey?.id ?? "";

  const [revealed, setRevealed] = useState(false);
  // Fall back to the locked state whenever a different key is targeted.
  const [prevName, setPrevName] = useState(name);
  if (name !== prevName) {
    setPrevName(name);
    setRevealed(false);
  }

  const { key, isLoading, error } = useRevealInstallKey(
    hasSecret ? name : null,
    revealed,
  );

  return (
    <BaseDialog
      open={!!name}
      onClose={onClose}
      size="md"
      aria-labelledby={titleId}
    >
      <div className="p-6 pb-0">
        <h2 id={titleId} className="text-base font-semibold text-text-primary">
          {displayName}
        </h2>
      </div>

      <div className="px-6 pt-2 pb-6 space-y-4">
        <p className="text-sm text-text-muted">
          {hasSecret
            ? "The full key and its fingerprint."
            : "No secret to reveal — devices register with the tenant ID alone. Its fingerprint still identifies the key."}
        </p>

        {/* Non-secret identity, shown first: stable, safe to share. */}
        <div>
          <span id={fingerprintLabelId} className={LABEL}>
            Fingerprint
          </span>
          <Card
            aria-labelledby={fingerprintLabelId}
            className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
          >
            <code
              className="flex-1 truncate text-xs font-mono text-text-muted select-all"
              title={fingerprint}
            >
              {fingerprint}
            </code>
            <CopyButton text={fingerprint} size="md" />
          </Card>
          <p className="mt-1.5 text-2xs text-text-muted">
            Safe to share. The registration webhook uses it to tell your
            endpoint which key a device registered with.
          </p>
        </div>

        {/* The secret, behind a deliberate reveal. The amber warning line below is what marks it
            sensitive — never a red/error tint. */}
        {hasSecret && (
          <div>
            <span id={secretLabelId} className={LABEL}>
              Install Key
            </span>
            {!revealed ? (
              <div className="flex flex-col items-center gap-2.5 text-center border border-dashed border-border-light rounded-lg px-5 py-6">
                <LockClosedIcon
                  className="w-7 h-7 text-text-muted"
                  strokeWidth={1.5}
                />
                <RestrictedAction action="installKey:reveal">
                  <Button
                    variant="primary"
                    size="sm"
                    icon={<EyeIcon className="w-4 h-4" strokeWidth={2} />}
                    onClick={() => setRevealed(true)}
                  >
                    Reveal key
                  </Button>
                </RestrictedAction>
              </div>
            ) : isLoading ? (
              <div className="flex justify-center py-8">
                <Spinner />
              </div>
            ) : error ? (
              <div
                role="alert"
                className="flex items-start gap-2 bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5 text-xs text-accent-red"
              >
                <ExclamationCircleIcon
                  className="w-4 h-4 shrink-0 mt-px"
                  strokeWidth={2}
                />
                <span>
                  Could not load the key. Check your connection and try again.
                </span>
              </div>
            ) : (
              <>
                <Card
                  aria-labelledby={secretLabelId}
                  className="rounded-lg px-3.5 py-2.5 flex items-center gap-2"
                >
                  <code
                    className="flex-1 truncate text-xs font-mono text-text-muted select-all"
                    title={key ?? ""}
                  >
                    {key}
                  </code>
                  <CopyButton text={key ?? ""} size="md" />
                </Card>
                <p className="mt-1.5 flex items-start gap-1 text-2xs text-accent-yellow">
                  <ExclamationTriangleIcon
                    className="w-3.5 h-3.5 shrink-0 mt-px"
                    strokeWidth={2}
                  />
                  <span>
                    Treat like a password. Anyone with it can register devices
                    with your namespace.
                  </span>
                </p>
              </>
            )}
          </div>
        )}
      </div>

      <div className="flex justify-end gap-2 px-6 py-4 border-t border-border">
        <Button variant="ghost" onClick={onClose}>
          Close
        </Button>
      </div>
    </BaseDialog>
  );
}
