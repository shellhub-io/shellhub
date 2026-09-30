import { useState } from "react";
import { Outlet, useLocation } from "react-router-dom";
import { useFirstRunEligible } from "@/hooks/useFirstRunEligible";
import DeviceTrail from "./DeviceTrail";

/**
 * Takes /dashboard over with the first-run trail on an account's first run: a member who can
 * accept devices, whose only namespace has no accepted device. A user with several namespaces is
 * past their first run and keeps the normal dashboard. Every other route passes through.
 *
 * Once the trail shows, it stays for the rest of that visit to the dashboard, even after pairing
 * makes the namespace no longer empty, so the user sees the last step. Leaving the route or
 * reloading ends it.
 */
export default function FirstRunGate() {
  const { pathname, key } = useLocation();
  const eligibility = useFirstRunEligible();
  const [heldFor, setHeldFor] = useState<string | null>(null);

  const onDashboard = pathname === "/dashboard";
  const eligible = onDashboard && eligibility === true;

  if (eligible && heldFor !== key) setHeldFor(key);

  if (onDashboard && eligibility === "loading") return null;
  if (onDashboard && (eligible || heldFor === key)) return <DeviceTrail />;
  return <Outlet />;
}
