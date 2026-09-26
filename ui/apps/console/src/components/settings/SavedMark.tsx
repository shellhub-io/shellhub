import { CheckIcon } from "@heroicons/react/24/outline";

/**
 * The brief "Saved" beside a setting's title after a change is stored. It is a status, so a
 * screen reader announces it without taking focus.
 */
export default function SavedMark() {
  return (
    <span
      role="status"
      className="inline-flex items-center gap-1 text-2xs font-normal text-accent-green animate-fade-in"
    >
      <CheckIcon aria-hidden="true" className="w-3 h-3" strokeWidth={2.5} />
      Saved
    </span>
  );
}
