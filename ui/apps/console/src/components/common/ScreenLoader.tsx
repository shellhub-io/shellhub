import { Spinner } from "@shellhub/design-system/primitives";

/**
 * The whole-screen wait a guard shows before the app can render, with the label beside the
 * spinner.
 */
export default function ScreenLoader({ label }: { label: string }) {
  return (
    <div
      role="status"
      className="min-h-screen flex items-center justify-center bg-background"
    >
      <div className="flex items-center gap-3">
        <Spinner />
        <span className="text-xs font-mono text-text-muted">{label}</span>
      </div>
    </div>
  );
}
