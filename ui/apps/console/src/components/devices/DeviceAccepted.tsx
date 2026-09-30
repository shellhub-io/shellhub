import type { ComponentProps, ReactNode } from "react";
import { Button } from "@shellhub/design-system/primitives";
import { ArrowRightIcon } from "@heroicons/react/24/outline";
import DeviceCard from "@/components/firstRun/DeviceCard";
import ScreenIntro from "@/components/layout/ScreenIntro";
import AuthActions from "@/components/auth/AuthActions";

interface DeviceAcceptedProps {
  device: ComponentProps<typeof DeviceCard>["device"];
  namespace?: string;
  note?: ReactNode;
  onViewDevice?: () => void;
  viewing?: boolean;
}

/**
 * The screen after a device is accepted: the device, where it went, and the way into its page.
 * The agent connects on its own, so there is nothing to wait for here.
 */
export default function DeviceAccepted({
  device,
  namespace,
  note,
  onViewDevice,
  viewing = false,
}: DeviceAcceptedProps) {
  return (
    <div className="space-y-5">
      <ScreenIntro
        eyebrow="Pairing"
        title="Device accepted"
        lead={
          <>
            {namespace ? (
              <>
                It's in{" "}
                <span className="font-medium text-text-primary">
                  {namespace}
                </span>{" "}
                now, and the agent connects on its own.
              </>
            ) : (
              "The agent connects on its own."
            )}{" "}
            You can close the terminal that showed the code.
            {note ? <> {note}</> : null}
          </>
        }
      />

      <DeviceCard
        device={device}
        status={
          <span className="inline-flex items-center gap-1.5 text-2xs font-mono text-accent-green">
            <span className="w-1.5 h-1.5 rounded-full bg-accent-green" />
            accepted
          </span>
        }
      />

      {onViewDevice && (
        <AuthActions
          primary={
            <Button
              size="lg"
              fullWidth
              loading={viewing}
              iconRight={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
              onClick={onViewDevice}
            >
              View device
            </Button>
          }
        />
      )}
    </div>
  );
}
