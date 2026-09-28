import { useEffect, useState, type ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import {
  CpuChipIcon,
  CheckCircleIcon,
  XCircleIcon,
  ArrowRightIcon,
  CommandLineIcon,
  CheckIcon,
  ChevronUpDownIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { Button, Dropdown, Spinner } from "@shellhub/design-system/primitives";
import type { ResolveDeviceLoginCodeResponse } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import {
  clearPendingDeviceCode,
  setPendingDeviceCode,
} from "@/utils/navigation";
import { isEnterpriseOrCloud } from "@/env";
import {
  NamespaceCreateForm,
  CommunityInstructions,
} from "@/components/common/CreateNamespace";
import { useAcceptDevice } from "@/hooks/useDeviceMutations";
import {
  useResolveDeviceCode,
  useAcceptDevicePairing,
} from "@/hooks/useDeviceCode";
import { useSwitchNamespace } from "@/hooks/useNamespaceMutations";
import {
  useNamespace,
  useNamespaces,
  type Namespace,
} from "@/hooks/useNamespaces";
import { getInitials } from "@/utils/string";
import PairingCodeForm from "@/components/common/PairingCodeForm";
import { LABEL } from "@/utils/styles";
import { useHasPermission } from "@/hooks/useHasPermission";
import { isSubscriptionBlocked } from "@/utils/billing";
import { getAcceptErrorMessage } from "@/utils/acceptErrors";

type DevicePreview = ResolveDeviceLoginCodeResponse;

type Branch =
  | { kind: "loading" }
  | { kind: "missing-code" }
  | { kind: "error" }
  | { kind: "switching" }
  | { kind: "ready"; device: DevicePreview }
  | { kind: "pick-namespace"; device: DevicePreview }
  | { kind: "already-accepted"; device: DevicePreview }
  | { kind: "success"; device: DevicePreview }
  | {
      kind: "pairing-success";
      device: DevicePreview;
      uid: string;
      tenantId: string;
      namespace: string;
      teamDevice: boolean;
    };

const LINK_CLASS =
  "text-xs text-text-muted hover:text-text-secondary transition-colors";

/**
 * Enrols a device from a code, either as a page or inside a dialog. initialCode is what an
 * accept-device link arrives with, so the code does not have to be typed twice. frame wraps the
 * flow in the page's card, and is told when a step needs a wider one.
 */
export default function AcceptDeviceFlow({
  initialCode = "",
  inDialog = false,
  frame,
}: {
  initialCode?: string;
  inDialog?: boolean;
  frame?: (wide: boolean, content: ReactNode) => ReactNode;
}) {
  const authTenant = useAuthStore((s) => s.tenant);

  const [code, setCode] = useState(initialCode);
  const {
    device,
    isLoading: isResolving,
    isError,
  } = useResolveDeviceCode(code);

  const acceptDevice = useAcceptDevice();
  const acceptPairing = useAcceptDevicePairing();
  const switchNamespace = useSwitchNamespace();

  const [actionBranch, setActionBranch] = useState<Branch | null>(null);
  const [actionError, setActionError] = useState("");
  const [selectedTenant, setSelectedTenant] = useState("");
  const { namespace: targetNamespace } = useNamespace(
    selectedTenant || authTenant || "",
  );
  const hasSubscription = isSubscriptionBlocked(targetNamespace?.billing);
  const canSubscribeInAuth = useHasPermission("billing:subscribe");
  const canSubscribe =
    canSubscribeInAuth && (!selectedTenant || selectedTenant === authTenant);

  const finish = (b: Branch) => {
    clearPendingDeviceCode();
    setActionBranch(b);
  };

  const branch: Branch = (() => {
    if (actionBranch) return actionBranch;
    if (!code) return { kind: "missing-code" };
    if (isResolving) return { kind: "loading" };
    if (isError) return { kind: "error" };
    if (!device) return { kind: "loading" };
    if (device.kind === "pairing") return { kind: "pick-namespace", device };
    if (device.status === "accepted")
      return { kind: "already-accepted", device };
    if (device.tenant_id && device.tenant_id !== authTenant)
      return { kind: "switching" };
    return { kind: "ready", device };
  })();

  useEffect(() => {
    if (branch.kind === "error" || branch.kind === "already-accepted") {
      clearPendingDeviceCode();
    }
  }, [branch.kind]);

  useEffect(() => {
    if (
      device?.tenant_id &&
      device.tenant_id !== authTenant &&
      !actionBranch &&
      switchNamespace.isIdle
    ) {
      switchNamespace
        .mutateAsync({
          tenantId: device.tenant_id,
          redirectTo: `/accept-device?code=${code}`,
        })
        .catch(() => finish({ kind: "error" }));
    }
  }, [device?.tenant_id, authTenant, code, actionBranch, switchNamespace]);

  const handleAccept = async (device: DevicePreview) => {
    setActionError("");
    try {
      if (!device.uid) return;
      await acceptDevice.mutateAsync({ path: { uid: device.uid } });
      finish({ kind: "success", device });
    } catch (err) {
      setActionError(getAcceptErrorMessage(err, hasSubscription, canSubscribe));
    }
  };

  const handleAcceptPairing = async (device: DevicePreview) => {
    if (!selectedTenant) return;
    setActionError("");
    try {
      const data = await acceptPairing.mutateAsync({
        path: { code },
        body: { tenant_id: selectedTenant },
      });
      finish({
        kind: "pairing-success",
        device,
        uid: data.uid ?? "",
        tenantId: data.tenant_id ?? "",
        namespace: data.namespace ?? "",
        teamDevice: !data.owner_id,
      });
    } catch (err) {
      setActionError(getAcceptErrorMessage(err, hasSubscription, canSubscribe));
    }
  };

  const content = (
    <>
      {branch.kind === "loading" && <StatusMessage label="Checking code..." />}

      {branch.kind === "switching" && (
        <StatusMessage label="Switching namespace..." />
      )}

      {branch.kind === "missing-code" && (
        <div className="motion-safe:animate-fade-in">
          <div className="text-center mb-6">
            <div className="w-14 h-14 mx-auto mb-4 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center">
              <CommandLineIcon
                className="w-7 h-7 text-primary"
                strokeWidth={1.5}
              />
            </div>
            <h2 className="text-lg font-semibold text-text-primary">
              Claim a device
            </h2>
            <p className="text-sm text-text-muted mt-1">
              Enter the code your device is showing. Not showing one? Run{" "}
              <CommandChip /> to generate it.
            </p>
          </div>

          <PairingCodeForm
            onSubmit={(c) => {
              if (!inDialog) setPendingDeviceCode(c);
              setCode(c);
            }}
            submitLabel="Claim device"
          />
          {!inDialog && <DashboardLink />}
        </div>
      )}

      {branch.kind === "error" && (
        <ResultMessage
          tone="error"
          icon={XCircleIcon}
          title="Invalid or Expired Code"
          description={
            <>
              This code is invalid or has expired. Run <CommandChip /> on the
              device to get a new one.
            </>
          }
          action={
            <div className="flex flex-col items-center gap-3">
              <Button
                variant="secondary"
                size="md"
                onClick={() => {
                  setActionBranch(null);
                  setCode("");
                }}
              >
                Enter another code
              </Button>
              {!inDialog && <DashboardLink />}
            </div>
          }
        />
      )}

      {branch.kind === "already-accepted" && (
        <ResultMessage
          tone="success"
          icon={CheckCircleIcon}
          title="Device Already Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              is already accepted into this namespace.
            </>
          }
          action={<ViewDeviceLink uid={branch.device.uid} />}
        />
      )}

      {branch.kind === "success" && (
        <ResultMessage
          tone="success"
          icon={CheckCircleIcon}
          title="Device Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              joined{" "}
              <span className="text-text-primary font-medium">
                {branch.device.namespace}
              </span>
              . You can return to your terminal.
            </>
          }
          action={<ViewDeviceLink uid={branch.device.uid} />}
        />
      )}

      {branch.kind === "ready" && (
        <div className="text-center">
          <div className="relative inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 mb-5">
            <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl motion-safe:animate-pulse-subtle" />
            <CpuChipIcon
              className="relative w-8 h-8 text-primary"
              strokeWidth={1.25}
            />
          </div>

          <h2 className="text-lg font-semibold text-text-primary mb-2">
            Accept this device?
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed mb-6">
            A device is asking to join{" "}
            <span className="text-text-primary font-medium">
              {branch.device.namespace}
            </span>
            . Review its identity before accepting.
          </p>

          <dl className="text-left text-sm bg-surface/60 border border-border rounded-xl divide-y divide-border/70 overflow-hidden mb-6">
            <SpecRow label="hostname" value={branch.device.name} />
            <SpecRow label="os" value={branch.device.info?.pretty_name} />
            <SpecRow label="mac" value={branch.device.identity?.mac} />
            <SpecRow label="namespace" value={branch.device.namespace} />
            <div className="flex items-center justify-between gap-4 px-4 py-2.5">
              <dt className="font-mono text-2xs uppercase tracking-wider text-text-muted">
                status
              </dt>
              <dd>
                <span className="inline-flex items-center gap-1.5 rounded-full bg-accent-yellow/10 border border-accent-yellow/20 px-2.5 py-0.5 font-mono text-2xs text-accent-yellow">
                  <span className="w-1.5 h-1.5 rounded-full bg-accent-yellow motion-safe:animate-pulse-subtle" />
                  pending
                </span>
              </dd>
            </div>
          </dl>

          {actionError && (
            <p
              className="text-sm text-accent-red mb-4 motion-safe:animate-shake"
              role="alert"
            >
              {actionError}
            </p>
          )}

          <Button
            variant="primary"
            size="md"
            fullWidth
            loading={acceptDevice.isPending}
            icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
            onClick={() => void handleAccept(branch.device)}
          >
            Accept device
          </Button>

          <CancelRow inDialog={inDialog} onReset={() => setCode("")} />
        </div>
      )}

      {branch.kind === "pick-namespace" && (
        <div className="text-center">
          <div className="relative inline-flex items-center justify-center w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 mb-5">
            <div className="absolute inset-0 rounded-2xl bg-primary/10 blur-xl motion-safe:animate-pulse-subtle" />
            <CpuChipIcon
              className="relative w-8 h-8 text-primary"
              strokeWidth={1.25}
            />
          </div>

          <h2 className="text-lg font-semibold text-text-primary mb-2">
            Accept this device?
          </h2>
          <p className="text-sm text-text-secondary leading-relaxed mb-6">
            A device is asking to join one of your namespaces. Review its
            identity and choose where it belongs.
          </p>

          <div className="grid gap-6 text-left">
            <div>
              <dl className="text-sm bg-surface/60 border border-border rounded-xl divide-y divide-border/70 overflow-hidden mb-2">
                <SpecRow label="code" value={formatPairingCode(code)} />
                <SpecRow label="hostname" value={branch.device.name} />
                <SpecRow label="os" value={branch.device.info?.pretty_name} />
                <SpecRow label="mac" value={branch.device.identity?.mac} />
              </dl>
              <p className="text-2xs text-text-muted mb-6">
                Check that the code matches the one your terminal shows.
              </p>

              <AcceptingAs code={code} canSwitch={!inDialog} />
            </div>

            <div>
              <div className="mb-6">
                <NamespacePicker
                  value={selectedTenant}
                  onChange={setSelectedTenant}
                  preferredTenant={authTenant ?? ""}
                />
              </div>

              {actionError && (
                <p
                  className="text-sm text-accent-red mb-4 motion-safe:animate-shake"
                  role="alert"
                >
                  {actionError}
                </p>
              )}

              <Button
                variant="primary"
                size="md"
                fullWidth
                loading={acceptPairing.isPending}
                disabled={!selectedTenant}
                icon={<CheckCircleIcon className="w-4 h-4" strokeWidth={2} />}
                onClick={() => void handleAcceptPairing(branch.device)}
              >
                Accept device
              </Button>
            </div>
          </div>

          <CancelRow inDialog={inDialog} onReset={() => setCode("")} />
        </div>
      )}

      {branch.kind === "pairing-success" && (
        <ResultMessage
          tone="success"
          icon={CheckCircleIcon}
          title="Device Accepted"
          description={
            <>
              <span className="font-mono text-text-primary">
                {branch.device.name}
              </span>{" "}
              joined{" "}
              <span className="text-text-primary font-medium">
                {branch.namespace}
              </span>
              . The agent will connect automatically. You can return to your
              terminal.
              {branch.teamDevice && (
                <>
                  {" "}
                  It took the place of a team device with the same MAC address,
                  so it stays the team&apos;s rather than being tied to you.
                </>
              )}
            </>
          }
          action={
            branch.uid ? (
              <Button
                variant="successSoft"
                size="md"
                iconRight={
                  <ArrowRightIcon className="w-4 h-4" strokeWidth={2} />
                }
                onClick={() =>
                  void switchNamespace.mutateAsync({
                    tenantId: branch.tenantId,
                    redirectTo: `/devices/${branch.uid}`,
                  })
                }
              >
                View device
              </Button>
            ) : undefined
          }
        />
      )}
    </>
  );

  return frame ? frame(branch.kind === "pick-namespace", content) : content;
}

