import { Link } from "react-router-dom";
import {
  CpuChipIcon,
  UserGroupIcon,
  KeyIcon,
  CommandLineIcon,
  ArrowRightIcon,
  BookOpenIcon,
} from "@heroicons/react/24/outline";
import { ConnectionGrid, GlowOrbs } from "@shellhub/design-system/components";
import { GithubIcon, IconBadge } from "@shellhub/design-system/primitives";

interface WelcomeScreenProps {
  namespaceName: string;
}

const steps = [
  {
    num: "01",
    title: "Connect your first device",
    description:
      "Install the ShellHub agent on any Linux device to start managing it remotely.",
    icon: <CpuChipIcon className="w-6 h-6" />,
    linkTo: "/devices/add",
    linkLabel: "Add device",
  },
  {
    num: "02",
    title: "Add a public key",
    description:
      "Set up SSH public key authentication for secure, passwordless access.",
    icon: <KeyIcon className="w-6 h-6" />,
    linkTo: "/sshkeys/public-keys",
    linkLabel: "Manage keys",
  },
  {
    num: "03",
    title: "Invite your team",
    description: "Add members to collaborate and manage devices together.",
    icon: <UserGroupIcon className="w-6 h-6" />,
    linkTo: "/team",
    linkLabel: "Manage team",
  },
];

export default function WelcomeScreen({ namespaceName }: WelcomeScreenProps) {
  return (
    <div className="min-h-full relative overflow-hidden px-4 pb-12">
      {/* Hero */}
      <div className="relative pt-16 pb-12 overflow-hidden">
        <ConnectionGrid />
        <GlowOrbs preset="duo" tone="primary" />

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
      <>
        <ol className="grid grid-cols-1 md:grid-cols-3 gap-4 max-w-4xl mx-auto">
          {steps.map((step, idx) => (
            <li
              key={step.num}
              className="group relative bg-card border border-border rounded-xl p-6 hover:border-primary/30 transition-all duration-300 animate-slide-up overflow-hidden"
              style={{ animationDelay: `${400 + idx * 100}ms` }}
            >
              <div className="shimmer absolute inset-0 rounded-xl opacity-0 group-hover:opacity-100 transition-opacity duration-500" />

              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <IconBadge
                    size="md"
                    color="primary"
                    className="group-hover:bg-primary/15 group-hover:border-primary/30 transition-all duration-300"
                  >
                    {step.icon}
                  </IconBadge>
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
            </li>
          ))}
        </ol>

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
            <GithubIcon className="w-3.5 h-3.5" />
            Community
          </a>
        </div>
      </>
    </div>
  );
}
