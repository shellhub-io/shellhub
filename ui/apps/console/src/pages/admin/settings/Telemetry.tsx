import { ChartBarIcon } from "@heroicons/react/24/outline";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Callout, Toggle } from "@shellhub/design-system/primitives";
import { configureTelemetryMutation, getTelemetryOptions } from "@/client";
import { adminQueryDefaults } from "@/hooks/useAdminRequests";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import PageHeader from "@/components/common/PageHeader";
import PageLoader from "@/components/common/PageLoader";
import { adminNavSectionTitle } from "@/components/layout/adminNav";
import { ADMIN_TELEMETRY_PATH } from "@/utils/adminRoute";

const description =
  "Anonymous usage counts this instance sends to telemetry.shellhub.io once a day.";

/**
 * The instance's telemetry switch and the events the next heartbeat sends. The environment
 * overrides the switch when SHELLHUB_TELEMETRY is false or the server has no telemetry client, and
 * the page then shows it locked off.
 */
export default function AdminTelemetry() {
  const { data: status, isLoading, isError } = useQuery({
    ...getTelemetryOptions(),
    ...adminQueryDefaults,
  });
  const invalidate = useInvalidateByIds("getTelemetry");
  const configure = useMutation({
    ...configureTelemetryMutation(),
    onSuccess: invalidate,
  });

  const header = (
    <PageHeader
      icon={<ChartBarIcon className="w-6 h-6" />}
      overline={adminNavSectionTitle(ADMIN_TELEMETRY_PATH)}
      title="Telemetry"
      description={description}
    />
  );

  if (isLoading) {
    return (
      <div>
        {header}
        <div className="mt-8">
          <PageLoader label="Loading settings..." showLabel padding="none" />
        </div>
      </div>
    );
  }

  if (isError || !status) {
    return (
      <div>
        {header}
        <Callout variant="error" className="max-w-2xl">
          Failed to load telemetry settings.
        </Callout>
      </div>
    );
  }

  const locked = status.disabled_by_environment;
  const enabled = status.enabled && !locked;

  return (
    <div>
      {header}

      {configure.isError && (
        <Callout variant="error" className="mb-5 max-w-2xl">
          Failed to update telemetry.
        </Callout>
      )}

      <div className="max-w-2xl space-y-5">
        <SettingsSwitchCard
          icon={<ChartBarIcon />}
          title="Send telemetry"
          description={
            locked
              ? "Disabled by environment: SHELLHUB_TELEMETRY is false or the server could not set up telemetry, so the instance sends nothing. The server log says which."
              : "Counts of namespaces, users, devices and sessions, with the edition and version. No names, addresses or IPs."
          }
          control={
            <Toggle
              enabled={enabled}
              disabled={locked || configure.isPending}
              onChange={() =>
                configure.mutate({ body: { enabled: !status.enabled } })
              }
              aria-label="Toggle telemetry"
            />
          }
        />

        <section
          aria-labelledby="telemetry-preview-title"
          className="px-5 py-4 rounded-xl border border-border bg-card"
        >
          <h2
            id="telemetry-preview-title"
            className="text-sm font-medium text-text-primary"
          >
            Next heartbeat
          </h2>
          <p className="mt-1 text-xs text-text-muted leading-relaxed">
            The events the next heartbeat sends. The client also adds an event
            ID, a timestamp, the operating system and Go versions the server
            runs on, and flags marking the event as a server's with no location
            lookup.
          </p>
          <pre
            aria-label="Heartbeat payload"
            className="mt-3 max-h-96 overflow-auto px-3 py-2 bg-background border border-border rounded-lg text-xs font-mono text-text-secondary"
          >
            {JSON.stringify(status.preview, null, 2)}
          </pre>
        </section>
      </div>
    </div>
  );
}
