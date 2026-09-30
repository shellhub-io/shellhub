import { Button } from "@shellhub/design-system/primitives";
import type { ResolveDeviceLoginCodeResponse } from "@/client";
import DeviceCard from "./DeviceCard";

interface PairStepProps {
  device: ResolveDeviceLoginCodeResponse;
  code: string;
  namespace: string;
  isPending: boolean;
  error: string;
  onPair: () => void;
  onBack: () => void;
}

/**
 * What the code resolved to, shown before anything is paired, so the user checks it is the
 * device they just installed.
 */
export default function PairStep({
  device,
  code,
  namespace,
  isPending,
  error,
  onPair,
  onBack,
}: PairStepProps) {
  return (
    <div className="flex flex-col gap-4">
      <DeviceCard
        device={device}
        code={code}
        status={
          <span className="text-2xs font-mono text-text-muted">
            waiting for you to pair
          </span>
        }
      />
      {error && (
        <p role="alert" className="text-xs text-accent-red">
          {error}
        </p>
      )}
      <div className="flex flex-wrap items-center gap-2.5">
        <Button onClick={onPair} loading={isPending} disabled={isPending}>
          Pair into {namespace}
        </Button>
        <Button variant="secondary" onClick={onBack} disabled={isPending}>
          Not this one
        </Button>
      </div>
    </div>
  );
}
