import {
  ArchiveBoxIcon,
  ArrowRightIcon,
  ArrowTopRightOnSquareIcon,
  CommandLineIcon,
  ComputerDesktopIcon,
  CpuChipIcon,
  CubeIcon,
  ServerStackIcon,
  WrenchScrewdriverIcon,
} from "@heroicons/react/24/outline";
import { DockerIcon } from "@shellhub/design-system/primitives";
import { Section, SectionHeader } from "@/components/marketing";
import { Reveal } from "@shellhub/design-system/components";
import { docsUrl } from "@/links";

export function SupportedPlatforms() {
  return (
    <Section className="relative overflow-hidden" container={false}>
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[700px] h-[700px] bg-primary/[0.03] rounded-full blur-3xl pointer-events-none" />
      <div className="max-w-7xl mx-auto px-8 relative z-10">
        <SectionHeader
          eyebrow="Supported Platforms"
          title={
            <>
              One agent. <span className="text-primary">Every platform.</span>
            </>
          }
          subtitle={
            <>
              The ShellHub agent runs on{" "}
              <span className="inline-block px-1.5 py-0.5 bg-border/40 rounded text-2xs font-mono font-medium text-text-primary/90">
                x86
              </span>{" "}
              and{" "}
              <span className="inline-block px-1.5 py-0.5 bg-border/40 rounded text-2xs font-mono font-medium text-text-primary/90">
                ARM
              </span>{" "}
              — from Docker containers to embedded Linux and FreeBSD.
            </>
          }
          subtitleClassName="max-w-xl"
        />

        {/* Bento Grid — 12-col for precise control */}
        <div className="grid grid-cols-2 lg:grid-cols-12 gap-3 auto-rows-auto">
          {/* Docker — featured (5 cols, 2 rows) */}
          <Reveal className="col-span-2 lg:col-span-5 lg:row-span-2">
            <div className="h-full bg-card border border-border rounded-xl p-6 lg:p-7 hover:border-[#1D63ED]/30 transition-all duration-300 group relative overflow-hidden">
              <div className="absolute -right-6 -bottom-6 w-32 h-32 bg-[#1D63ED]/[0.04] rounded-full blur-2xl pointer-events-none" />
              <div className="relative">
                <div className="flex items-center gap-3 mb-5">
                  <div className="w-10 h-10 rounded-lg flex items-center justify-center bg-[#1D63ED]/10 border border-[#1D63ED]/20 text-[#1D63ED]">
                    <DockerIcon className="w-5 h-5" />
                  </div>
                  <span className="text-2xs font-mono font-semibold text-[#1D63ED] uppercase tracking-[0.1em]">
                    Most Popular
                  </span>
                </div>
                <h3 className="text-xl font-bold tracking-[-0.02em] mb-2">
                  Docker
                </h3>
                <p className="text-sm text-text-secondary leading-relaxed mb-5">
                  Run the agent as a container. The recommended way to get
                  started.
                </p>
                <div className="bg-surface border border-border rounded-lg px-4 py-3">
                  <code className="font-mono text-xs text-text-secondary">
                    <span className="text-primary">$</span> curl -sSf
                    &lt;server&gt;/install.sh | sh
                  </code>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Snap — wider card (4 cols) */}
          <Reveal delay={0.04} className="col-span-1 lg:col-span-4">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-green/30 transition-all duration-300">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-lg flex shrink-0 items-center justify-center bg-accent-green/10 border border-accent-green/20 text-accent-green">
                  <ArchiveBoxIcon className="w-4 h-4" />
                </div>
                <div>
                  <h4 className="text-sm font-bold mb-1">Snap</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    Install from the Snap Store with automatic updates and
                    rollback support.
                  </p>
                </div>
              </div>
            </div>
          </Reveal>

          {/* Podman (3 cols) */}
          <Reveal delay={0.08} className="col-span-1 lg:col-span-3">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-primary/30 transition-all duration-300">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-primary/10 border border-primary/20 text-primary mb-3">
                <CubeIcon className="w-4 h-4" />
              </div>
              <h4 className="text-sm font-bold mb-1">Podman</h4>
              <p className="text-xs text-text-secondary leading-relaxed">
                Same container workflow as Docker, but rootless and daemonless.
              </p>
            </div>
          </Reveal>

          {/* Standalone (4 cols) */}
          <Reveal delay={0.12} className="col-span-1 lg:col-span-4">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-yellow/30 transition-all duration-300">
              <div className="flex items-start gap-3">
                <div className="w-9 h-9 rounded-lg flex shrink-0 items-center justify-center bg-accent-yellow/10 border border-accent-yellow/20 text-accent-yellow">
                  <ServerStackIcon className="w-4 h-4" />
                </div>
                <div>
                  <h4 className="text-sm font-bold mb-1">Standalone</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    Direct binary managed by systemd. No container runtime
                    needed.
                  </p>
                </div>
              </div>
            </div>
          </Reveal>

          {/* WSL (3 cols) */}
          <Reveal delay={0.16} className="col-span-1 lg:col-span-3">
            <div className="h-full bg-card border border-border rounded-xl p-5 hover:border-accent-cyan/30 transition-all duration-300">
              <div className="w-9 h-9 rounded-lg flex items-center justify-center bg-accent-cyan/10 border border-accent-cyan/20 text-accent-cyan mb-3">
                <ComputerDesktopIcon className="w-4 h-4" />
              </div>
              <h4 className="text-sm font-bold mb-1">WSL</h4>
              <p className="text-xs text-text-secondary leading-relaxed">
                Run the agent inside WSL to manage Windows dev environments.
              </p>
            </div>
          </Reveal>

          {/* ─── Embedded row — full width, horizontal strip ─── */}
          <Reveal delay={0.2} className="col-span-2 lg:col-span-12">
            <div className="grid grid-cols-3 gap-3">
              {[
                {
                  name: "Yocto Project",
                  sub: "Integrate via our official meta-layer",
                  color: "accent-cyan",
                  href: "https://github.com/shellhub-io/meta-shellhub",
                  icon: CpuChipIcon,
                },
                {
                  name: "Buildroot",
                  sub: "Add via our external tree package",
                  color: "accent-yellow",
                  href: "https://github.com/shellhub-io/buildroot",
                  icon: WrenchScrewdriverIcon,
                },
                {
                  name: "FreeBSD",
                  sub: "Runs natively with a dedicated ports entry",
                  color: "accent-red",
                  href: "https://github.com/shellhub-io/ports",
                  icon: CommandLineIcon,
                },
              ].map((p) => {
                const Icon = p.icon;
                return (
                  <a
                    key={p.name}
                    href={p.href}
                    target="_blank"
                    rel="noopener noreferrer"
                    aria-label={p.name}
                    className="bg-surface border border-border rounded-xl p-4 hover:border-border-light transition-all duration-300 group"
                  >
                    <div className="flex items-center gap-3">
                      <div
                        className={`w-8 h-8 rounded-md flex items-center justify-center bg-${p.color}/10 border border-${p.color}/20 text-${p.color}`}
                      >
                        <Icon className="w-4 h-4" aria-hidden="true" />
                      </div>
                      <div className="flex-1">
                        <h4 className="text-sm font-bold leading-tight group-hover:text-text-primary transition-colors">
                          {p.name}
                        </h4>
                        <span className="text-2xs text-text-secondary">
                          {p.sub}
                        </span>
                      </div>
                      <ArrowTopRightOnSquareIcon
                        className="w-3.5 h-3.5 text-text-muted group-hover:text-text-secondary transition-colors shrink-0"
                        strokeWidth={2}
                        aria-hidden="true"
                      />
                    </div>
                  </a>
                );
              })}
            </div>
          </Reveal>
        </div>

        <Reveal className="text-center mt-10">
          <a
            href={docsUrl}
            target="_blank"
            rel="noopener noreferrer"
            className="inline-flex items-center gap-1.5 text-xs font-medium text-primary hover:gap-2.5 transition-all group"
          >
            View all supported platforms
            <ArrowRightIcon
              className="w-3 h-3 group-hover:translate-x-0.5 transition-transform"
              strokeWidth={2.5}
              aria-hidden="true"
            />
          </a>
        </Reveal>
      </div>
    </Section>
  );
}
