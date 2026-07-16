import { type ReactNode, useState } from "react";
import { Link, useLocation, useParams } from "react-router-dom";
import {
  ArrowLeftIcon,
  ClockIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { ExclamationCircleIcon as ExclamationCircleSolidIcon } from "@heroicons/react/24/solid";
import { IconBadge } from "@shellhub/design-system/primitives";
import { type InstallKey } from "@/client";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import InstallKeyEventsTable from "./InstallKeyEventsTable";
import InstallKeyActions from "./InstallKeyActions";
import RevealInstallKeyDialog from "./RevealInstallKeyDialog";
import StatusChip from "./StatusChip";
import KeyValueChip from "./KeyValueChip";
import UsageMeter, { usageLabel } from "./UsageMeter";
import { modeInfo } from "./constants";
import {
  getExpiryInfo,
  getKeyBlockers,
  getUsageInfo,
  isSystemKey,
} from "./helpers";

/** One labelled fact in the key summary strip. */
function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="min-w-0">
      <div className="text-2xs font-medium uppercase tracking-wide text-text-muted">
        {label}
      </div>
      <div className="mt-0.5 text-xs text-text-primary">{children}</div>
    </div>
  );
}

/**
 * A single install key's page: a summary of the key up top, then its full, paginated registration
 * activity below. Keyed by the key's id (digest).
 */
export default function InstallKeyHistoryPage() {
  const { id = "" } = useParams();
  const location = useLocation();
  const state = location.state as { name?: string; key?: InstallKey } | null;

  // Source the key from the list rather than router state alone: the summary must survive a refresh or
  // a deep link (state is only set when arriving from the list) and stay live after a revoke/disable
  // here (the mutation invalidates the list). Fall back to the router-state copy while the list loads.
  // The list caps at 100 keys; past that a deep link degrades to the leaner header.
  const { installKeys } = useInstallKeys({ perPage: 100 });
  const key = installKeys.find((k) => k.id === id) ?? state?.key ?? null;
  const name = key
    ? isSystemKey(key)
      ? "Tenant-only registration"
      : key.name
    : (state?.name ?? "");
  const [revealOpen, setRevealOpen] = useState(false);

  const mode = key ? modeInfo(key.mode) : null;
  const ModeIcon = mode?.icon;

  return (
    <div>
      <Link
        to="/install-keys"
        className="mb-4 inline-flex items-center gap-1.5 text-xs text-text-muted transition-colors hover:text-text-primary"
      >
        <ArrowLeftIcon className="w-3.5 h-3.5" strokeWidth={2} />
        Install Keys
      </Link>

      <div className="mb-7">
        <div className="flex items-center gap-3">
          <IconBadge size="lg" color="primary">
            <TicketIcon className="w-6 h-6" />
          </IconBadge>
          <div className="min-w-0">
            <h1 className="text-xl font-semibold text-text-primary leading-tight">
              {name || "Install Key"}
            </h1>
            <p className="mt-0.5 text-sm text-text-muted">
              Every device that registered with this key.
            </p>
          </div>
        </div>

        {key && (
          <div className="mt-5 flex items-center gap-3 rounded-xl border border-border bg-card/40 px-5 py-4">
            <div className="flex flex-1 flex-wrap items-start gap-x-10 gap-y-4">
              {mode && ModeIcon && (
                <Fact label="Mode">
                  <span className="inline-flex items-center gap-1.5">
                    <ModeIcon
                      className="w-3.5 h-3.5 text-primary"
                      strokeWidth={1.8}
                    />
                    {mode.label}
                  </span>
                </Fact>
              )}
              <Fact label="Usage">
                {(() => {
                  const usage = getUsageInfo(key);
                  const { inert, overused, revoked, disabled } =
                    getKeyBlockers(key);
                  const reached = overused && !revoked && !disabled;
                  return (
                    <div className="space-y-1.5">
                      <span className="font-mono">{usageLabel(usage)}</span>
                      <div className="w-32">
                        <UsageMeter
                          usage={usage}
                          dimmed={inert}
                          reached={reached}
                        />
                      </div>
                    </div>
                  );
                })()}
              </Fact>
              <Fact label="Expires">
                {(() => {
                  const { expired, revoked, disabled } = getKeyBlockers(key);
                  const quiet = revoked || disabled;
                  return (
                    <span
                      className="flex items-center gap-1 font-mono"
                      title={expired ? "Expired" : undefined}
                    >
                      {expired ? (
                        <ExclamationCircleSolidIcon
                          className={`w-3.5 h-3.5 shrink-0 ${quiet ? "" : "text-accent-red"}`}
                        />
                      ) : (
                        <ClockIcon className="w-3.5 h-3.5 shrink-0" />
                      )}
                      {getExpiryInfo(key.expires_at).label}
                    </span>
                  );
                })()}
              </Fact>
              {key.tags && key.tags.length > 0 && (
                <Fact label="Tags">
                  <span className="flex flex-wrap items-center gap-1">
                    {key.tags.map((tag) => (
                      <StatusChip key={tag} label={tag} tone="primary" mono />
                    ))}
                  </span>
                </Fact>
              )}
              <Fact label="Install Key">
                <KeyValueChip
                  label="Key"
                  labelTone="primary"
                  value={
                    key.key_hint ? `${key.key_hint}••••••` : "••••••••••••••"
                  }
                  onClick={() => setRevealOpen(true)}
                  title="Reveal Install Key"
                  ariaLabel="Reveal Install Key"
                />
              </Fact>
            </div>
            <InstallKeyActions installKey={key} />
          </div>
        )}
      </div>

      <RevealInstallKeyDialog
        installKey={revealOpen ? key : null}
        onClose={() => setRevealOpen(false)}
      />

      <InstallKeyEventsTable id={id} perPage={15} />
    </div>
  );
}
