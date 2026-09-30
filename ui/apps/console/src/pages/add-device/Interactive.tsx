import { useState } from "react";
import { Link } from "react-router-dom";
import BaseDialog from "@/components/common/BaseDialog";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";
import { METHODS, PAIRING_METHODS, type Method } from "@/pages/install/methods";
import { useAuthStore } from "@/stores/authStore";
import InstallCommand from "./InstallCommand";
import MethodCards from "./MethodCards";
import Step from "./Step";

/**
 * Adding a device interactively, with someone at the machine. Where the installer can pair, the command carries no credential and
 * the agent prints a link to accept it in the browser; elsewhere it carries the tenant and the
 * device waits to be accepted.
 */
export default function Interactive() {
  const { tenant } = useAuthStore();
  const [method, setMethod] = useState<Method>("auto");
  const [pairOpen, setPairOpen] = useState(false);
  const pairs = PAIRING_METHODS.includes(method);
  const manual = METHODS.find((m) => m.id === method)?.manual;

  return (
    <div className="space-y-6">
      <Step
        n={1}
        labelId="interactive-method-label"
        label="Installation method"
      >
        <MethodCards
          labelledBy="interactive-method-label"
          value={method}
          onChange={setMethod}
        />
      </Step>

      <Step
        n={2}
        label={manual ? "Follow the documentation" : "Run on your device"}
      >
        <InstallCommand
          method={method}
          credential={pairs ? undefined : `TENANT_ID=${tenant}`}
          outcome={
            pairs ? (
              <>
                The agent prints a link. Open it to accept this device, and
                it&apos;s ready to use. Away from that machine? Enter{" "}
                <button
                  type="button"
                  onClick={() => setPairOpen(true)}
                  className="text-primary font-medium hover:text-primary/80 transition-colors"
                >
                  the code it shows
                </button>{" "}
                instead.
              </>
            ) : (
              <>
                The device registers through the tenant-only{" "}
                <Link
                  to="/devices/add/fleet"
                  className="text-primary font-medium hover:text-primary/80 transition-colors"
                >
                  provisioning key
                </Link>{" "}
                and waits there until you accept it.
              </>
            )
          }
        />
      </Step>

      <BaseDialog
        open={pairOpen}
        onClose={() => setPairOpen(false)}
        size="md"
        className="p-6"
        aria-label="Claim a device"
      >
        <AcceptDeviceFlow inDialog />
      </BaseDialog>
    </div>
  );
}
