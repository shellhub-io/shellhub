import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ------------------------------------------------------------------ */
/*  Data                                                               */
/* ------------------------------------------------------------------ */

const painPoints = [
  {
    title: "VPN tunnels in CI pipelines",
    desc: "Establishing VPN connections from ephemeral CI runners to target devices adds latency, complexity, and fragile points of failure to every deployment.",
    color: C.primary,
  },
  {
    title: "SSH key sprawl",
    desc: "Distributing, rotating, and revoking SSH keys across CI systems, bastion hosts, and hundreds of target devices is a security and operational burden.",
    color: C.yellow,
  },
  {
    title: "No stable device addressing",
    desc: "Devices behind NAT or CGNAT have no fixed addresses. Ansible inventories and Terraform providers simply cannot reach them.",
    color: C.red,
  },
  {
    title: "Limited deployment visibility",
    desc: "When automated pipelines SSH into devices there is often no centralized log of what ran where -- making debugging failed rollouts painful.",
    color: C.cyan,
  },
];

const features = [
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <polyline points="4 17 10 11 4 5" />
        <line x1="12" y1="19" x2="20" y2="19" />
      </svg>
    ),
    color: C.primary,
    title: "Standard SSH Transport",
    desc: "Ansible, Terraform, and CI/CD tools connect through ShellHub using standard SSH -- no custom plugins or connectors required.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round">
        <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z" />
        <polyline points="14 2 14 8 20 8" />
        <line x1="16" y1="13" x2="8" y2="13" />
        <line x1="16" y1="17" x2="8" y2="17" />
      </svg>
    ),
    color: C.cyan,
    title: "SCP/SFTP Transfers",
    desc: "Push build artifacts, configuration files, and firmware updates to remote devices as part of your deployment pipeline.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 20h9" />
        <path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z" />
      </svg>
    ),
    color: C.green,
    title: "Audit Logging",
    desc: "Every automated session is recorded -- see exactly which pipeline ran what command on which device and when.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round">
        <path d="M20.59 13.41l-7.17 7.17a2 2 0 01-2.83 0L2 12V2h10l8.59 8.59a2 2 0 010 2.82z" />
        <line x1="7" y1="7" x2="7.01" y2="7" />
      </svg>
    ),
    color: C.yellow,
    title: "Device Tags",
    desc: "Target deployments to specific device groups using tags -- deploy to 'staging' or 'production' fleets with a single inventory.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      </svg>
    ),
    color: C.primary,
    title: "Firewall Rules",
    desc: "Restrict CI runner access to only the devices they need with IP-based and identity-based firewall policies.",
  },
  {
    icon: (
      <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round">
        <path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z" />
      </svg>
    ),
    color: C.green,
    title: "Namespaces",
    desc: "Isolate staging and production environments in separate namespaces with independent access policies and device groups.",
  },
];

/* ------------------------------------------------------------------ */
/*  Ansible terminal lines                                             */
/* ------------------------------------------------------------------ */

const ansibleLines: { text: string; color?: string; dim?: boolean; indent?: boolean }[] = [
  { text: "$ ansible-playbook -i inventory deploy.yml", color: C.green },
  { text: "" },
  { text: "PLAY [Deploy to edge devices] ***********************", color: C.cyan },
  { text: "" },
  { text: "TASK [Gathering Facts] ******************************", color: C.cyan },
  { text: "ok: [device-01.shellhub]", color: C.green, indent: true },
  { text: "ok: [device-02.shellhub]", color: C.green, indent: true },
  { text: "" },
  { text: "TASK [Update firmware] ******************************", color: C.cyan },
  { text: "changed: [device-01.shellhub]", color: C.yellow, indent: true },
  { text: "changed: [device-02.shellhub]", color: C.yellow, indent: true },
  { text: "" },
  { text: "TASK [Restart service] ******************************", color: C.cyan },
  { text: "ok: [device-01.shellhub]", color: C.green, indent: true },
  { text: "ok: [device-02.shellhub]", color: C.green, indent: true },
  { text: "" },
  { text: "PLAY RECAP ******************************************", color: C.cyan },
  { text: "device-01.shellhub  : ok=3  changed=1  failed=0", color: C.green, indent: true },
  { text: "device-02.shellhub  : ok=3  changed=1  failed=0", color: C.green, indent: true },
];

/* ------------------------------------------------------------------ */
/*  Terraform HCL lines                                                */
/* ------------------------------------------------------------------ */

const terraformLines: { text: string; color?: string; dim?: boolean }[] = [
  { text: 'resource "shellhub_device" "edge_fleet" {', color: C.primary },
  { text: '  count      = var.fleet_size', dim: true },
  { text: '  namespace  = "production"', dim: true },
  { text: '  hostname   = "edge-${count.index + 1}"', dim: true },
  { text: "" },
  { text: "  tags = {", dim: true },
  { text: '    environment = "production"', color: C.green },
  { text: '    region      = var.region', color: C.green },
  { text: '    managed_by  = "terraform"', color: C.green },
  { text: "  }", dim: true },
  { text: "" },
  { text: "  connection {", dim: true },
  { text: '    type = "ssh"', color: C.cyan },
  { text: '    host = "${self.sshid}@shellhub.io"', color: C.cyan },
  { text: "  }", dim: true },
  { text: "}", color: C.primary },
];

