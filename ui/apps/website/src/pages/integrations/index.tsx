import { Link } from "react-router-dom";
import {
  ArrowDownIcon,
  ArrowsRightLeftIcon,
  CloudIcon,
  CodeBracketIcon,
  CpuChipIcon,
  CubeIcon,
  DocumentIcon,
  CheckIcon,
  FolderIcon,
  LinkIcon,
  ShieldCheckIcon,
  UserIcon,
} from "@heroicons/react/24/outline";
import {
  Badge,
  Button,
  Card,
  IconBadge,
  WindowChrome,
} from "@shellhub/design-system/primitives";
import { ArrowRight } from "@/components/ArrowRight";
import { SiteLayout } from "@/components/SiteLayout";
import { Section, SectionHeader } from "@/components/marketing";
import { docsUrl } from "@/links";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ------------------------------------------------------------------ */
/*  Syntax-highlighted line helpers                                    */
/* ------------------------------------------------------------------ */
function Ln({
  color,
  children,
}: {
  color?: string;
  children: React.ReactNode;
}) {
  return <div style={{ color: color ?? C.text }}>{children}</div>;
}
function Comment({ children }: { children: React.ReactNode }) {
  return <Ln color={C.textMuted}>{children}</Ln>;
}
function Kw({ children }: { children: React.ReactNode }) {
  return <span style={{ color: C.primary }}>{children}</span>;
}
function Str({ children }: { children: React.ReactNode }) {
  return <span style={{ color: C.green }}>{children}</span>;
}
function Val({ children }: { children: React.ReactNode }) {
  return <span style={{ color: C.yellow }}>{children}</span>;
}
function Dim({ children }: { children: React.ReactNode }) {
  return <span style={{ color: C.textMuted }}>{children}</span>;
}
function Cyn({ children }: { children: React.ReactNode }) {
  return <span style={{ color: C.cyan }}>{children}</span>;
}

