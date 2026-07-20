import { useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import { useDevices } from "@/hooks/useDevices";

const POLL_INTERVAL_MS = 3000;

interface WizardAcceptedWatcherProps {
  onConnected: (device: { uid: string; name: string }) => void;
}

/**
 * Headless poller for the link path: while the wizard is on step 1 (install or
 * code face), it watches for the first accepted device and reports it. Living
 * above both faces means accepting via the printed link still advances the
 * wizard even after the user switched to the code-entry face.
 *
 * The wizard only runs while the namespace has no accepted device, so the first
 * accepted device to appear is the one the user just enrolled.
 */
export default function WizardAcceptedWatcher({
  onConnected,
}: WizardAcceptedWatcherProps) {
  const tenant = useAuthStore((s) => s.tenant);
  const { devices, refetch } = useDevices({ perPage: 10, enabled: !!tenant });

  useEffect(() => {
    const id = window.setInterval(() => void refetch(), POLL_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [refetch]);

  const accepted = devices.find((d) => d.status === "accepted");
  useEffect(() => {
    if (accepted?.uid) {
      onConnected({ uid: accepted.uid, name: accepted.name });
    }
  }, [accepted, onConnected]);

  return null;
}
