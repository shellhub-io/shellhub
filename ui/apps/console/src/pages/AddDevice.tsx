import { useState, type JSX } from "react";
import { Link } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import BaseDialog from "@/components/common/BaseDialog";
import { useAuthStore } from "../stores/authStore";
import {
  ArchiveBoxIcon,
  ArrowTopRightOnSquareIcon,
  BookOpenIcon,
  CheckCircleIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  ClockIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  InformationCircleIcon,
  PlusIcon,
  ServerStackIcon,
  SparklesIcon,
  WrenchIcon,
} from "@heroicons/react/24/outline";
import CopyButton from "../components/common/CopyButton";
import PairingCodeDialog from "@/components/common/PairingCodeDialog";
import { useDevicePairingEnrollment } from "@/hooks/useDevicePairingEnrollment";
import InputField from "@/components/common/fields/InputField";
import NumericInput from "@/components/common/fields/NumericInput";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import { LABEL_BASE } from "@/utils/styles";
import { formatCountdown } from "@/utils/date";
import {
  Button,
  Card,
  DockerIcon,
  WindowChrome,
} from "@shellhub/design-system/primitives";

/* ─── Types ─── */
type Method =
  | "auto"
  | "docker"
  | "podman"
  | "snap"
  | "standalone"
  | "wsl"
  | "yocto"
  | "buildroot"
  | "freebsd";

interface MethodInfo {
  id: Method;
  label: string;
  tag?: string;
  description: string;
  icon: JSX.Element;
  manual?: boolean;
  docsUrl?: string;
}

const INITIAL_VISIBLE = 3;

/* Methods that run the agent in a container, where a tenant-less install can
 * claim a pre-authorized pairing code and be accepted automatically. Other
 * methods still take a tenant and land the device in the pending list. */
const LIVE_METHODS: Method[] = ["auto", "docker", "podman"];

/* ─── Constants ─── */
const METHODS: MethodInfo[] = [
  {
    id: "auto",
    label: "Auto Detect",
    tag: "Recommended",
    description:
      "Automatically detects Docker, Snap, or Standalone and uses the best available method.",
    icon: <SparklesIcon className="w-5 h-5" />,
  },
  {
    id: "docker",
    label: "Docker",
    description:
      "Run the agent as a Docker container. Requires Docker daemon running on the host.",
    icon: <DockerIcon className="w-5 h-5" />,
  },
  {
    id: "standalone",
    label: "Standalone",
    description:
      "Install directly using runc and systemd. No container runtime required.",
    icon: <ServerStackIcon className="w-5 h-5" />,
  },
  {
    id: "podman",
    label: "Podman",
    description:
      "Alternative to Docker with rootless container capabilities. Requires Podman daemon.",
    icon: <CubeIcon className="w-5 h-5" />,
  },
  {
    id: "snap",
    label: "Snap",
    description:
      "Easy installation via Snap store with automatic updates. Requires snapd service.",
    icon: <ArchiveBoxIcon className="w-5 h-5" />,
  },
  {
    id: "wsl",
    label: "WSL",
    description:
      "Optimized for Windows Subsystem for Linux 2 with systemd and mirrored networking.",
    icon: <ComputerDesktopIcon className="w-5 h-5" />,
  },
  {
    id: "yocto",
    label: "Yocto Project",
    tag: "Manual",
    description:
      "For embedded Linux systems built with the Yocto build system.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/yocto",
    icon: <CpuChipIcon className="w-5 h-5" />,
  },
  {
    id: "buildroot",
    label: "Buildroot",
    tag: "Manual",
    description: "For embedded Linux systems built with Buildroot toolchain.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/buildroot",
    icon: <WrenchIcon className="w-5 h-5" />,
  },
  {
    id: "freebsd",
    label: "FreeBSD",
    tag: "Manual",
    description:
      "For FreeBSD systems. Requires ports tree and manual compilation.",
    manual: true,
    docsUrl: "https://docs.shellhub.io/overview/supported-platforms/freebsd",
    icon: <CommandLineIcon className="w-5 h-5" />,
  },
];