function DashboardLink() {
  return (
    <Link to="/dashboard" className={LINK_CLASS}>
      Go to dashboard
    </Link>
  );
}

function CancelRow({
  inDialog,
  onReset,
}: {
  inDialog: boolean;
  onReset: () => void;
}) {
  return (
    <div className="text-center mt-4">
      {inDialog ? (
        <button type="button" onClick={onReset} className={LINK_CLASS}>
          Use a different code
        </button>
      ) : (
        <DashboardLink />
      )}
    </div>
  );
}

function NamespacePicker({
  value,
  onChange,
  preferredTenant,
}: {
  value: string;
  onChange: (tenant: string) => void;
  preferredTenant: string;
}) {
  const { namespaces, isLoading } = useNamespaces();
  const [open, setOpen] = useState(false);

  useEffect(() => {
    if (value || namespaces.length === 0) return;
    const preferred = namespaces.find((n) => n.tenant_id === preferredTenant);
    onChange(preferred?.tenant_id ?? namespaces[0].tenant_id ?? "");
  }, [value, namespaces, preferredTenant, onChange]);

  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-sm text-text-muted">
        <Spinner size="sm" />
        Loading namespaces...
      </div>
    );
  }

  if (namespaces.length === 0) {
    return (
      <div className="space-y-3">
        <p className="text-sm text-text-secondary leading-relaxed text-justify">
          You don't have any namespaces yet. Create one to get started.
        </p>
        {isEnterpriseOrCloud() ? (
          <NamespaceCreateForm />
        ) : (
          <CommunityInstructions />
        )}
      </div>
    );
  }

  const selected = namespaces.find((n) => n.tenant_id === value);

  return (
    <div>
      <p className={LABEL}>Namespace</p>
      <Dropdown open={open} onOpenChange={setOpen}>
        <Dropdown.Trigger>
          <button
            type="button"
            aria-label={`Namespace: ${selected?.name ?? "none"}`}
            className="w-full flex items-center gap-3 rounded-xl border border-border bg-card px-3.5 py-3 text-left hover:border-border-light transition-colors focus:outline-none focus-visible:border-primary/50 focus-visible:ring-1 focus-visible:ring-primary/20"
          >
            {selected ? (
              <NamespaceCard namespace={selected} />
            ) : (
              <span className="flex-1 text-sm text-text-muted">
                Choose a namespace
              </span>
            )}
            <ChevronUpDownIcon className="w-4 h-4 shrink-0 text-text-muted" />
          </button>
        </Dropdown.Trigger>

        <Dropdown.Panel aria-label="Namespaces">
          {namespaces.map((namespace) => {
            const checked = namespace.tenant_id === value;

            return (
              <Dropdown.Item
                key={namespace.tenant_id}
                label={namespace.name}
                role="menuitemradio"
                aria-checked={checked}
                onSelect={() => onChange(namespace.tenant_id)}
                className="gap-3 py-2.5"
              >
                <NamespaceCard namespace={namespace} />
                {checked && (
                  <CheckIcon className="w-4 h-4 shrink-0 text-primary" />
                )}
              </Dropdown.Item>
            );
          })}
        </Dropdown.Panel>
      </Dropdown>
    </div>
  );
}

