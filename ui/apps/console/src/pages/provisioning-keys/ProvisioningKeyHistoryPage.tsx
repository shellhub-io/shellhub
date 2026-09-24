import { type ReactNode, useState } from "react";
import { useLocation, useParams } from "react-router-dom";
import {
  NoSymbolIcon,
  PauseCircleIcon,
  TicketIcon,
} from "@heroicons/react/24/outline";
import { IconBadge } from "@shellhub/design-system/primitives";
import { type ProvisioningKey } from "@/client";
import { useProvisioningKeys } from "@/hooks/useProvisioningKeys";
import PageLoader from "@/components/common/PageLoader";
import Breadcrumb from "@/components/common/Breadcrumb";
import ResourceNotFound from "@/components/common/ResourceNotFound";
import ProvisioningKeyEventsTable from "./ProvisioningKeyEventsTable";
import ProvisioningKeyActions from "./ProvisioningKeyActions";
import RevealProvisioningKeyDialog from "./RevealProvisioningKeyDialog";
import StatusChip from "./StatusChip";
import KeyValueChip from "./KeyValueChip";
import UsageMeter from "./UsageMeter";
import { modeInfo } from "./constants";
import { provisioningKeyDisplayName, isPairingKey } from "./helpers";
import ExpiryLabel from "./ExpiryLabel";

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
 * A single provisioning key's page: a summary of the key up top, then its full, paginated registration
 * activity below. Keyed by the key's id (digest).
 */
export default function ProvisioningKeyHistoryPage() {
  const { id = "" } = useParams();
  const location = useLocation();
  const state = location.state as {
    name?: string;
    key?: ProvisioningKey;
  } | null;

  const { provisioningKeys, isLoading } = useProvisioningKeys({ perPage: 100 });
  const key = provisioningKeys.find((k) => k.id === id) ?? state?.key ?? null;
  const name = key ? provisioningKeyDisplayName(key) : (state?.name ?? "");
  const [revealOpen, setRevealOpen] = useState(false);

  const mode = key ? modeInfo(key.mode) : null;
  const ModeIcon = mode?.icon;

  if (isLoading && !key) {
    return <PageLoader label="Loading provisioning key" />;
  }

  if (!key) {
    return (
      <ResourceNotFound
        icon={TicketIcon}
        resource="Provisioning key"
        backTo="/settings/provisioning-keys"
      />
    );
  }

  return (
    <div>
      <Breadcrumb
        items={[
          { label: "Settings", to: "/settings" },
          { label: "Provisioning Keys", to: "/settings/provisioning-keys" },
          { label: name || "Provisioning Key" },
        ]}
        className="mb-4"
      />

      <div className="mb-7">
        <div className="flex items-center gap-3">
          <IconBadge size="lg" color="primary">
            <TicketIcon className="w-6 h-6" />
          </IconBadge>
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <h1 className="text-xl font-semibold text-text-primary leading-tight">
                {name || "Provisioning Key"}
              </h1>
              {key?.revoked && (
                <StatusChip icon={NoSymbolIcon} label="Revoked" tone="red" />
              )}
              {key?.disabled && !key.revoked && (
                <StatusChip
                  icon={PauseCircleIcon}
                  label="Disabled"
                  tone="muted"
                />
              )}
            </div>
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
                <div className="w-32">
                  <UsageMeter provisioningKey={key} />
                </div>
              </Fact>
              <Fact label="Expires">
                <ExpiryLabel provisioningKey={key} />
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
              {/* The provisioning key chip: user and legacy keys show it (click reveals the key — or, for the
                  secretless legacy key, its fingerprint). Pairing has nothing to reveal (it accepts on the
                  printed code), so it's the only one without the chip. */}
              {!isPairingKey(key) && (
                <Fact label="Provisioning Key">
                  <KeyValueChip
                    label="Key"
                    labelTone="primary"
                    value={
                      key.key_hint ? `${key.key_hint}••••••` : "••••••••••••••"
                    }
                    onClick={() => setRevealOpen(true)}
                    title="Reveal Provisioning Key"
                    ariaLabel="Reveal Provisioning Key"
                  />
                </Fact>
              )}
            </div>
            <ProvisioningKeyActions provisioningKey={key} />
          </div>
        )}
      </div>

      <RevealProvisioningKeyDialog
        provisioningKey={revealOpen ? key : null}
        onClose={() => setRevealOpen(false)}
      />

      <ProvisioningKeyEventsTable id={id} />
    </div>
  );
}