/* ─── Page ─── */
export default function AddDevice() {
  const { tenant } = useAuthStore();
  const [mode, setMode] = useState<"single" | "fleet">("single");
  const [method, setMethod] = useState<Method>("auto");
  const [pairOpen, setPairOpen] = useState(false);
  const [showAllMethods, setShowAllMethods] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [hostname, setHostname] = useState("");
  const [identity, setIdentity] = useState("");
  const [keepaliveInterval, setKeepaliveInterval] = useState("");
  const keepaliveIntervalError =
    keepaliveInterval && parseInt(keepaliveInterval, 10) < 1
      ? "Interval must be a positive number"
      : "";

  const selectedMethod = METHODS.find((m) => m.id === method)!;
  const baseMethods = METHODS.slice(0, INITIAL_VISIBLE);
  const selectedInBase = baseMethods.some((m) => m.id === method);
  const visibleMethods = showAllMethods
    ? METHODS
    : selectedInBase
      ? baseMethods
      : [...baseMethods, selectedMethod];

  const origin = window.location.origin;

  // "This device" on a container method hands the installer a pre-authorized
  // code so the device is accepted automatically and this page confirms it live.
  // The fleet path, non-container methods, and mint failures all fall back to the
  // tenant, which lands devices in the pending tab.
  const liveEnrollment = mode === "single" && LIVE_METHODS.includes(method);
  const enrollment = useDevicePairingEnrollment(liveEnrollment, tenant ?? "");
  const usePairingCode = liveEnrollment && enrollment.phase !== "error";

  const buildCommand = () => {
    const parts = ["curl -sSf", `${origin}/install.sh`, "|"];
    if (method !== "auto") parts.push(`INSTALL_METHOD=${method}`);
    if (usePairingCode) {
      parts.push(`CODE=${enrollment.code || "…"}`);
    } else {
      parts.push(`TENANT_ID=${tenant}`);
    }
    if (hostname.trim()) parts.push(`PREFERRED_HOSTNAME=${hostname.trim()}`);
    if (identity.trim()) parts.push(`PREFERRED_IDENTITY=${identity.trim()}`);
    if (keepaliveInterval.trim() && !keepaliveIntervalError)
      parts.push(`KEEPALIVE_INTERVAL=${parseInt(keepaliveInterval, 10)}`);
    parts.push("sh");
    return parts.join(" ");
  };

  const command = buildCommand();
  // While the pre-authorized code is still being minted there is no runnable
  // command yet, so don't show or let the user copy a placeholder.
  const commandReady = !usePairingCode || Boolean(enrollment.code);

  // Pop a success modal the moment a device connects and is auto-accepted.
  // Tracking the dismissed device by uid re-arms it for the next device without
  // a state-resetting effect.
  const [dismissedUid, setDismissedUid] = useState<string | null>(null);
  const connectedModalOpen =
    enrollment.phase === "connected" &&
    Boolean(enrollment.device) &&
    enrollment.device?.uid !== dismissedUid;

  return (
    <div className="max-w-2xl animate-fade-in">
      <Breadcrumb
        items={[{ label: "Devices", to: "/devices" }, { label: "Add Device" }]}
      />

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
          <PlusIcon className="w-6 h-6 text-primary" />
        </div>
        <div>
          <h1 className="text-xl font-bold text-text-primary">Add Device</h1>
          <p className="text-sm text-text-muted mt-1">
            Install the ShellHub agent to connect a device to your namespace.
          </p>
        </div>
      </div>

      {/* Mode: one device you have shell on (auto-accept) vs many (fleet). */}
      <div
        role="radiogroup"
        aria-label="How many devices"
        className="flex gap-1 p-1 mb-3 bg-card border border-border rounded-xl"
      >
        {(
          [
            {
              id: "single" as const,
              icon: ComputerDesktopIcon,
              title: "One device",
              sub: "Accepted automatically",
            },
            {
              id: "fleet" as const,
              icon: ServerStackIcon,
              title: "A fleet",
              sub: "Accept each device manually",
            },
          ] as const
        ).map((opt) => {
          const active = mode === opt.id;
          return (
            <button
              key={opt.id}
              type="button"
              role="radio"
              aria-label={opt.title}
              aria-checked={active}
              onClick={() => setMode(opt.id)}
              className={`flex-1 flex items-center gap-3 px-3.5 h-14 rounded-lg border text-left transition-colors ${
                active
                  ? "bg-primary/15 border-primary/30 text-primary"
                  : "border-transparent text-text-muted hover:text-text-secondary"
              }`}
            >
              <opt.icon className="w-5 h-5 shrink-0" strokeWidth={1.8} />
              <span className="flex flex-col min-w-0">
                <span className="text-sm font-semibold leading-tight">
                  {opt.title}
                </span>
                <span
                  className={`text-2xs ${active ? "text-primary/75" : "text-text-muted"}`}
                >
                  {opt.sub}
                </span>
              </span>
            </button>
          );
        })}
      </div>

      {/* Secondary, narrow case (headless, no shell, no pre-shared key): claim a
          device that's already running and showing a code. Kept a quiet link
          right under the mode select so all three paths sit together. */}
      <button
        type="button"
        onClick={() => setPairOpen(true)}
        className="mb-8 text-xs text-text-muted hover:text-text-secondary transition-colors"
      >
        <span>
          Already running and showing a code?{" "}
          <span className="text-primary font-medium">Claim it</span>
        </span>
      </button>

      {/* Step 1 — Method */}
      <div className="mb-6">
        <div className="flex items-center gap-2.5 mb-3">
          <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
            1
          </span>
          <span id="add-device-method-label" className={LABEL_BASE}>
            Installation method
          </span>
        </div>
        <div className="space-y-2">
          <RadioGroupField
            labelledBy="add-device-method-label"
            value={method}
            onChange={setMethod}
          >
            {visibleMethods.map((m) => (
              <RadioCard
                key={m.id}
                value={m.id}
                icon={m.icon}
                label={m.label}
                description={m.description}
                adornment={
                  m.tag && (
                    <span
                      className={`px-1.5 py-0.5 text-3xs font-bold uppercase tracking-wider rounded border ${
                        m.tag === "Manual"
                          ? "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
                          : "bg-accent-green/15 text-accent-green border-accent-green/20"
                      }`}
                    >
                      {m.tag}
                    </span>
                  )
                }
              />
            ))}
          </RadioGroupField>

          {!showAllMethods && (
            <button
              type="button"
              onClick={() => setShowAllMethods(true)}
              className="flex items-center justify-center gap-1.5 w-full py-2 text-2xs font-mono text-text-muted hover:text-primary transition-colors"
            >
              <ChevronDownIcon className="w-3 h-3" strokeWidth={2} />
              Show all methods
            </button>
          )}
          {showAllMethods && (
            <button
              type="button"
              onClick={() => setShowAllMethods(false)}
              className="flex items-center justify-center gap-1.5 w-full py-2 text-2xs font-mono text-text-muted hover:text-primary transition-colors"
            >
              <ChevronUpIcon className="w-3 h-3" strokeWidth={2} />
              Show fewer
            </button>
          )}
        </div>
      </div>

      {/* Step 2 — Command or Manual Instructions */}
      <div className="mb-6">
        <div className="flex items-center gap-2.5 mb-3">
          <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
            2
          </span>
          <span className={LABEL_BASE}>
            {selectedMethod.manual
              ? "Follow the documentation"
              : mode === "fleet"
                ? "Bake into your image or provisioning"
                : "Run on your device"}
          </span>
        </div>

        {selectedMethod.manual ? (
          <Card className="p-5">
            <div className="flex items-start gap-3">
              <div className="w-9 h-9 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center shrink-0">
                <BookOpenIcon className="w-4.5 h-4.5 text-accent-yellow" />
              </div>
              <div>
                <p className="text-sm text-text-primary font-medium mb-1">
                  Manual installation required
                </p>
                <p className="text-2xs text-text-muted leading-relaxed mb-3">
                  {selectedMethod.label} requires manual setup. Follow the
                  platform-specific documentation for step-by-step instructions.
                </p>
                <Button
                  variant="warningSoft"
                  as="a"
                  size="sm"
                  href={selectedMethod.docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  iconRight={
                    <ArrowTopRightOnSquareIcon
                      className="w-3 h-3"
                      strokeWidth={2}
                    />
                  }
                >
                  View {selectedMethod.label} guide
                </Button>
              </div>
            </div>
          </Card>
        ) : (
          <WindowChrome
            variant="terminal"
            size="sm"
            titleBarSlot={
              commandReady ? <CopyButton text={command} showLabel /> : null
            }
          >
            <div className="relative">
              <pre className="text-accent-cyan whitespace-pre overflow-x-auto">
                <span className="text-text-muted select-none">$ </span>
                {commandReady ? (
                  command
                ) : (
                  <span className="text-text-muted">
                    Preparing your install command…
                  </span>
                )}
              </pre>
              {/* Chrome renders overlay scrollbars here (no persistent bar), so
                  cue the horizontal scroll with a fade on the right edge. */}
              <div className="pointer-events-none absolute inset-y-0 right-0 w-10 bg-gradient-to-l from-surface to-transparent" />
            </div>
          </WindowChrome>
        )}
      </div>

      {/* Advanced options (only for non-manual methods) */}
      {!selectedMethod.manual && (
        <div className="mb-6">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="flex items-center gap-1.5 text-2xs font-mono text-text-muted hover:text-text-secondary transition-colors"
          >
            <ChevronRightIcon
              className={`w-3 h-3 transition-transform ${showAdvanced ? "rotate-90" : ""}`}
              strokeWidth={2}
            />
            Advanced options
          </button>

          {showAdvanced && (
            <Card className="mt-3 p-4 space-y-4 animate-fade-in">
              <InputField
                id="add-device-hostname"
                label="Preferred Hostname"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={hostname}
                onChange={setHostname}
                placeholder="my-device"
                hint="Override the device hostname reported to ShellHub."
              />
              <InputField
                id="add-device-identity"
                label="Preferred Identity"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={identity}
                onChange={setIdentity}
                placeholder="server-01"
                hint="Set a custom identity string for the device."
              />
              <NumericInput
                id="add-device-keepalive"
                label="Keep Alive Interval"
                labelAdornment={
                  <span className="text-text-muted/50 normal-case tracking-normal text-2xs">
                    (optional)
                  </span>
                }
                value={keepaliveInterval}
                onChange={setKeepaliveInterval}
                placeholder="30"
                hint="Interval in seconds between keep-alive messages sent by the agent. Defaults to 30."
                error={keepaliveIntervalError || undefined}
              />
            </Card>
          )}
        </div>
      )}

      {/* Live status: the code carries the acceptance, so the page confirms the
          device the moment it connects — no trip through the pending list. */}
      {usePairingCode ? (
        <div
          className={`flex items-center gap-3 rounded-xl px-4 py-3.5 mb-6 border ${
            enrollment.phase === "connected"
              ? "border-accent-green/35 bg-accent-green/[0.06]"
              : enrollment.phase === "expired"
                ? "border-border-light bg-card"
                : "border-primary/30 bg-primary/[0.04]"
          }`}
        >
          {enrollment.phase === "connected" ? (
            <CheckCircleIcon className="w-5 h-5 text-accent-green shrink-0" />
          ) : enrollment.phase === "expired" ? (
            <ClockIcon className="w-5 h-5 text-text-muted shrink-0" />
          ) : (
            <span className="relative flex h-3 w-3 shrink-0">
              <span className="absolute inline-flex h-full w-full rounded-full bg-primary opacity-60 animate-ping" />
              <span className="relative inline-flex h-3 w-3 rounded-full bg-primary" />
            </span>
          )}

          <div className="flex-1 min-w-0">
            {enrollment.phase === "connected" ? (
              <>
                <p className="text-sm font-medium text-accent-green">
                  {enrollment.device?.name || "Device"} is connected
                </p>
                <p className="text-2xs text-text-muted mt-0.5">
                  Accepted automatically. It&apos;s ready to use.
                </p>
              </>
            ) : enrollment.phase === "expired" ? (
              <>
                <p className="text-sm font-medium text-text-primary">
                  Code expired
                </p>
                <p className="text-2xs text-text-muted mt-0.5">
                  The pairing code timed out.
                </p>
              </>
            ) : (
              <>
                <p className="text-sm font-medium text-text-primary">
                  Waiting for this device to connect…
                </p>
                <p className="text-2xs text-text-muted mt-0.5">
                  Run the command above. The device shows up here automatically.
                  {enrollment.secondsLeft > 0
                    ? ` Code expires in ${formatCountdown(enrollment.secondsLeft)}.`
                    : ""}
                </p>
              </>
            )}
          </div>

          {enrollment.phase === "connected" ? (
            <div className="flex items-center gap-2 shrink-0">
              <Button variant="ghost" size="sm" onClick={enrollment.regenerate}>
                New code
              </Button>
              {enrollment.device?.uid ? (
                <Button
                  as={Link}
                  to={`/devices/${enrollment.device.uid}`}
                  variant="successSoft"
                  size="sm"
                >
                  View device
                </Button>
              ) : null}
            </div>
          ) : enrollment.phase === "expired" ? (
            <Button
              variant="secondary"
              size="sm"
              onClick={enrollment.regenerate}
            >
              New code
            </Button>
          ) : (
            <Button
              variant="ghost"
              size="sm"
              onClick={enrollment.regenerate}
              className="shrink-0"
            >
              New code
            </Button>
          )}
        </div>
      ) : mode === "fleet" ? (
        <div className="flex items-start gap-3 bg-primary/[0.04] border border-primary/15 rounded-xl px-4 py-3.5 mb-6">
          <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
          <div className="text-xs text-text-secondary leading-relaxed">
            Bake this command into your image or provisioning so every device
            installs the agent on first boot. Each one lands in the{" "}
            <Link
              to="/devices?status=pending"
              className="text-primary font-medium hover:text-primary/80 transition-colors"
            >
              Pending tab
            </Link>{" "}
            for you to accept one by one, or automate it with the API.
          </div>
        </div>
      ) : (
        <div className="flex items-start gap-3 bg-primary/[0.04] border border-primary/15 rounded-xl px-4 py-3.5 mb-6">
          <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
          <div className="text-xs text-text-secondary leading-relaxed">
            After installing, your device will appear in the{" "}
            <Link
              to="/devices?status=pending"
              className="text-primary font-medium hover:text-primary/80 transition-colors"
            >
              Pending tab
            </Link>{" "}
            and must be accepted before you can connect to it.
          </div>
        </div>
      )}

      {/* Docs link */}
      <div className="flex items-center gap-2 text-2xs text-text-muted">
        <BookOpenIcon className="w-3.5 h-3.5" />
        <a
          href="https://docs.shellhub.io/user-guides/devices/adding"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:text-primary transition-colors"
        >
          Check the documentation for more installation methods
        </a>
        <ArrowTopRightOnSquareIcon className="w-3 h-3" strokeWidth={2} />
      </div>

      <PairingCodeDialog open={pairOpen} onClose={() => setPairOpen(false)} />

      <BaseDialog
        open={connectedModalOpen}
        onClose={() => setDismissedUid(enrollment.device?.uid ?? null)}
        size="sm"
        aria-label="Device connected"
      >
        <div className="p-6 text-center">
          <div className="w-14 h-14 mx-auto mb-4 rounded-2xl bg-accent-green/15 border border-accent-green/25 flex items-center justify-center">
            <CheckCircleIcon className="w-8 h-8 text-accent-green" />
          </div>
          <h2 className="text-lg font-semibold text-text-primary">
            <span className="font-mono">{enrollment.device?.name}</span> is
            connected
          </h2>
          <p className="text-sm text-text-muted mt-1">
            Accepted automatically. It&apos;s ready to use.
          </p>
          <div className="flex gap-2 mt-6">
            <Button
              variant="secondary"
              fullWidth
              onClick={() => {
                setDismissedUid(enrollment.device?.uid ?? null);
                enrollment.regenerate();
              }}
            >
              Add another
            </Button>
            {enrollment.device?.uid ? (
              <Button
                as={Link}
                to={`/devices/${enrollment.device.uid}`}
                variant="primary"
                fullWidth
              >
                View device
              </Button>
            ) : null}
          </div>
        </div>
      </BaseDialog>
    </div>
  );
}