/* ------------------------------------------------------------------ */
/*  CI / CD pipeline steps                                             */
/* ------------------------------------------------------------------ */

const pipelineSteps = [
  { label: "Build", status: "done" as const },
  { label: "Test", status: "done" as const },
  { label: "Deploy via ShellHub", status: "done" as const },
];

const pipelineDeployLines: { text: string; color?: string; dim?: boolean }[] = [
  { text: "deploy:", color: C.primary },
  { text: "  runs-on: ubuntu-latest", dim: true },
  { text: "  steps:", dim: true },
  { text: '    - name: Deploy to edge fleet', color: C.cyan },
  { text: "      run: |", dim: true },
  { text: '        ssh user@device-01.prod@shellhub.io \\', color: C.green },
  { text: '          "cd /opt/app && git pull && make restart"', color: C.green },
  { text: '        ssh user@device-02.prod@shellhub.io \\', color: C.green },
  { text: '          "cd /opt/app && git pull && make restart"', color: C.green },
];

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

function TerminalChrome({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="p-6">
      <div className="flex items-center gap-2 mb-5">
        <div className="w-3 h-3 rounded-full bg-accent-red/60" />
        <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
        <div className="w-3 h-3 rounded-full bg-accent-green/60" />
        <span className="ml-2 text-2xs text-text-muted font-mono">{title}</span>
      </div>
      {children}
    </div>
  );
}

function CodeLine({ text, color, dim, indent }: { text: string; color?: string; dim?: boolean; indent?: boolean }) {
  if (text === "") return <div className="h-3" />;
  return (
    <div
      className="font-mono text-2xs leading-relaxed whitespace-pre"
      style={{
        color: dim ? C.textMuted : color || C.text,
        paddingLeft: indent ? "1rem" : undefined,
      }}
    >
      {text}
    </div>
  );
}

/* ------------------------------------------------------------------ */
/*  Page                                                               */
/* ------------------------------------------------------------------ */

