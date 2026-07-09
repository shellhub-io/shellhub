import { ExclamationCircleIcon } from "@heroicons/react/24/outline";

/**
 * Canonical rendering for a form-level (`root`) error surfaced via RHF's
 * `setError("root", …)`. Shared by every form so server-error presentation
 * stays identical across drawers.
 */
export default function FormRootError({ message }: { message?: string }) {
  if (!message) return null;

  return (
    <p
      role="alert"
      className="text-xs font-mono text-accent-red flex items-center gap-1.5"
    >
      <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
      {message}
    </p>
  );
}
