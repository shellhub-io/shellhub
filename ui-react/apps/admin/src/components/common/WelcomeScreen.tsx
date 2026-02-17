import { useState } from "react";
import { Link } from "react-router-dom";
import {
  ShieldCheckIcon,
  CpuChipIcon,
  UserGroupIcon,
  CommandLineIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  ArrowRightIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";

interface WelcomeScreenProps {
  namespaceName: string;
  tenantId: string;
}

function ConnectionGrid() {
  return (
    <div className="connection-grid">
      {/* Horizontal lines */}
      <div
        className="connection-line"
        style={{ top: "20%", left: 0, width: "60%", animationDelay: "0s" }}
      />
      <div
        className="connection-line"
        style={{
          top: "45%",
          left: "30%",
          width: "70%",
          animationDelay: "1.5s",
        }}
      />
      <div
        className="connection-line"
        style={{
          top: "70%",
          left: "10%",
          width: "50%",
          animationDelay: "0.8s",
        }}
      />
      <div
        className="connection-line"
        style={{
          top: "85%",
          left: "40%",
          width: "60%",
          animationDelay: "2.2s",
        }}
      />

      {/* Vertical lines */}
      <div
        className="connection-line-v"
        style={{ left: "25%", top: 0, height: "60%", animationDelay: "0.5s" }}
      />
      <div
        className="connection-line-v"
        style={{
          left: "55%",
          top: "20%",
          height: "80%",
          animationDelay: "1.8s",
        }}
      />
      <div
        className="connection-line-v"
        style={{ left: "80%", top: "10%", height: "50%", animationDelay: "3s" }}
      />

      {/* Intersection dots */}
      <div
        className="connection-dot"
        style={{ top: "20%", left: "25%", animationDelay: "0.3s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "45%", left: "55%", animationDelay: "1.2s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "70%", left: "25%", animationDelay: "2s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "45%", left: "80%", animationDelay: "0.8s" }}
      />
      <div
        className="connection-dot"
        style={{ top: "85%", left: "55%", animationDelay: "2.5s" }}
      />
    </div>
  );
}

const steps = [
  {
    num: "01",
    title: "Connect your first device",
    description:
      "Install the ShellHub agent on any Linux device to start managing it remotely.",
    icon: <CpuChipIcon className="w-6 h-6" />,
    hasCommand: true,
  },
  {
    num: "02",
    title: "Invite your team",
    description: "Add members to collaborate and manage devices together.",
    icon: <UserGroupIcon className="w-6 h-6" />,
    linkTo: "/team",
    linkLabel: "Manage team",
  },
  {
    num: "03",
    title: "Set up firewall rules",
    description: "Control who can access your devices and from where.",
    icon: <ShieldCheckIcon className="w-6 h-6" />,
    linkTo: "/firewall/rules",
    linkLabel: "Configure rules",
  },
];

export default function WelcomeScreen({
  namespaceName,
  tenantId,
}: WelcomeScreenProps) {
  const [copied, setCopied] = useState(false);

  const installCmd = `curl -sSf "${window.location.origin}/install.sh?tenant_id=${tenantId}" | sh`;

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="-mx-8 -mt-8 min-h-[calc(100vh-3.5rem)]">
      {/* Hero */}
      <div className="relative px-8 pt-16 pb-12 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/10 via-transparent to-transparent" />
        <div className="absolute top-10 left-1/4 w-96 h-96 bg-primary/8 rounded-full blur-3xl" />
        <div className="absolute bottom-0 right-1/4 w-72 h-72 bg-accent-cyan/6 rounded-full blur-3xl" />

        <div className="relative text-center max-w-lg mx-auto">
          <div className="animate-float mb-8 inline-block">
            <div className="w-20 h-20 rounded-2xl bg-primary/15 border border-primary/25 flex items-center justify-center mx-auto shadow-lg shadow-primary/10">
              <CommandLineIcon
                className="w-10 h-10 text-primary"
                strokeWidth={1.2}
              />
            </div>
          </div>

          <p className="text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-3 animate-fade-in">
            Welcome to
          </p>
          <h1
            className="text-3xl font-bold text-text-primary mb-3 animate-fade-in"
            style={{ animationDelay: "100ms" }}
          >
            {namespaceName}
          </h1>
          <p
            className="text-sm text-text-muted animate-fade-in"
            style={{ animationDelay: "200ms" }}
          >
            Your namespace is ready. Let&apos;s connect your first device.
          </p>

          {/* Waiting indicator */}
          <div
            className="mt-6 inline-flex items-center gap-2.5 bg-card/80 border border-border rounded-full px-4 py-2 animate-fade-in"
            style={{ animationDelay: "300ms" }}
          >
            <span className="relative flex h-2.5 w-2.5">
              <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-accent-yellow opacity-75" />
              <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-accent-yellow" />
            </span>
            <span className="text-2xs font-mono text-text-muted">
              Waiting for first device...
            </span>
          </div>
        </div>
      </div>

      {/* Steps */}
      <div className="px-8 pb-12">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-4xl mx-auto">
          {steps.map((step, idx) => (
            <div
              key={step.num}
              className="group relative bg-card border border-border rounded-xl p-6 hover:border-primary/30 transition-all duration-300 animate-slide-up overflow-hidden"
              style={{ animationDelay: `${400 + idx * 100}ms` }}
            >
              <div className="shimmer absolute inset-0 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary group-hover:bg-primary/15 group-hover:border-primary/30 transition-all duration-300">
                    {step.icon}
                  </div>
                  <span className="text-2xs font-mono font-bold text-text-muted/40">
                    {step.num}
                  </span>
                </div>

                <h3 className="text-sm font-semibold text-text-primary mb-1.5 group-hover:text-primary transition-colors">
                  {step.title}
                </h3>
                <p className="text-xs text-text-muted leading-relaxed mb-4">
                  {step.description}
                </p>

                {step.hasCommand && (
                  <div className="relative">
                    <div className="bg-background border border-border rounded-lg p-3 pr-10 font-mono text-2xs text-text-secondary break-all leading-relaxed">
                      <span className="text-primary/60">$ </span>
                      {installCmd}
                    </div>
                    <button
                      onClick={handleCopy}
                      className="absolute top-2 right-2 p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
                      title="Copy command"
                    >
                      {copied ? (
                        <CheckIcon
                          className="w-3.5 h-3.5 text-accent-green"
                          strokeWidth={2}
                        />
                      ) : (
                        <ClipboardDocumentIcon
                          className="w-3.5 h-3.5"
                          strokeWidth={2}
                        />
                      )}
                    </button>
                  </div>
                )}

                {step.linkTo && (
                  <Link
                    to={step.linkTo}
                    className="inline-flex items-center gap-1.5 text-xs font-medium text-primary hover:text-primary-400 transition-colors"
                  >
                    {step.linkLabel}
                    <ArrowRightIcon
                      className="w-3 h-3 group-hover:translate-x-0.5 transition-transform"
                      strokeWidth={2.5}
                    />
                  </Link>
                )}
              </div>
            </div>
          ))}
        </div>

        {/* Footer links */}
        <div
          className="flex items-center justify-center gap-6 mt-10 animate-fade-in"
          style={{ animationDelay: "800ms" }}
        >
          <a
            href="https://docs.shellhub.io"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <BookOpenIcon className="w-3.5 h-3.5" />
            Documentation
          </a>
          <span className="w-px h-3 bg-border" />
          <a
            href="https://github.com/shellhub-io/shellhub"
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
          >
            <svg
              className="w-3.5 h-3.5"
              fill="currentColor"
              viewBox="0 0 24 24"
            >
              <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
            </svg>
            Community
          </a>
        </div>
      </div>
    </div>
  );
}