function InitialsAvatar({
  label,
  shape,
}: {
  label: string;
  shape: "circle" | "square";
}) {
  return (
    <span
      className={cn(
        "w-8 h-8 shrink-0 bg-primary/15 border border-primary/20 flex items-center justify-center text-primary text-2xs font-bold font-mono",
        shape === "circle" ? "rounded-full" : "rounded-lg",
      )}
    >
      {getInitials(label) || "?"}
    </span>
  );
}

function NamespaceCard({
  namespace,
}: {
  namespace: Pick<Namespace, "name" | "tenant_id">;
}) {
  return (
    <span className="flex flex-1 min-w-0 items-center gap-3">
      <InitialsAvatar label={namespace.name} shape="square" />
      <span className="min-w-0">
        <span className="block text-sm font-medium text-text-primary truncate">
          {namespace.name}
        </span>
        <span className="block text-2xs font-mono text-text-muted truncate">
          {namespace.tenant_id}
        </span>
      </span>
    </span>
  );
}

function CommandChip() {
  return (
    <code className="inline-flex items-center gap-1 rounded bg-surface border border-border px-1.5 py-0.5 font-mono text-xs text-text-primary whitespace-nowrap">
      <CommandLineIcon
        className="w-3.5 h-3.5 text-text-muted"
        strokeWidth={1.5}
      />
      shellhub-agent login
    </code>
  );
}

