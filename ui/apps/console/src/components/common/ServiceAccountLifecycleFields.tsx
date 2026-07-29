import { BoltIcon } from "@heroicons/react/24/outline";
import { Toggle } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import KeyExpiryField from "@/components/common/KeyExpiryField";

/**
 * The service-account key lifecycle controls, shared by the two places a
 * service account is created. Expiration is the same field a person's key
 * gets; single-use is the part that stays service-account only, because on a
 * person's key it would lock them out after one login.
 *
 * expiresIn is an EXPIRY_OPTIONS value as a string ("-1" = never).
 */
export default function ServiceAccountLifecycleFields({
  expiresIn,
  onExpiresInChange,
  singleUse,
  onSingleUseChange,
}: {
  expiresIn: string;
  onExpiresInChange: (value: string) => void;
  singleUse: boolean;
  onSingleUseChange: (value: boolean) => void;
}) {
  return (
    <>
      <KeyExpiryField
        expiresIn={expiresIn}
        onExpiresInChange={onExpiresInChange}
      />

      <div
        className={cn(
          "flex items-center gap-3 px-3.5 py-3 border rounded-xl transition-colors",
          singleUse
            ? "border-primary/40 bg-primary/[0.06]"
            : "border-border bg-card",
        )}
      >
        <span
          className={cn(
            "grid place-items-center w-8 h-8 rounded-lg bg-surface shrink-0",
            singleUse ? "text-primary" : "text-text-secondary",
          )}
        >
          <BoltIcon className="w-4 h-4" strokeWidth={2} />
        </span>
        <span className="min-w-0">
          <span className="block text-sm font-semibold text-text-primary">
            Single-use key
          </span>
          <span className="block text-xs text-text-muted">
            The key is spent by its first established SSH session, then stops
            working. Good for a one-shot deploy or backup pull.
          </span>
        </span>
        <Toggle
          enabled={singleUse}
          onChange={onSingleUseChange}
          aria-label="Single-use key"
          className="ml-auto shrink-0"
        />
      </div>
    </>
  );
}
