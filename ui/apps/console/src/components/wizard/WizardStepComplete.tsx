import { CheckIcon, CpuChipIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace } from "@/hooks/useNamespaces";
import { useDevices } from "@/hooks/useDevices";

interface WizardStepCompleteProps {
  device: { uid: string; name: string } | null;
}

export default function WizardStepComplete({
  device,
}: WizardStepCompleteProps) {
  const tenantId = useAuthStore((s) => s.tenant) ?? "";
  const { namespace: ns } = useNamespace(tenantId);
  const namespace = ns?.name;

  // Pull the accepted device's details (OS) for the card.
  const { devices } = useDevices({ perPage: 10, enabled: !!device });
  const os =
    devices.find((d) => d.uid === device?.uid)?.info?.pretty_name ??
    "Linux device";

  return (
    <div className="py-2 flex flex-col gap-6">
      <div className="flex flex-col items-center text-center gap-3 pt-2">
        <div className="relative">
          <div className="absolute inset-0 rounded-full bg-accent-green/20 blur-xl" />
          <div className="relative w-14 h-14 rounded-2xl bg-accent-green/15 border border-accent-green/30 flex items-center justify-center">
            <CheckIcon className="w-7 h-7 text-accent-green" strokeWidth={2} />
          </div>
        </div>
        <div>
          <h2 className="text-lg font-semibold text-text-primary">
            Your first device is connected
          </h2>
          <p className="text-sm text-text-muted mt-1">
            It&rsquo;s in{" "}
            {namespace ? (
              <>
                <span className="text-text-secondary font-medium">
                  {namespace}
                </span>{" "}
                namespace
              </>
            ) : (
              "your namespace"
            )}{" "}
            and ready to use.
          </p>
        </div>
      </div>

      {device && (
        <div className="flex items-center gap-3.5 bg-card border border-border rounded-2xl p-4">
          <div className="w-11 h-11 rounded-xl bg-surface border border-border-light flex items-center justify-center text-text-secondary shrink-0">
            <CpuChipIcon className="w-6 h-6" strokeWidth={1.5} />
          </div>
          <div className="flex-1 min-w-0">
            <p className="font-mono text-sm font-semibold text-text-primary truncate">
              {device.name}
            </p>
            <p className="text-xs text-text-muted truncate">{os}</p>
          </div>
          <span className="inline-flex items-center gap-2 font-mono text-2xs text-accent-green bg-accent-green/10 border border-accent-green/25 rounded-full px-2.5 py-1">
            <span className="relative flex h-2 w-2">
              <span className="absolute inline-flex h-full w-full rounded-full bg-accent-green opacity-50 animate-ping" />
              <span className="relative inline-flex h-2 w-2 rounded-full bg-accent-green" />
            </span>
            online
          </span>
        </div>
      )}
    </div>
  );
}
