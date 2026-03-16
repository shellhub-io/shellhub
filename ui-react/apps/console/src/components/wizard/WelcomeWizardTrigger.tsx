import { useState } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useStats } from "@/hooks/useStats";
import { hasAnyDevices } from "@/utils/stats";
import { hasSeenWelcome, markWelcomeSeen } from "@/utils/welcomeState";
import WelcomeWizard from "./WelcomeWizard";

/**
 * Mounts the WelcomeWizard automatically when:
 *   - The current tenant has never seen the wizard
 *   - The namespace has zero devices of any status
 *
 * Rendered inside AppLayout so it works regardless of which page the user
 * lands on. The tenant is marked as "seen" when the user closes the wizard,
 * not when it opens — so closing early doesn't permanently suppress it until
 * the user consciously dismisses it.
 *
 * Uses the shared stats query so that when the wizard completes and is closed,
 * a stats refresh causes the Dashboard to re-render with up-to-date data
 * instead of lingering on the WelcomeScreen empty state.
 */
export default function WelcomeWizardTrigger() {
  const tenant = useAuthStore((s) => s.tenant);
  const { stats, refetch } = useStats();
  const [dismissed, setDismissed] = useState(false);

  // Derive open state during render — no setState-in-effect needed
  const show
    = !!stats
      && !!tenant
      && !dismissed
      && !hasSeenWelcome(tenant)
      && !hasAnyDevices(stats);

  const handleClose = () => {
    if (tenant) markWelcomeSeen(tenant);
    setDismissed(true);
    // Refresh the shared query so the Dashboard re-renders with current data
    void refetch();
  };

  return <WelcomeWizard open={show} onClose={handleClose} />;
}