/* ================================================================== */
/*  Page                                                               */
/* ================================================================== */
export default function Integrations() {
  return (
    <SiteLayout>
      {/* ─────────── Hero ─────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 left-1/4 w-96 h-96 bg-accent-cyan/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <Badge shape="pill" color="cyan" className="mb-6 tracking-label">
              Integrations
            </Badge>
          </Reveal>
          <Reveal>
            <h1 className="text-[clamp(2rem,5vw,3.5rem)] font-bold tracking-[-0.03em] leading-[1.1] mb-6 max-w-3xl mx-auto">
              Works with your{" "}
              <span className="bg-gradient-to-r from-accent-cyan via-primary to-accent-blue bg-clip-text text-transparent">
                existing tools
              </span>
            </h1>
          </Reveal>
          <Reveal>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed mb-10">
              ShellHub speaks standard SSH. Any tool that connects over SSH
              works out of the box &mdash; no plugins, no custom agents, no
              vendor lock-in.
            </p>
          </Reveal>
          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <Button
                as="a"
                href={docsUrl}
                target="_blank"
                rel="noopener noreferrer"
                variant="primary"
                size="xl"
                glow
                iconRight={<ArrowRight />}
              >
                Browse Docs
              </Button>
              <Button
                as={Link}
                to="/getting-started"
                variant="outline"
                size="xl"
              >
                Get Started Free
              </Button>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ─────────── Automation & IaC ─────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left copy */}
          <div>
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrow="Automation & Infrastructure as Code"
              title="Ansible & Terraform, zero VPN"
              subtitle="Use ShellHub as the SSH transport for your Ansible playbooks and Terraform provisioners. Manage devices behind NAT, firewalls, or cellular networks without VPN tunnels or bastion hosts."
            />

            <div className="space-y-3">
              {[
                {
                  label: "Drop-in SSH replacement",
                  desc: "Set ProxyCommand once and every tool uses ShellHub transparently",
                },
                {
                  label: "NAT traversal built in",
                  desc: "Reach devices behind CGNAT, firewalls, and private networks",
                },
                {
                  label: "Ansible connection plugin",
                  desc: "Native Ansible connection plugin for seamless integration",
                },
                {
                  label: "Terraform provisioner",
                  desc: "Run remote-exec provisioners through ShellHub SSH",
                },
              ].map((cap, i) => (
                <Reveal key={i} delay={i * 0.04}>
                  <div className="flex items-start gap-3">
                    <CheckIcon
                      className="w-4 h-4 text-accent-green shrink-0 mt-0.5"
                      aria-hidden="true"
                    />
                    <div>
                      <p className="text-sm font-medium text-text-primary">
                        {cap.label}
                      </p>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        {cap.desc}
                      </p>
                    </div>
                  </div>
                </Reveal>
              ))}
            </div>
          </div>

          {/* Right terminal mockup */}
          <Reveal delay={0.1}>
            <ShimmerCard>
              <WindowChrome
                variant="terminal"
                title="ansible-playbook"
                bodyClassName="overflow-x-auto"
              >
                <Comment># Deploy firmware update via ShellHub SSH</Comment>
                <Ln>
                  <Val>$</Val> ansible-playbook -i inventory.yml deploy.yml
                </Ln>
                <div className="mt-3" />
                <Ln color={C.textSec}>
                  PLAY [Update edge devices] ****************************
                </Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>
                  TASK [Gathering Facts] ********************************
                </Ln>
                <Ln>
                  <Str>ok</Str>: [gateway-east-01.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Str>ok</Str>: [sensor-rack-07.d0a1c2.<Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Str>ok</Str>: [plc-floor-03.d0a1c2.<Cyn>shellhub.io</Cyn>]
                </Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>
                  TASK [Copy firmware binary] ****************************
                </Ln>
                <Ln>
                  <Val>changed</Val>: [gateway-east-01.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Val>changed</Val>: [sensor-rack-07.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Val>changed</Val>: [plc-floor-03.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>
                  TASK [Apply update &amp; restart] **************************
                </Ln>
                <Ln>
                  <Val>changed</Val>: [gateway-east-01.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Val>changed</Val>: [sensor-rack-07.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <Ln>
                  <Val>changed</Val>: [plc-floor-03.d0a1c2.
                  <Cyn>shellhub.io</Cyn>]
                </Ln>
                <div className="mt-3" />
                <Ln color={C.textSec}>
                  PLAY RECAP ********************************************
                </Ln>
                <Ln>
                  gateway-east-01 : <Str>ok=3</Str> <Val>changed=2</Val>{" "}
                  unreachable=0 failed=<Str>0</Str>
                </Ln>
                <Ln>
                  sensor-rack-07 &nbsp;: <Str>ok=3</Str> <Val>changed=2</Val>{" "}
                  unreachable=0 failed=<Str>0</Str>
                </Ln>
                <Ln>
                  plc-floor-03 &nbsp;&nbsp;&nbsp;: <Str>ok=3</Str>{" "}
                  <Val>changed=2</Val> unreachable=0 failed=<Str>0</Str>
                </Ln>
              </WindowChrome>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ─────────── CI/CD Pipelines ─────────── */}
      <Section>
        <SectionHeader
          eyebrow="CI/CD Pipelines"
          title="Deploy from any CI system"
          subtitle={
            <>
              GitHub Actions, GitLab CI, Jenkins, CircleCI &mdash; if it can run
              SSH, it can deploy through ShellHub. No custom plugins required.
            </>
          }
        />

        <Reveal delay={0.1}>
          <ShimmerCard className="max-w-3xl mx-auto">
            <WindowChrome
              variant="terminal"
              title=".github/workflows/deploy.yml"
              titleBarSlot={
                <Badge shape="pill" color="green">
                  Passing
                </Badge>
              }
              className="border-accent-green/25 shadow-[0_0_40px_rgba(130,165,104,0.08)]"
              bodyClassName="overflow-x-auto"
            >
              <Ln>
                <Kw>name</Kw>: <Str>Deploy to Edge Devices</Str>
              </Ln>
              <Ln>
                <Kw>on</Kw>:
              </Ln>
              <Ln>
                {" "}
                <Kw>push</Kw>:
              </Ln>
              <Ln>
                {" "}
                <Kw>branches</Kw>: [<Str>main</Str>]
              </Ln>
              <div className="mt-2" />
              <Ln>
                <Kw>jobs</Kw>:
              </Ln>
              <Ln>
                {" "}
                <Kw>deploy</Kw>:
              </Ln>
              <Ln>
                {" "}
                <Kw>runs-on</Kw>: <Str>ubuntu-latest</Str>
              </Ln>
              <Ln>
                {" "}
                <Kw>steps</Kw>:
              </Ln>
              <Ln>
                {" "}
                - <Kw>name</Kw>: <Str>Configure ShellHub SSH</Str>
              </Ln>
              <Ln>
                {" "}
                <Kw>run</Kw>: |
              </Ln>
              <Ln>
                {" "}
                <Dim>mkdir -p ~/.ssh</Dim>
              </Ln>
              <Ln>
                {" "}
                <Dim>
                  echo <Str>"$&#123;&#123; secrets.SSH_KEY &#125;&#125;"</Str>{" "}
                  &gt; ~/.ssh/id_rsa
                </Dim>
              </Ln>
              <Ln>
                {" "}
                <Dim>chmod 600 ~/.ssh/id_rsa</Dim>
              </Ln>
              <div className="mt-2" />
              <Ln>
                {" "}
                - <Kw>name</Kw>: <Str>Deploy application</Str>
              </Ln>
              <Ln>
                {" "}
                <Kw>run</Kw>: |
              </Ln>
              <Ln>
                {" "}
                <Dim>
                  ssh <Cyn>admin@device.namespace.shellhub.io</Cyn> \
                </Dim>
              </Ln>
              <Ln>
                {" "}
                <Dim>
                  {" "}
                  <Str>
                    "cd /opt/app &amp;&amp; git pull &amp;&amp; systemctl
                    restart app"
                  </Str>
                </Dim>
              </Ln>
              <div className="mt-2" />
              <Ln>
                {" "}
                - <Kw>name</Kw>: <Str>Verify deployment</Str>
              </Ln>
              <Ln>
                {" "}
                <Kw>run</Kw>: |
              </Ln>
              <Ln>
                {" "}
                <Dim>
                  ssh <Cyn>admin@device.namespace.shellhub.io</Cyn> \
                </Dim>
              </Ln>
              <Ln>
                {" "}
                <Dim>
                  {" "}
                  <Str>"curl -sf http://localhost:8080/health"</Str>
                </Dim>
              </Ln>
            </WindowChrome>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ─────────── Development Tools ─────────── */}
      <Section>
        <SectionHeader
          eyebrow="Development Tools"
          title="Develop directly on remote devices"
          subtitle="Connect your favorite editor or build system to devices through ShellHub. Full IDE support with zero configuration changes."
        />

        <div className="grid md:grid-cols-2 gap-6">
          {/* VS Code Remote SSH */}
          <Reveal delay={0}>
            <ShimmerCard className="h-full">
              <div className="relative bg-card border border-accent-blue/25 rounded-xl overflow-hidden h-full hover:border-accent-blue/40 transition-all duration-300 shadow-[0_0_40px_rgba(86,162,225,0.06)]">
                <div className="absolute inset-0 bg-gradient-to-br from-accent-blue/[0.04] via-transparent to-transparent pointer-events-none" />
                <div className="relative">
                  {/* VS Code title bar */}
                  <div className="flex items-center gap-2 px-4 py-2.5 border-b border-border bg-[#1E1E2E]">
                    <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                    <span className="ml-2 text-2xs text-text-muted font-mono">
                      VS Code &mdash; Remote SSH
                    </span>
                    <div className="ml-auto">
                      <span className="px-2 py-0.5 text-2xs font-mono bg-accent-blue/10 text-accent-blue border border-accent-blue/20 rounded-full">
                        SSH: gateway-east-01
                      </span>
                    </div>
                  </div>

                  <div className="flex min-h-[240px]">
                    {/* Sidebar */}
                    <div className="w-44 shrink-0 border-r border-border bg-surface/40 p-3">
                      <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-2">
                        Explorer
                      </p>
                      <div className="space-y-1">
                        {(
                          [
                            {
                              name: "src/",
                              indent: 0,
                              isDir: true,
                              active: false,
                            },
                            {
                              name: "main.py",
                              indent: 1,
                              isDir: false,
                              active: true,
                            },
                            {
                              name: "config.yml",
                              indent: 1,
                              isDir: false,
                              active: false,
                            },
                            {
                              name: "utils/",
                              indent: 1,
                              isDir: true,
                              active: false,
                            },
                            {
                              name: "tests/",
                              indent: 0,
                              isDir: true,
                              active: false,
                            },
                            {
                              name: "Dockerfile",
                              indent: 0,
                              isDir: false,
                              active: false,
                            },
                            {
                              name: "requirements.txt",
                              indent: 0,
                              isDir: false,
                              active: false,
                            },
                          ] as const
                        ).map((f, i) => (
                          <div
                            key={i}
                            className={`flex items-center gap-1.5 px-1.5 py-0.5 rounded text-2xs font-mono ${f.active ? "bg-accent-blue/10 text-accent-blue" : "text-text-secondary hover:bg-white/[0.03]"}`}
                            style={{ paddingLeft: `${f.indent * 12 + 6}px` }}
                          >
                            {f.isDir ? (
                              <FolderIcon
                                className="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                            ) : (
                              <DocumentIcon
                                className="w-3 h-3 shrink-0"
                                aria-hidden="true"
                              />
                            )}
                            {f.name}
                          </div>
                        ))}
                      </div>
                    </div>

                    {/* Editor area */}
                    <div className="flex-1 p-4 font-mono text-2xs leading-[1.8]">
                      <div className="flex items-center gap-3 mb-3 border-b border-border pb-2">
                        <span className="px-2 py-0.5 text-2xs bg-accent-blue/10 text-accent-blue border-b-2 border-accent-blue rounded-t">
                          main.py
                        </span>
                        <span className="px-2 py-0.5 text-2xs text-text-muted">
                          config.yml
                        </span>
                      </div>
                      <div className="flex gap-3">
                        <div className="text-text-muted select-none text-right w-5 shrink-0">
                          {[1, 2, 3, 4, 5, 6, 7, 8].map((n) => (
                            <div key={n}>{n}</div>
                          ))}
                        </div>
                        <div>
                          <Ln>
                            <Kw>import</Kw> <Str>flask</Str>
                          </Ln>
                          <Ln>
                            <Kw>from</Kw> config <Kw>import</Kw> settings
                          </Ln>
                          <Ln color={C.textMuted}>&nbsp;</Ln>
                          <Ln>app = flask.Flask(__name__)</Ln>
                          <Ln color={C.textMuted}>&nbsp;</Ln>
                          <Ln>
                            <Kw>@app</Kw>.route(<Str>"/health"</Str>)
                          </Ln>
                          <Ln>
                            <Kw>def</Kw> <Val>health</Val>():
                          </Ln>
                          <Ln>
                            {" "}
                            <Kw>return</Kw> {"{"}
                            <Str>"status"</Str>: <Str>"ok"</Str>
                            {"}"}
                          </Ln>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Status bar */}
                  <div className="flex items-center justify-between px-4 py-1.5 bg-accent-blue/10 border-t border-border text-2xs font-mono">
                    <div className="flex items-center gap-3">
                      <span className="text-accent-blue flex items-center gap-1">
                        <LinkIcon className="w-3 h-3" aria-hidden="true" />
                        SSH: gateway-east-01
                      </span>
                      <span className="text-text-muted">Python 3.11</span>
                    </div>
                    <span className="text-text-muted">Ln 7, Col 12</span>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>

          {/* Embedded Linux */}
          <Reveal delay={0.1}>
            <ShimmerCard className="h-full">
              <div className="relative bg-card border border-accent-yellow/25 rounded-xl overflow-hidden h-full hover:border-accent-yellow/40 transition-all duration-300 shadow-[0_0_40px_rgba(191,140,93,0.06)]">
                <div className="absolute inset-0 bg-gradient-to-br from-accent-yellow/[0.04] via-transparent to-transparent pointer-events-none" />
                <div className="relative">
                  {/* Header */}
                  <div className="px-6 pt-6 pb-4">
                    <div className="flex items-center gap-3 mb-3">
                      <IconBadge color="yellow">
                        <CpuChipIcon
                          className="w-5 h-5 text-accent-yellow"
                          aria-hidden="true"
                        />
                      </IconBadge>
                      <div>
                        <h3 className="text-sm font-bold">Embedded Linux</h3>
                        <p className="text-2xs text-text-muted">
                          Yocto &amp; Buildroot
                        </p>
                      </div>
                    </div>
                    <p className="text-xs text-text-secondary leading-relaxed">
                      Include the ShellHub agent in your embedded Linux image.
                      Ship devices pre-configured for remote access from the
                      factory floor.
                    </p>
                  </div>

                  {/* Build config mockup */}
                  <div className="mx-4 mb-4">
                    <div className="bg-surface/60 border border-border rounded-lg overflow-hidden">
                      <div className="flex items-center gap-2 px-4 py-2 border-b border-border">
                        <span className="text-2xs text-text-muted font-mono">
                          local.conf
                        </span>
                        <span className="ml-auto px-1.5 py-0.5 text-2xs font-mono bg-accent-yellow/10 text-accent-yellow rounded">
                          Yocto
                        </span>
                      </div>
                      <div className="p-4 font-mono text-2xs leading-[1.7]">
                        <Comment># Enable ShellHub agent in the image</Comment>
                        <Ln>
                          <Kw>IMAGE_INSTALL</Kw>:append ={" "}
                          <Str>" shellhub-agent"</Str>
                        </Ln>
                        <Ln color={C.textMuted}>&nbsp;</Ln>
                        <Comment># Configure server address</Comment>
                        <Ln>
                          <Kw>SHELLHUB_SERVER</Kw> ={" "}
                          <Str>"https://cloud.shellhub.io"</Str>
                        </Ln>
                        <Ln>
                          <Kw>SHELLHUB_TENANT_ID</Kw> ={" "}
                          <Str>"your-tenant-id"</Str>
                        </Ln>
                        <Ln color={C.textMuted}>&nbsp;</Ln>
                        <Comment># Auto-register on first boot</Comment>
                        <Ln>
                          <Kw>SHELLHUB_AUTO_REGISTER</Kw> = <Str>"true"</Str>
                        </Ln>
                      </div>
                    </div>
                  </div>

                  {/* Buildroot config */}
                  <div className="mx-4 mb-5">
                    <div className="bg-surface/60 border border-border rounded-lg overflow-hidden">
                      <div className="flex items-center gap-2 px-4 py-2 border-b border-border">
                        <span className="text-2xs text-text-muted font-mono">
                          .config
                        </span>
                        <span className="ml-auto px-1.5 py-0.5 text-2xs font-mono bg-accent-yellow/10 text-accent-yellow rounded">
                          Buildroot
                        </span>
                      </div>
                      <div className="p-4 font-mono text-2xs leading-[1.7]">
                        <Ln>
                          <Kw>BR2_PACKAGE_SHELLHUB_AGENT</Kw>=<Str>y</Str>
                        </Ln>
                        <Ln>
                          <Kw>BR2_PACKAGE_SHELLHUB_AGENT_SERVER</Kw>=
                          <Str>"https://cloud.shellhub.io"</Str>
                        </Ln>
                        <Ln>
                          <Kw>BR2_PACKAGE_SHELLHUB_AGENT_TENANT_ID</Kw>=
                          <Str>"your-tenant-id"</Str>
                        </Ln>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </Section>

      {/* ─────────── Docker & Containers ─────────── */}
      <Section>
        <div className="grid lg:grid-cols-2 gap-12 items-center">
          {/* Left: diagram */}
          <Reveal>
            <ShimmerCard>
              <Card className="p-8 overflow-hidden">
                <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-6">
                  Connection Flow
                </p>

                {/* Diagram */}
                <div className="flex flex-col items-center gap-0">
                  {/* Developer */}
                  <div className="flex items-center gap-3 px-5 py-3 bg-surface border border-border rounded-lg w-full max-w-xs">
                    <div className="w-8 h-8 rounded-full bg-primary/15 flex items-center justify-center">
                      <UserIcon
                        className="w-4 h-4 text-primary"
                        aria-hidden="true"
                      />
                    </div>
                    <div>
                      <p className="text-xs font-medium">Developer</p>
                      <p className="text-2xs text-text-muted font-mono">
                        ssh user@container
                      </p>
                    </div>
                  </div>

                  {/* Arrow */}
                  <div className="flex flex-col items-center py-2">
                    <div className="w-px h-6 bg-primary/40" />
                    <ArrowDownIcon
                      className="w-3 h-3 text-primary/60"
                      aria-hidden="true"
                    />
                  </div>

                  {/* ShellHub Cloud */}
                  <div className="flex items-center gap-3 px-5 py-3 bg-primary/[0.06] border border-primary/25 rounded-lg w-full max-w-xs">
                    <div className="w-8 h-8 rounded-full bg-primary/15 flex items-center justify-center">
                      <CloudIcon
                        className="w-4 h-4 text-primary"
                        aria-hidden="true"
                      />
                    </div>
                    <div>
                      <p className="text-xs font-medium text-primary">
                        ShellHub
                      </p>
                      <p className="text-2xs text-text-muted font-mono">
                        SSH tunnel &amp; auth
                      </p>
                    </div>
                  </div>

                  {/* Arrow */}
                  <div className="flex flex-col items-center py-2">
                    <div className="w-px h-6 bg-accent-cyan/40" />
                    <ArrowDownIcon
                      className="w-3 h-3 text-accent-cyan/60"
                      aria-hidden="true"
                    />
                  </div>

                  {/* Docker Host */}
                  <div className="w-full max-w-xs border border-border rounded-lg overflow-hidden">
                    <div className="px-4 py-2 bg-surface/60 border-b border-border flex items-center gap-2">
                      <CubeIcon
                        className="w-4 h-4 text-accent-cyan"
                        aria-hidden="true"
                      />
                      <span className="text-2xs font-mono text-text-muted">
                        Docker Host
                      </span>
                    </div>
                    <div className="p-3 bg-card space-y-2">
                      {[
                        { name: "web-app", port: "8080", color: C.green },
                        { name: "api-server", port: "3000", color: C.cyan },
                        { name: "worker", port: "--", color: C.yellow },
                      ].map((c) => (
                        <div
                          key={c.name}
                          className="flex items-center gap-2 px-3 py-1.5 bg-surface rounded border border-border"
                        >
                          <div
                            className="w-2 h-2 rounded-full"
                            style={{ background: c.color }}
                          />
                          <span className="text-2xs font-mono text-text-secondary flex-1">
                            {c.name}
                          </span>
                          <span className="text-2xs font-mono text-text-muted">
                            :{c.port}
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              </Card>
            </ShimmerCard>
          </Reveal>

          {/* Right: copy */}
          <div>
            <SectionHeader
              align="left"
              size="sub"
              className="mb-8"
              eyebrow="Docker & Containers"
              title="SSH into containers, not just hosts"
              subtitle="SSH directly into Docker containers without exposing ports or running docker exec. ShellHub routes connections to individual containers through the agent on the host."
            />

            <Reveal delay={0.05}>
              <ShimmerCard>
                <WindowChrome
                  variant="terminal"
                  bodyClassName="overflow-x-auto"
                >
                  <Comment># Connect to a specific container</Comment>
                  <Ln>
                    <Val>$</Val> ssh{" "}
                    <Cyn>root@web-app.host01.ns.shellhub.io</Cyn>
                  </Ln>
                  <Ln color={C.textSec}>
                    Connected to web-app (container: a1b2c3d4)
                  </Ln>
                  <Ln color={C.textMuted}>&nbsp;</Ln>
                  <Ln>
                    <Val>root@web-app:~#</Val>{" "}
                    <Dim>curl localhost:8080/health</Dim>
                  </Ln>
                  <Ln>
                    <Str>{'{"status":"healthy","uptime":"14d 6h"}'}</Str>
                  </Ln>
                </WindowChrome>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </Section>

      {/* ─────────── API-First ─────────── */}
      <Section>
        <SectionHeader
          eyebrow="API-First"
          title="Automate everything with the REST API"
          subtitle="Every feature in ShellHub is accessible through a well-documented REST API. Build custom dashboards, automate workflows, or integrate with your internal tools."
        />

        <Reveal delay={0.1}>
          <ShimmerCard className="max-w-4xl mx-auto">
            <Card className="overflow-hidden">
              {/* API tabs */}
              <div className="flex items-center gap-1 px-5 py-3 border-b border-border bg-surface/60">
                <Badge shape="pill" color="green">
                  GET
                </Badge>
                <span className="text-2xs font-mono text-text-secondary ml-2">
                  /api/devices
                </span>
                <div className="ml-auto flex items-center gap-2">
                  <span className="px-2 py-0.5 text-2xs font-mono text-accent-green bg-accent-green/10 rounded">
                    200 OK
                  </span>
                </div>
              </div>

              <div className="grid md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border">
                {/* Request */}
                <div className="p-5">
                  <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-3">
                    Request
                  </p>
                  <div className="bg-surface/60 border border-border rounded-lg p-4 font-mono text-2xs leading-[1.7]">
                    <Ln>
                      <Val>$</Val> curl -s \
                    </Ln>
                    <Ln>
                      {" "}
                      -H <Str>"Authorization: Bearer $TOKEN"</Str> \
                    </Ln>
                    <Ln>
                      {" "}
                      -H <Str>"Content-Type: application/json"</Str> \
                    </Ln>
                    <Ln>
                      {" "}
                      <Cyn>https://cloud.shellhub.io/api/devices</Cyn>
                    </Ln>
                  </div>
                  <div className="mt-4 space-y-2">
                    <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-2">
                      Headers
                    </p>
                    {[
                      { key: "Authorization", val: "Bearer eyJhb..." },
                      { key: "Content-Type", val: "application/json" },
                      { key: "X-Tenant-ID", val: "d0a1c2e4..." },
                    ].map((h) => (
                      <div
                        key={h.key}
                        className="flex items-center gap-2 text-2xs font-mono"
                      >
                        <span className="text-primary">{h.key}:</span>
                        <span className="text-text-muted truncate">
                          {h.val}
                        </span>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Response */}
                <div className="p-5">
                  <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-3">
                    Response
                  </p>
                  <div className="bg-surface/60 border border-border rounded-lg p-4 font-mono text-2xs leading-[1.7]">
                    <Ln>[</Ln>
                    <Ln>
                      {"  "}
                      {"{"}
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"uid"</Kw>: <Str>"a1b2c3d4e5f6"</Str>,
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"name"</Kw>: <Str>"gateway-east-01"</Str>,
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"identity"</Kw>: {"{"}
                    </Ln>
                    <Ln>
                      {"      "}
                      <Kw>"mac"</Kw>: <Str>"00:1a:2b:3c:4d:5e"</Str>
                    </Ln>
                    <Ln>
                      {"    "}
                      {"}"},
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"info"</Kw>: {"{"}
                    </Ln>
                    <Ln>
                      {"      "}
                      <Kw>"version"</Kw>: <Str>"0.15.1"</Str>,
                    </Ln>
                    <Ln>
                      {"      "}
                      <Kw>"arch"</Kw>: <Str>"amd64"</Str>
                    </Ln>
                    <Ln>
                      {"    "}
                      {"}"},
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"status"</Kw>: <Str>"online"</Str>,
                    </Ln>
                    <Ln>
                      {"    "}
                      <Kw>"last_seen"</Kw>: <Str>"2026-02-14T..."</Str>
                    </Ln>
                    <Ln>
                      {"  "}
                      {"}"}
                    </Ln>
                    <Ln>]</Ln>
                  </div>
                </div>
              </div>

              {/* More endpoints */}
              <div className="px-5 py-3 border-t border-border bg-surface/40 flex flex-wrap gap-2">
                {[
                  { method: "POST", path: "/api/sessions", color: C.yellow },
                  { method: "GET", path: "/api/stats", color: C.green },
                  {
                    method: "PUT",
                    path: "/api/devices/{uid}",
                    color: C.blue,
                  },
                  {
                    method: "DELETE",
                    path: "/api/sessions/{uid}",
                    color: C.red,
                  },
                  { method: "GET", path: "/api/namespaces", color: C.green },
                ].map((ep) => (
                  <span
                    key={ep.path}
                    className="inline-flex items-center gap-1.5 px-2 py-1 bg-white/[0.03] border border-border rounded text-2xs font-mono text-text-muted hover:border-border-light transition-colors"
                  >
                    <span style={{ color: ep.color }} className="font-semibold">
                      {ep.method}
                    </span>
                    {ep.path}
                  </span>
                ))}
              </div>
            </Card>
          </ShimmerCard>
        </Reveal>
      </Section>

      {/* ─────────── Standard SSH callout ─────────── */}
      <Section>
        <Reveal className="text-center mb-14">
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
            Standard SSH. Zero plugins.
          </h2>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
            ShellHub works with any tool that supports SSH. No proprietary
            clients, no custom APIs, no vendor lock-in.
          </p>
        </Reveal>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            {
              color: C.primary,
              title: "SSH as transport",
              desc: "Any tool that uses SSH for remote execution works with ShellHub out of the box. Ansible, rsync, scp, Git over SSH.",
              icon: (
                <ArrowsRightLeftIcon
                  className="w-5 h-5"
                  style={{ color: C.primary }}
                  aria-hidden="true"
                />
              ),
            },
            {
              color: C.cyan,
              title: "No agent changes",
              desc: "The ShellHub agent runs independently on your devices. Install once and use it with every tool in your stack.",
              icon: (
                <ShieldCheckIcon
                  className="w-5 h-5"
                  style={{ color: C.cyan }}
                  aria-hidden="true"
                />
              ),
            },
            {
              color: C.green,
              title: "API-first design",
              desc: "Programmatic access to devices, sessions, and configurations. Build custom integrations with the REST API.",
              icon: (
                <CodeBracketIcon
                  className="w-5 h-5"
                  style={{ color: C.green }}
                  aria-hidden="true"
                />
              ),
            },
          ].map((b, i) => (
            <Reveal key={i} delay={i * 0.06}>
              <ShimmerCard className="h-full">
                <Card hover className="p-6 h-full">
                  <div
                    className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                    style={{
                      background: `${b.color}15`,
                      borderColor: `${b.color}25`,
                    }}
                  >
                    {b.icon}
                  </div>
                  <h4 className="text-sm font-semibold mb-2">{b.title}</h4>
                  <p className="text-xs text-text-secondary leading-relaxed">
                    {b.desc}
                  </p>
                </Card>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </Section>

      {/* ─────────── CTA ─────────── */}
      <Section>
        <Reveal>
          <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
            <ConnectionGrid />
            <div className="absolute inset-0 bg-gradient-to-br from-accent-cyan/[0.06] via-transparent to-primary/[0.04] pointer-events-none" />

            <div className="relative z-10">
              <SectionHeader
                variant="cta"
                eyebrow="Ready to integrate?"
                title="Start integrating today"
                subtitle="Get ShellHub running and connect it to your existing workflow in minutes. Standard SSH means zero learning curve."
              />

              <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
                <Button
                  as={Link}
                  to="/getting-started"
                  variant="primary"
                  size="xl"
                  glow
                  iconRight={<ArrowRight />}
                >
                  Get Started Free
                </Button>
                <Button
                  as="a"
                  href={docsUrl}
                  target="_blank"
                  rel="noopener noreferrer"
                  variant="outline"
                  size="xl"
                >
                  Read the Docs
                </Button>
              </div>
            </div>
          </div>
        </Reveal>
      </Section>
    </SiteLayout>
  );
}
