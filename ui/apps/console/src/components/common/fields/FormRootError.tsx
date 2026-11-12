import { ExclamationCircleIcon } from "@heroicons/react/24/outline";

/**
 * A form-level error announced as an alert: the `root` error a react-hook-form
 * form sets, or the save error a form holds in its own state. Renders nothing
 * while the message is empty, null or undefined.
 */
export default function FormRootError({
  message,
}: {
  message?: string | null;
}) {
  if (!message) return null;

  return (
    <p
      role="alert"
      className="text-xs text-accent-red flex items-center gap-1.5"
    >
      <ExclamationCircleIcon className="w-3.5 h-3.5 shrink-0" strokeWidth={2} />
      {message}
    </p>
  );
}
