import PairingCodeForm from "@/components/common/PairingCodeForm";

interface PairingCodeFieldProps {
  onSubmit: (code: string) => void;
  isPending: boolean;
  error: string;
}

/**
 * The code the agent prints, entered the way the accept-device page takes it, with the trail's
 * error and hint around it. Pasting the printed link works too.
 */
export default function PairingCodeField({
  onSubmit,
  isPending,
  error,
}: PairingCodeFieldProps) {
  return (
    <div>
      <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-3">
        Pairing code
      </p>
      <PairingCodeForm onSubmit={onSubmit} pending={isPending} />
      {error && (
        <p role="alert" className="mt-3 text-xs text-accent-red">
          {error}
        </p>
      )}
      <p className="mt-3 text-xs text-text-muted">
        Paste the code or the whole link. Opening the link on this machine works
        too; this page notices the device either way.
      </p>
    </div>
  );
}
