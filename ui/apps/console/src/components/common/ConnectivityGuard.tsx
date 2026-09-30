import { useEffect } from "react";
import { Outlet } from "react-router-dom";
import { Spinner } from "@shellhub/design-system/primitives";
import { useConnectivityStore } from "@/stores/connectivityStore";
import FramedShell from "@/components/layout/FramedShell";
import ScreenIntro from "@/components/layout/ScreenIntro";

function ApiUnavailablePage() {
  return (
    <FramedShell>
      <ScreenIntro
        eyebrow="Connection issue"
        title="Waiting for the API"
        lead="The ShellHub API is not responding. This is likely temporary, and the app resumes on its own once the connection is back."
      />
      <div className="flex items-center gap-2.5 text-xs font-mono text-text-secondary">
        <Spinner size="xs" tone="subtle" />
        Checking connection…
      </div>
    </FramedShell>
  );
}

/**
 * Holds the app on a loading screen until the API has answered once. Everything below assumes it
 * can reach the API, so this is what turns an unreachable backend into one explanation rather
 * than a failure on every query at once.
 */
export default function ConnectivityGuard() {
  const { initialCheckDone, initialGatePassed, checkInitial } =
    useConnectivityStore();

  useEffect(() => {
    if (!initialCheckDone) void checkInitial();
  }, [initialCheckDone, checkInitial]);

  if (!initialCheckDone) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex items-center gap-3">
          <Spinner />
          <span className="text-xs font-mono text-text-muted">Connecting…</span>
        </div>
      </div>
    );
  }

  if (!initialGatePassed) {
    return <ApiUnavailablePage />;
  }

  return <Outlet />;
}
