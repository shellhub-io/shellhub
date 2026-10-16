import { useState } from "react";
import { Link } from "react-router-dom";
import BaseDialog from "@/components/common/BaseDialog";
import AcceptDeviceFlow from "@/components/devices/AcceptDeviceFlow";
import { useNamespaceLegacy } from "@/hooks/useNamespaceLegacy";
import {
  METHODS,
  PAIRING_METHODS,
  methodsForNewNamespace,
  type Method,
} from "@/pages/install/methods";
import { useAuthStore } from "@/stores/authStore";
import InstallCommand from "./InstallCommand";
import MethodCards from "./MethodCards";
import Step from "./Step";

/**
 * Adding a device interactively, with someone at the machine. In a new namespace the command
 * carries no credential where the installer can pair, and the agent prints a link to accept it in
 * the browser; snap and FreeBSD are left out, since neither can pair. In a legacy namespace every command carries
 * the tenant and the device waits in the pending queue to be accepted.
 */
export default function Interactive() {
  const { tenant } = useAuthStore();
  const legacy = useNamespaceLegacy() === true;
  const [method, setMethod] = useState<Method>("auto");
  const [pairOpen, setPairOpen] = useState(false);
  const methods = legacy ? METHODS : methodsForNewNamespace();
  const selected = methods.some((m) => m.id === method) ? method : "auto";
  const pairs = !legacy && PAIRING_METHODS.includes(selected);
  const manual = METHODS.find((m) => m.id === selected)?.manual;

  return (
    <div className="space-y-6">
      <Step
        n={1}
        labelId="interactive-method-label"
        label="Installation method"
      >
        <MethodCards
          labelledBy="interactive-method-label"
          value={selected}
          onChange={setMethod}
          methods={methods}
        />
      </Step>

      <Step
        n={2}
        label={manual ? "Follow the documentation" : "Run on your device"}
      >
        <InstallCommand
          method={selected}
          credential={legacy ? `TENANT_ID=${tenant}` : undefined}
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
            ) : legacy ? (
              <>
                The device registers with the tenant ID and waits in the{" "}
                <Link
                  to="/devices/pending"
                  className="text-primary font-medium hover:text-primary/80 transition-colors"
                >
                  pending queue
                </Link>{" "}
                until you accept it.
              </>
            ) : (
              <>
                The guide enrolls the device by pairing or with a provisioning
                key from the{" "}
                <Link
                  to="/devices/add/fleet"
                  className="text-primary font-medium hover:text-primary/80 transition-colors"
                >
                  Fleet tab
                </Link>
                .
              </>
            )
          }
        />
      </Step>

      {!legacy && (
        <BaseDialog
          open={pairOpen}
          onClose={() => setPairOpen(false)}
          size="md"
          className="p-6"
          aria-label="Claim a device"
        >
          <AcceptDeviceFlow inDialog />
        </BaseDialog>
      )}
    </div>
  );
}
