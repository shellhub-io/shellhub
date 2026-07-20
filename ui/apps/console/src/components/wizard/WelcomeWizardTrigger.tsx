import { useState } from "react";
import type { GetStatusDevicesResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { useStats } from "@/hooks/useStats";
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
  const { stats, refetch } = useStats();

  // Hold off until stats+tenant resolve, so the gate's lazy initializer below
  // captures the actual page-load device state rather than a loading null.
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
  /** Device stats snapshot taken when the gate first mounts (page load). */
  initialStats: GetStatusDevicesResponse;
  refetch: () => void;
}

/**
 * Owns the open/dismiss state. The eligibility decision is frozen at mount via
 * a lazy useState initializer, so later stats changes never reopen the wizard.
 *
 * Refetches stats when closed so the Dashboard re-renders with up-to-date data
 * instead of lingering on the WelcomeScreen empty state.
 */
function WelcomeWizardGate({
  tenant,
  initialStats,
  refetch,
}: WelcomeWizardGateProps) {
  const [dismissed, setDismissed] = useState(false);
  const [eligible] = useState(
    () => !hasSeenWelcome(tenant) && !hasAcceptedDevices(initialStats),
  );

  // `?wizard=demo` (dev only) forces the wizard open, bypassing the gate, but
  // dismissing it still closes it for the session.
  const show = !dismissed && (isWizardDemo() || eligible);

  // Close for now: hide it this session only, so it reappears next visit while
  // the namespace is still empty.
  const handleClose = () => {
    setDismissed(true);
    void refetch();
  };

  // Dismiss for good: the user skipped explicitly or finished onboarding, so
  // never show it for this tenant again.
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