function StatusMessage({ label }: { label: string }) {
  return (
    <div
      className="flex flex-col items-center gap-3 py-6"
      role="status"
      aria-live="polite"
    >
      <Spinner size="2xl" />
      <p className="text-sm text-text-muted">{label}</p>
    </div>
  );
}

function formatPairingCode(code: string) {
  const normalized = code.replace(/[^0-9A-Za-z]/g, "").toUpperCase();
  return normalized.length === 8
    ? `${normalized.slice(0, 4)}-${normalized.slice(4)}`
    : normalized;
}

function AcceptingAs({
  code,
  canSwitch,
}: {
  code: string;
  canSwitch: boolean;
}) {
  const name = useAuthStore((s) => s.name);
  const email = useAuthStore((s) => s.email);
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const switchAccount = () => {
    logout();
    void navigate(
      `/login?redirect=${encodeURIComponent(`/accept-device?code=${code}`)}`,
    );
  };

  return (
    <section
      aria-label="Accepting as"
      className="text-left rounded-xl border border-border bg-surface/60 p-4"
    >
      <div className="flex items-end justify-between gap-3">
        <div className="min-w-0">
          <p className={LABEL}>Accepting as</p>
          <span className="flex items-center gap-2.5 min-w-0">
            <InitialsAvatar label={name || email || ""} shape="circle" />
            <span className="min-w-0">
              <span className="block text-sm font-medium text-text-primary truncate">
                {name || email}
              </span>
              {name && email && (
                <span className="block text-xs text-text-muted truncate">
                  {email}
                </span>
              )}
            </span>
          </span>
        </div>
        {canSwitch && (
          <button
            type="button"
            onClick={switchAccount}
            className="shrink-0 text-xs text-primary hover:underline"
          >
            Not you? Switch account
          </button>
        )}
      </div>
      <p className="mt-3 pt-3 border-t border-border/70 text-xs text-text-muted leading-relaxed">
        You&apos;ll own this device. If you leave this namespace or lose
        permission to accept devices in it, the device is removed with you.
      </p>
    </section>
  );
}

