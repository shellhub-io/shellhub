import { CheckIcon } from "@heroicons/react/24/outline";
import CopyableValue from "@/components/common/CopyableValue";

/**
 * The confirmation shown once a secret has been created, followed by the secret itself. Renders
 * siblings, so the caller's wrapper sets the spacing.
 */
export default function SecretReveal({
  title,
  note,
  label,
  value,
}: {
  title: string;
  note: string;
  label: string;
  value: string;
}) {
  return (
    <>
      <div className="flex items-start gap-3 bg-accent-green/[0.06] border border-accent-green/20 rounded-xl px-4 py-3.5">
        <CheckIcon className="w-5 h-5 text-accent-green shrink-0 mt-0.5" />
        <div>
          <p className="text-sm font-medium text-text-primary">{title}</p>
          <p className="text-2xs text-text-muted mt-0.5">{note}</p>
        </div>
      </div>
      <CopyableValue label={label} value={value} />
    </>
  );
}