export default function DevopsCiCd() {
  const [navSolid, setNavSolid] = useState(false);
  const [mobileMenu, setMobileMenu] = useState(false);
  const handleScroll = useCallback(() => { setNavSolid(window.scrollY > 50); }, []);
  useEffect(() => {
    window.addEventListener("scroll", handleScroll, { passive: true });
    return () => window.removeEventListener("scroll", handleScroll);
  }, [handleScroll]);

  return (
    <div className="min-h-screen bg-background text-text-primary font-sans antialiased overflow-x-hidden">
      <Navbar navSolid={navSolid} mobileMenu={mobileMenu} setMobileMenu={setMobileMenu} />

      {/* ── Hero ─────────────────────────────────────────────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 right-1/4 w-96 h-96 bg-primary/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full mb-6">
              Use Case
            </span>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-4xl mx-auto">
              DevOps & CI/CD{" "}
              <span className="bg-gradient-to-r from-primary via-accent-blue to-accent-cyan bg-clip-text text-transparent">
                Automation
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              Automate deployments to remote and edge devices with Ansible, Terraform, and CI/CD pipelines -- using standard SSH through ShellHub.
            </p>
          </Reveal>
          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <a
                href="/v2/getting-started"
                className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
              >
                <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                <span className="relative">Get Started Free</span>
                <svg className="relative w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                </svg>
              </a>
              <a
                href="/v2/docs"
                className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
              >
                Read the Docs
              </a>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ── Ansible Integration ──────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <div>
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Ansible Integration
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Use ShellHub as your SSH transport for Ansible
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  Point your Ansible inventory at ShellHub device identifiers and run playbooks against devices behind NAT, CGNAT, or firewalls -- no VPN required. ShellHub acts as a transparent SSH gateway so existing playbooks work without modification.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Standard SSH connection", desc: "No custom connection plugins -- Ansible uses its built-in SSH transport" },
                  { label: "NAT traversal included", desc: "Reach devices that have no public IP or inbound ports" },
                  { label: "Fleet targeting with tags", desc: "Build dynamic inventories based on ShellHub device tags" },
                  { label: "Session-level audit trail", desc: "Every Ansible task is logged with full session recording" },
                ].map((item, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{item.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{item.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>

            <Reveal delay={0.1}>
              <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                <TerminalChrome title="ansible-playbook">
                  <div className="space-y-0">
                    {ansibleLines.map((line, i) => (
                      <CodeLine key={i} {...line} />
                    ))}
                  </div>
                </TerminalChrome>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ── Terraform Integration ────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            <Reveal delay={0.1} className="order-2 lg:order-1">
              <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
                <TerminalChrome title="main.tf">
                  <div className="space-y-0">
                    {terraformLines.map((line, i) => (
                      <CodeLine key={i} {...line} />
                    ))}
                  </div>
                </TerminalChrome>
              </ShimmerCard>
            </Reveal>

            <div className="order-1 lg:order-2">
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Terraform Integration
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Infrastructure as code for device provisioning
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  Define your device fleet declaratively with Terraform. Provision namespaces, assign tags, configure firewall rules, and manage access policies -- all versioned in Git and applied through your standard IaC workflow.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Declarative device management", desc: "Define fleet topology and access policies as HCL resources" },
                  { label: "GitOps-friendly", desc: "Version your device configuration alongside application code" },
                  { label: "Drift detection", desc: "Terraform plan shows configuration differences before applying" },
                  { label: "Reproducible environments", desc: "Spin up identical staging and production fleets from the same config" },
                ].map((item, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{item.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{item.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* ── CI/CD Pipeline ───────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              CI/CD Pipeline
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Deploy to remote devices from any CI system
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              GitHub Actions, GitLab CI, Jenkins -- any pipeline that can run SSH can deploy through ShellHub.
            </p>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="max-w-3xl mx-auto">
              <div className="relative bg-card border border-primary/30 rounded-xl overflow-hidden shadow-[0_0_40px_rgba(102,122,204,0.1)]">
                <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-transparent pointer-events-none" />
                <div className="relative">
                  <TerminalChrome title=".github/workflows/deploy.yml">
                    {/* Pipeline step indicators */}
                    <div className="flex items-center gap-6 mb-6 pb-5 border-b border-border">
                      {pipelineSteps.map((step, i) => (
                        <div key={i} className="flex items-center gap-2">
                          <div className="w-5 h-5 rounded-full bg-accent-green/15 border border-accent-green/30 flex items-center justify-center">
                            <svg className="w-3 h-3 text-accent-green" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                              <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                            </svg>
                          </div>
                          <span className="text-2xs font-mono text-text-secondary">{step.label}</span>
                          {i < pipelineSteps.length - 1 && (
                            <svg className="w-3 h-3 text-text-muted ml-2" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                              <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                            </svg>
                          )}
                        </div>
                      ))}
                    </div>

                    {/* Workflow YAML */}
                    <div className="space-y-0 mb-6">
                      {pipelineDeployLines.map((line, i) => (
                        <CodeLine key={i} {...line} />
                      ))}
                    </div>

                    {/* Status bar */}
                    <div className="pt-4 border-t border-border flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <div className="w-2 h-2 rounded-full bg-accent-green animate-pulse" />
                        <span className="text-2xs font-mono text-accent-green">Pipeline succeeded</span>
                      </div>
                      <span className="text-2xs text-text-muted font-mono">2 devices updated in 34s</span>
                    </div>
                  </TerminalChrome>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ── Pain Points ──────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              The Problem
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Why automating remote deployments is hard
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Traditional approaches to device automation break down when devices are behind NAT, on unstable networks, or spread across locations.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {painPoints.map((p, i) => (
              <Reveal key={i} delay={i * 0.06}>
                <ShimmerCard className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                  <div className="w-2 h-2 rounded-full mb-4" style={{ background: p.color }} />
                  <h4 className="text-sm font-semibold mb-2">{p.title}</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">{p.desc}</p>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── Key Features ─────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Capabilities
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              ShellHub features for DevOps teams
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Everything your automation toolchain needs to securely deploy to remote and edge devices at scale.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {features.map((f, i) => (
              <Reveal key={i} delay={i * 0.04}>
                <ShimmerCard className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                  <div
                    className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                    style={{ background: `${f.color}15`, borderColor: `${f.color}25` }}
                  >
                    {f.icon}
                  </div>
                  <h4 className="text-sm font-semibold mb-2">{f.title}</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">{f.desc}</p>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA ──────────────────────────────────────────────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal>
            <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
              <ConnectionGrid />
              <div className="absolute inset-0 bg-gradient-to-br from-primary/[0.06] via-transparent to-accent-cyan/[0.04] pointer-events-none" />

              <div className="relative z-10">
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Ready to automate?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Automate your device deployments today
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Connect ShellHub to your CI/CD pipeline and start deploying to remote devices in minutes -- no VPN, no static IPs, no hassle.
                </p>

                <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                  <a
                    href="/v2/getting-started"
                    className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
                  >
                    <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                    <span className="relative">Get Started Free</span>
                    <svg className="relative w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                    </svg>
                  </a>
                  <a
                    href="/v2/docs"
                    className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
                  >
                    Read the Docs
                  </a>
                </div>
              </div>
            </div>
          </Reveal>
        </div>
      </section>

      <Footer />
    </div>
  );
}
