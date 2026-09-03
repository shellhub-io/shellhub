import { useState } from "react";
import type { GetStatusDevicesResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { useGetStatusDevices } from "@/client/api";
import { hasAcceptedDevices } from "@/utils/stats";
import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizard from "./WelcomeWizard";
import { isWizardDemo } from "./demo";

/**
 * Mounts the WelcomeWizard automatically when:
 *   - The current tenant has never dismissed the wizard for good
 *   - The namespace has no accepted device yet
 *
 * Onboarding is "done" only once a device is accepted, not when any device
 * appears, so a device left pending (the user ran the command but closed
 * before accepting) still reopens the wizard.
 *
 * Rendered inside AppLayout so it works regardless of which page the user
 * lands on. Closing merely defers it (reappears next visit); only an explicit
 * skip or finishing marks the tenant as "seen".
 *
 * Eligibility is decided ONCE, from the device state at page load: this outer
 * component waits for stats+tenant to resolve, then mounts the gate, which
 * freezes the decision in a lazy initializer. Deleting the last device
 * mid-session therefore can't pop the wizard back open in the user's face;
 * only a fresh page load reconsiders.
 */
export default function WelcomeWizardTrigger() {
  const tenant = useAuthStore((s) => s.tenant);
  const { data: stats, refetch } = useGetStatusDevices();

  if (!stats || !tenant) return null;

  return (
    <WelcomeWizardGate
      tenant={tenant}
      initialStats={stats}
      refetch={() => void refetch()}
    />
  );
}

interface WelcomeWizardGateProps {
  tenant: string;
  initialStats: GetStatusDevicesResponse;
  refetch: () => void;
}

function WelcomeWizardGate({
  tenant,
  initialStats,
  refetch,
}: WelcomeWizardGateProps) {
  const [dismissed, setDismissed] = useState(false);
  const [eligible] = useState(
    () => !hasSeenWelcome(tenant) && !hasAcceptedDevices(initialStats),
  );

  const show = !dismissed && (isWizardDemo() || eligible);

  const handleClose = () => {
    setDismissed(true);
    void refetch();
  };

  const handleDismiss = () => {
    markWelcomeSeen(tenant);
    setDismissed(true);
    void refetch();
  };

  return (
    <WelcomeWizard
      open={show}
      onClose={handleClose}
      onDismiss={handleDismiss}
    />
  );
}
