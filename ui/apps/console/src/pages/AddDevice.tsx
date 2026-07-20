import { useState, type JSX } from "react";
import { Link } from "react-router-dom";
import Breadcrumb from "@/components/common/Breadcrumb";
import { useAuthStore } from "../stores/authStore";
import {
  ArchiveBoxIcon,
  ArrowTopRightOnSquareIcon,
  BookOpenIcon,
  ChevronDownIcon,
  ChevronRightIcon,
  ChevronUpIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  InformationCircleIcon,
  KeyIcon,
  PlusIcon,
  ServerStackIcon,
  SparklesIcon,
  WrenchIcon,
} from "@heroicons/react/24/outline";
import CopyButton from "../components/common/CopyButton";
import PairingCodeDialog from "@/components/common/PairingCodeDialog";
import CreateInstallKeyDrawer from "@/pages/install-keys/CreateInstallKeyDrawer";
import { modeInfo } from "@/pages/install-keys/constants";
import { useInstallKeys } from "@/hooks/useInstallKeys";
import { useRevealInstallKey } from "@/hooks/useRevealInstallKey";
import InputField from "@/components/common/fields/InputField";
import NumericInput from "@/components/common/fields/NumericInput";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import { LABEL_BASE } from "@/utils/styles";
import {
  Button,
  Card,
  DockerIcon,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

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

/* Methods that run the agent in a container support a tenant-less install: the
 * agent boots without credentials, mints its own pairing code at runtime and
 * prints an accept URL for the browser, so nothing sensitive rides on the
 * command line. Other methods still take a tenant and land the device in the
 * pending list. */
const CODELESS_METHODS: Method[] = ["auto", "docker", "podman"];

/* Two audiences, following the pattern the field has settled on (Tailscale,
 * NetBird): you pick who you're adding, and the mechanism follows. "This
 * machine" installs clean and confirms in the browser; "Fleet" bakes a reusable
 * install key into many machines. */
type Audience = "machine" | "fleet";

const AUDIENCES: {
  id: Audience;
  icon: typeof ComputerDesktopIcon;
  title: string;
  sub: string;
}[] = [
  {
    id: "machine",
    icon: ComputerDesktopIcon,
    title: "One device",
    sub: "Install it, then accept in your browser.",
  },
  {
    id: "fleet",
    icon: ServerStackIcon,
    title: "Fleet",
    sub: "Provision many, unattended, with a reusable install key.",
  },
];

/* Explains where each install-key mode lands a device, shown next to the key
 * picker so the operator knows the outcome before baking the command. */
const MODE_OUTCOME: Record<string, string> = {
  automatic: "accepted the moment it connects",
  manual: "left pending for you to accept",
  webhook: "decided by your integrator",
  allowlist: "accepted if its MAC is allowed",
};

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
  const [aud, setAud] = useState<Audience>("machine");
  const [method, setMethod] = useState<Method>("auto");
  const [pairOpen, setPairOpen] = useState(false);
  const [selectedKeyName, setSelectedKeyName] = useState<string | null>(null);
  const [createKeyOpen, setCreateKeyOpen] = useState(false);
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

  // Install keys that can enroll a fleet: skip the system/legacy key (it's the
  // keyless default, not something you bake in) and anything unusable.
  const { installKeys } = useInstallKeys({ perPage: 50 });
  const usableKeys = installKeys.filter(
    (k) => !k.system && !k.revoked && !k.disabled,
  );
  const selectedKey =
    usableKeys.find((k) => k.name === selectedKeyName) ?? usableKeys[0];
  // Reveal the plaintext only for the picked key, only in fleet mode — it rides
  // in the copyable command, which is exactly what an install key is for.
  const { key: revealedKey } = useRevealInstallKey(
    aud === "fleet" ? (selectedKey?.name ?? null) : null,
    aud === "fleet",
  );

  // "This machine" on a container method installs tenant-less: the agent mints
  // its own pairing code and prints an accept URL, so the command line stays
  // clean. Non-container methods take a tenant and land the device pending.
  const codeless = aud === "machine" && CODELESS_METHODS.includes(method);

  const buildCommand = () => {
    const parts = ["curl -sSf", `${origin}/install.sh`, "|"];
    if (method !== "auto") parts.push(`INSTALL_METHOD=${method}`);
    if (aud === "fleet") {
      parts.push(`INSTALL_KEY=${revealedKey || "…"}`);
    } else if (!codeless) {
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
  // Fleet needs a usable key before it can show a runnable command.
  const fleetBlocked = aud === "fleet" && !selectedKey;

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

      {/* Audience tabs: pick who you're adding; the mechanism follows. */}
      <div
        role="tablist"
        aria-label="What are you adding"
        className="flex gap-1 border-b border-border mb-3"
      >
        {AUDIENCES.map((a) => {
          const active = aud === a.id;
          return (
            <button
              key={a.id}
              type="button"
              role="tab"
              aria-label={a.title}
              aria-selected={active}
              onClick={() => setAud(a.id)}
              className={cn(
                "flex items-start gap-2.5 px-4 pt-3 pb-3 -mb-px border-b-2 text-left transition-colors max-w-[20rem]",
                active
                  ? "border-primary text-primary"
                  : "border-transparent text-text-muted hover:text-text-secondary",
              )}
            >
              <a.icon className="w-4 h-4 shrink-0 mt-0.5" strokeWidth={1.8} />
              <span className="flex flex-col gap-0.5 min-w-0">
                <span className="text-sm font-semibold leading-tight">
                  {a.title}
                </span>
                <span
                  className={cn(
                    "text-2xs leading-snug font-normal hidden sm:block",
                    active ? "text-text-secondary" : "text-text-muted",
                  )}
                >
                  {a.sub}
                </span>
              </span>
            </button>
          );
        })}
      </div>

      <div className="mb-8" />

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
                      className={cn(
                        "px-1.5 py-0.5 text-3xs font-bold uppercase tracking-wider rounded border",
                        m.tag === "Manual"
                          ? "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
                          : "bg-accent-green/15 text-accent-green border-accent-green/20",
                      )}
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

      {/* Fleet — pick (or create) the reusable install key baked into the
          command. Its mode decides where each device lands. */}
      {aud === "fleet" && (
        <div className="mb-6">
          <div className="flex items-center gap-2.5 mb-3">
            <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
              2
            </span>
            <span id="add-device-key-label" className={LABEL_BASE}>
              Install key
            </span>
          </div>

          {usableKeys.length === 0 ? (
            <Card className="p-5">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center shrink-0">
                  <KeyIcon className="w-4.5 h-4.5 text-primary" />
                </div>
                <div>
                  <p className="text-sm text-text-primary font-medium mb-1">
                    No install key yet
                  </p>
                  <p className="text-2xs text-text-muted leading-relaxed mb-3">
                    A fleet enrolls with a reusable install key. Create one,
                    then bake it into your image or provisioning.
                  </p>
                  <Button
                    variant="primary"
                    size="sm"
                    onClick={() => setCreateKeyOpen(true)}
                  >
                    Create an install key
                  </Button>
                </div>
              </div>
            </Card>
          ) : (
            <div className="space-y-2">
              <RadioGroupField
                labelledBy="add-device-key-label"
                value={selectedKey?.name ?? ""}
                onChange={setSelectedKeyName}
              >
                {usableKeys.map((k) => {
                  const info = modeInfo(k.mode);
                  const ModeIcon = info.icon;
                  return (
                    <RadioCard
                      key={k.name}
                      value={k.name}
                      icon={
                        <span className="grid place-items-center w-7 h-7 rounded-lg bg-primary/10 text-primary shrink-0">
                          <ModeIcon className="w-4 h-4" strokeWidth={1.8} />
                        </span>
                      }
                      label={k.name}
                      description={`${info.label} · devices ${MODE_OUTCOME[k.mode]}`}
                    />
                  );
                })}
              </RadioGroupField>

              <div className="flex items-center justify-between pt-1">
                <button
                  type="button"
                  onClick={() => setCreateKeyOpen(true)}
                  className="flex items-center gap-1.5 text-2xs font-medium text-primary hover:text-primary/80 transition-colors"
                >
                  <PlusIcon className="w-3 h-3" strokeWidth={2.5} />
                  New install key
                </button>
                <Link
                  to="/install-keys"
                  className="text-2xs text-text-muted hover:text-text-secondary transition-colors"
                >
                  Manage keys
                </Link>
              </div>
            </div>
          )}
        </div>
      )}

      {/* Command or Manual Instructions */}
      <div className="mb-6">
        <div className="flex items-center gap-2.5 mb-3">
          <span className="w-5 h-5 rounded-full bg-primary/15 border border-primary/25 flex items-center justify-center text-2xs font-bold text-primary">
            {aud === "fleet" ? "3" : "2"}
          </span>
          <span className={LABEL_BASE}>
            {selectedMethod.manual
              ? "Follow the documentation"
              : aud === "fleet"
                ? "Bake into your image or provisioning"
                : "Run on your device"}
          </span>
        </div>

        {fleetBlocked ? (
          <div className="text-xs text-text-muted bg-card border border-border rounded-lg px-4 py-3.5">
            Create an install key above to get your command.
          </div>
        ) : selectedMethod.manual ? (
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
            titleBarSlot={<CopyButton text={command} showLabel />}
          >
            <div className="relative">
              <pre className="text-accent-cyan whitespace-pre overflow-x-auto">
                <span className="text-text-muted select-none">$ </span>
                {command}
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
              className={cn(
                "w-3 h-3 transition-transform",
                showAdvanced && "rotate-90",
              )}
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

      {/* Outcome, per audience: browser accept (container), install-key mode
          (fleet), or the pending list (a machine on a non-container method). */}
      {fleetBlocked ? null : codeless ? (
        <div className="flex items-start gap-3 bg-primary/[0.04] border border-primary/15 rounded-xl px-4 py-3.5 mb-6">
          <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
          <div className="text-xs text-text-secondary leading-relaxed">
            Run the command above. The agent prints a link — open it to accept
            this device in your browser, and it&apos;s ready to use. Away from
            that machine? Use{" "}
            <button
              type="button"
              onClick={() => setPairOpen(true)}
              className="text-primary font-medium hover:text-primary/80 transition-colors"
            >
              the code it shows
            </button>{" "}
            instead.
          </div>
        </div>
      ) : aud === "fleet" ? (
        <div className="flex items-start gap-3 bg-primary/[0.04] border border-primary/15 rounded-xl px-4 py-3.5 mb-6">
          <InformationCircleIcon className="w-4 h-4 text-primary shrink-0 mt-0.5" />
          <div className="text-xs text-text-secondary leading-relaxed">
            Bake this command into your image or provisioning so every machine
            enrolls on first boot with this key. Where each device lands is set
            by the key&apos;s mode above.
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

      {/* Create a key without leaving the page; on success it becomes the
          selected key and the command below fills in. */}
      <CreateInstallKeyDrawer
        open={createKeyOpen}
        onClose={() => setCreateKeyOpen(false)}
        onCreated={(name) => setSelectedKeyName(name)}
      />
    </div>
  );
}