function SpecRow({ label, value }: { label: string; value?: string }) {
  return (
    <div className="flex items-center justify-between gap-4 px-4 py-2.5">
      <dt className="font-mono text-2xs uppercase tracking-wider text-text-muted">
        {label}
      </dt>
      <dd className="font-mono text-text-primary truncate">{value || "—"}</dd>
    </div>
  );
}

function ViewDeviceLink({ uid }: { uid?: string }) {
  if (!uid) return null;
  return (
    <Button
      as={Link}
      to={`/devices/${uid}`}
      variant="successSoft"
      size="md"
      iconRight={<ArrowRightIcon className="w-4 h-4" strokeWidth={2} />}
    >
      View device
    </Button>
  );
}

const TONES = {
  error: {
    ring: "bg-accent-red/10 border-accent-red/20",
    icon: "text-accent-red",
  },
  success: {
    ring: "bg-accent-green/10 border-accent-green/20",
    icon: "text-accent-green",
  },
} as const;

function ResultMessage({
  tone,
  icon: Icon,
  title,
  description,
  action,
}: {
  tone: keyof typeof TONES;
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  title: string;
  description: React.ReactNode;
  action?: React.ReactNode;
}) {
  return (
    <div className="text-center motion-safe:animate-slide-up">
      <div
        className={cn(
          "inline-flex items-center justify-center w-14 h-14 rounded-2xl border mb-5",
          TONES[tone].ring,
        )}
      >
        <Icon className={cn("w-7 h-7", TONES[tone].icon)} strokeWidth={1.5} />
      </div>
      <h2 className="text-lg font-semibold text-text-primary mb-3">{title}</h2>
      <p className="text-sm text-text-secondary leading-relaxed mb-6">
        {description}
      </p>
      {action}
    </div>
  );
}
