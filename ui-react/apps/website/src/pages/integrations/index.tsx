import { useEffect, useState, useCallback } from "react";
import { Navbar } from "../landing/Navbar";
import { Footer } from "@shellhub/design-system/components";
import { Reveal, ShimmerCard, ConnectionGrid } from "../landing/components";
import { C } from "../landing/constants";

/* ------------------------------------------------------------------ */
/*  Terminal chrome (macOS-style dots + optional title)                */
/* ------------------------------------------------------------------ */
function TerminalChrome({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <ShimmerCard className="bg-card border border-border rounded-xl overflow-hidden">
      <div className="flex items-center gap-2 px-5 py-3 border-b border-border bg-surface/60">
        <div className="w-3 h-3 rounded-full bg-accent-red/60" />
        <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
        <div className="w-3 h-3 rounded-full bg-accent-green/60" />
        <span className="ml-2 text-2xs text-text-muted font-mono">{title}</span>
      </div>
      <div className="p-5 font-mono text-2xs leading-[1.7] overflow-x-auto">{children}</div>
    </ShimmerCard>
  );
}

/* ------------------------------------------------------------------ */
/*  Syntax-highlighted line helpers                                    */
/* ------------------------------------------------------------------ */
function Ln({ color, children }: { color?: string; children: React.ReactNode }) {
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

      {/* ─────────── Hero ─────────── */}
      <section className="relative pt-32 pb-24 overflow-hidden">
        <ConnectionGrid />
        <div className="absolute inset-0 bg-gradient-radial from-primary/8 via-transparent to-transparent pointer-events-none" />
        <div className="absolute top-1/3 left-1/4 w-96 h-96 bg-accent-cyan/6 rounded-full blur-3xl pointer-events-none" />

        <div className="max-w-7xl mx-auto px-8 relative z-10 text-center">
          <Reveal>
            <span className="inline-block px-3 py-1 text-2xs font-mono font-semibold uppercase tracking-[0.15em] bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/20 rounded-full mb-6">
              Integrations
            </span>
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
              ShellHub speaks standard SSH. Any tool that connects over SSH works out of the box &mdash; no plugins, no custom agents, no vendor lock-in.
            </p>
          </Reveal>
          <Reveal>
            <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
              <a
                href="/v2/docs/"
                className="relative inline-flex items-center gap-2.5 px-8 py-3.5 text-[15px] font-semibold bg-primary text-[#111214] rounded-xl shadow-[0_0_20px_rgba(102,122,204,0.3)] hover:shadow-[0_0_32px_rgba(102,122,204,0.5)] hover:scale-[1.03] active:scale-[0.98] transition-all duration-300 group overflow-hidden"
              >
                <span className="absolute inset-0 bg-gradient-to-r from-white/0 via-white/25 to-white/0 translate-x-[-200%] group-hover:translate-x-[200%] transition-transform duration-700 ease-in-out" />
                <span className="relative">Browse Docs</span>
                <svg className="relative w-4 h-4 group-hover:translate-x-1 transition-transform duration-300" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
                </svg>
              </a>
              <a
                href="/v2/getting-started"
                className="inline-flex items-center gap-2 px-6 py-3.5 text-[15px] font-medium text-text-secondary border border-border rounded-xl hover:text-text-primary hover:border-border-light hover:bg-white/[0.04] transition-all duration-300"
              >
                Get Started Free
              </a>
            </div>
          </Reveal>
        </div>
      </section>

      {/* ─────────── Automation & IaC ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Left copy */}
            <div>
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Automation &amp; Infrastructure as Code
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Ansible &amp; Terraform, zero VPN
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  Use ShellHub as the SSH transport for your Ansible playbooks and Terraform provisioners. Manage devices behind NAT, firewalls, or cellular networks without VPN tunnels or bastion hosts.
                </p>
              </Reveal>

              <div className="space-y-3">
                {[
                  { label: "Drop-in SSH replacement", desc: "Set ProxyCommand once and every tool uses ShellHub transparently" },
                  { label: "NAT traversal built in", desc: "Reach devices behind CGNAT, firewalls, and private networks" },
                  { label: "Ansible connection plugin", desc: "Native Ansible connection plugin for seamless integration" },
                  { label: "Terraform provisioner", desc: "Run remote-exec provisioners through ShellHub SSH" },
                ].map((cap, i) => (
                  <Reveal key={i} delay={i * 0.04}>
                    <div className="flex items-start gap-3">
                      <svg className="w-4 h-4 text-accent-green shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                        <path strokeLinecap="round" strokeLinejoin="round" d="m4.5 12.75 6 6 9-13.5" />
                      </svg>
                      <div>
                        <p className="text-sm font-medium text-text-primary">{cap.label}</p>
                        <p className="text-xs text-text-secondary leading-relaxed">{cap.desc}</p>
                      </div>
                    </div>
                  </Reveal>
                ))}
              </div>
            </div>

            {/* Right terminal mockup */}
            <Reveal delay={0.1}>
              <TerminalChrome title="ansible-playbook">
                <Comment># Deploy firmware update via ShellHub SSH</Comment>
                <Ln>
                  <Val>$</Val> ansible-playbook -i inventory.yml deploy.yml
                </Ln>
                <div className="mt-3" />
                <Ln color={C.textSec}>PLAY [Update edge devices] ****************************</Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>TASK [Gathering Facts] ********************************</Ln>
                <Ln><Str>ok</Str>: [gateway-east-01.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Str>ok</Str>: [sensor-rack-07.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Str>ok</Str>: [plc-floor-03.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>TASK [Copy firmware binary] ****************************</Ln>
                <Ln><Val>changed</Val>: [gateway-east-01.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Val>changed</Val>: [sensor-rack-07.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Val>changed</Val>: [plc-floor-03.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <div className="mt-2" />
                <Ln color={C.textSec}>TASK [Apply update &amp; restart] **************************</Ln>
                <Ln><Val>changed</Val>: [gateway-east-01.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Val>changed</Val>: [sensor-rack-07.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <Ln><Val>changed</Val>: [plc-floor-03.d0a1c2.<Cyn>shellhub.io</Cyn>]</Ln>
                <div className="mt-3" />
                <Ln color={C.textSec}>PLAY RECAP ********************************************</Ln>
                <Ln>
                  gateway-east-01 : <Str>ok=3</Str> <Val>changed=2</Val> unreachable=0 failed=<Str>0</Str>
                </Ln>
                <Ln>
                  sensor-rack-07 &nbsp;: <Str>ok=3</Str> <Val>changed=2</Val> unreachable=0 failed=<Str>0</Str>
                </Ln>
                <Ln>
                  plc-floor-03 &nbsp;&nbsp;&nbsp;: <Str>ok=3</Str> <Val>changed=2</Val> unreachable=0 failed=<Str>0</Str>
                </Ln>
              </TerminalChrome>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ─────────── CI/CD Pipelines ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              CI/CD Pipelines
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Deploy from any CI system
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              GitHub Actions, GitLab CI, Jenkins, CircleCI &mdash; if it can run SSH, it can deploy through ShellHub. No custom plugins required.
            </p>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="max-w-3xl mx-auto">
              <div className="relative bg-card border border-accent-green/25 rounded-xl overflow-hidden shadow-[0_0_40px_rgba(130,165,104,0.08)]">
                <div className="absolute inset-0 bg-gradient-to-br from-accent-green/[0.04] via-transparent to-transparent pointer-events-none" />
                <div className="relative">
                  {/* Tab bar */}
                  <div className="flex items-center gap-2 px-5 py-3 border-b border-border bg-surface/60">
                    <div className="w-3 h-3 rounded-full bg-accent-red/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-yellow/60" />
                    <div className="w-3 h-3 rounded-full bg-accent-green/60" />
                    <span className="ml-2 text-2xs text-text-muted font-mono">.github/workflows/deploy.yml</span>
                    <div className="ml-auto flex items-center gap-1.5">
                      <span className="px-2 py-0.5 text-2xs font-mono font-semibold uppercase tracking-[0.1em] bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                        Passing
                      </span>
                    </div>
                  </div>

                  {/* YAML content */}
                  <div className="p-5 font-mono text-2xs leading-[1.7] overflow-x-auto">
                    <Ln><Kw>name</Kw>: <Str>Deploy to Edge Devices</Str></Ln>
                    <Ln><Kw>on</Kw>:</Ln>
                    <Ln>  <Kw>push</Kw>:</Ln>
                    <Ln>    <Kw>branches</Kw>: [<Str>main</Str>]</Ln>
                    <div className="mt-2" />
                    <Ln><Kw>jobs</Kw>:</Ln>
                    <Ln>  <Kw>deploy</Kw>:</Ln>
                    <Ln>    <Kw>runs-on</Kw>: <Str>ubuntu-latest</Str></Ln>
                    <Ln>    <Kw>steps</Kw>:</Ln>
                    <Ln>      - <Kw>name</Kw>: <Str>Configure ShellHub SSH</Str></Ln>
                    <Ln>        <Kw>run</Kw>: |</Ln>
                    <Ln>          <Dim>mkdir -p ~/.ssh</Dim></Ln>
                    <Ln>          <Dim>echo <Str>"$&#123;&#123; secrets.SSH_KEY &#125;&#125;"</Str> &gt; ~/.ssh/id_rsa</Dim></Ln>
                    <Ln>          <Dim>chmod 600 ~/.ssh/id_rsa</Dim></Ln>
                    <div className="mt-2" />
                    <Ln>      - <Kw>name</Kw>: <Str>Deploy application</Str></Ln>
                    <Ln>        <Kw>run</Kw>: |</Ln>
                    <Ln>          <Dim>ssh <Cyn>admin@device.namespace.shellhub.io</Cyn> \</Dim></Ln>
                    <Ln>          <Dim>  <Str>"cd /opt/app &amp;&amp; git pull &amp;&amp; systemctl restart app"</Str></Dim></Ln>
                    <div className="mt-2" />
                    <Ln>      - <Kw>name</Kw>: <Str>Verify deployment</Str></Ln>
                    <Ln>        <Kw>run</Kw>: |</Ln>
                    <Ln>          <Dim>ssh <Cyn>admin@device.namespace.shellhub.io</Cyn> \</Dim></Ln>
                    <Ln>          <Dim>  <Str>"curl -sf http://localhost:8080/health"</Str></Dim></Ln>
                  </div>
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ─────────── Development Tools ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              Development Tools
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Develop directly on remote devices
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Connect your favorite editor or build system to devices through ShellHub. Full IDE support with zero configuration changes.
            </p>
          </Reveal>

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
                      <span className="ml-2 text-2xs text-text-muted font-mono">VS Code &mdash; Remote SSH</span>
                      <div className="ml-auto">
                        <span className="px-2 py-0.5 text-2xs font-mono bg-accent-blue/10 text-accent-blue border border-accent-blue/20 rounded-full">
                          SSH: gateway-east-01
                        </span>
                      </div>
                    </div>

                    <div className="flex min-h-[240px]">
                      {/* Sidebar */}
                      <div className="w-44 shrink-0 border-r border-border bg-surface/40 p-3">
                        <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-2">Explorer</p>
                        <div className="space-y-1">
                          {([
                            { name: "src/", indent: 0, isDir: true, active: false },
                            { name: "main.py", indent: 1, isDir: false, active: true },
                            { name: "config.yml", indent: 1, isDir: false, active: false },
                            { name: "utils/", indent: 1, isDir: true, active: false },
                            { name: "tests/", indent: 0, isDir: true, active: false },
                            { name: "Dockerfile", indent: 0, isDir: false, active: false },
                            { name: "requirements.txt", indent: 0, isDir: false, active: false },
                          ] as const).map((f, i) => (
                            <div
                              key={i}
                              className={`flex items-center gap-1.5 px-1.5 py-0.5 rounded text-2xs font-mono ${f.active ? "bg-accent-blue/10 text-accent-blue" : "text-text-secondary hover:bg-white/[0.03]"}`}
                              style={{ paddingLeft: `${f.indent * 12 + 6}px` }}
                            >
                              {f.isDir ? (
                                <svg className="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" />
                                </svg>
                              ) : (
                                <svg className="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                                  <path strokeLinecap="round" strokeLinejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m2.25 0H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
                                </svg>
                              )}
                              {f.name}
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Editor area */}
                      <div className="flex-1 p-4 font-mono text-2xs leading-[1.8]">
                        <div className="flex items-center gap-3 mb-3 border-b border-border pb-2">
                          <span className="px-2 py-0.5 text-2xs bg-accent-blue/10 text-accent-blue border-b-2 border-accent-blue rounded-t">main.py</span>
                          <span className="px-2 py-0.5 text-2xs text-text-muted">config.yml</span>
                        </div>
                        <div className="flex gap-3">
                          <div className="text-text-muted select-none text-right w-5 shrink-0">
                            {[1, 2, 3, 4, 5, 6, 7, 8].map(n => (
                              <div key={n}>{n}</div>
                            ))}
                          </div>
                          <div>
                            <Ln><Kw>import</Kw> <Str>flask</Str></Ln>
                            <Ln><Kw>from</Kw> config <Kw>import</Kw> settings</Ln>
                            <Ln color={C.textMuted}>&nbsp;</Ln>
                            <Ln>app = flask.Flask(__name__)</Ln>
                            <Ln color={C.textMuted}>&nbsp;</Ln>
                            <Ln><Kw>@app</Kw>.route(<Str>"/health"</Str>)</Ln>
                            <Ln><Kw>def</Kw> <Val>health</Val>():</Ln>
                            <Ln>    <Kw>return</Kw> {`{`}<Str>"status"</Str>: <Str>"ok"</Str>{`}`}</Ln>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Status bar */}
                    <div className="flex items-center justify-between px-4 py-1.5 bg-accent-blue/10 border-t border-border text-2xs font-mono">
                      <div className="flex items-center gap-3">
                        <span className="text-accent-blue flex items-center gap-1">
                          <svg className="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 0 1 1.242 7.244l-4.5 4.5a4.5 4.5 0 0 1-6.364-6.364l1.757-1.757m13.35-.622 1.757-1.757a4.5 4.5 0 0 0-6.364-6.364l-4.5 4.5a4.5 4.5 0 0 0 1.242 7.244" />
                          </svg>
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
                        <div className="w-10 h-10 rounded-lg bg-accent-yellow/10 border border-accent-yellow/20 flex items-center justify-center">
                          <svg className="w-5 h-5 text-accent-yellow" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 0 0 2.25-2.25V6.75a2.25 2.25 0 0 0-2.25-2.25H6.75A2.25 2.25 0 0 0 4.5 6.75v10.5a2.25 2.25 0 0 0 2.25 2.25Z" />
                          </svg>
                        </div>
                        <div>
                          <h3 className="text-sm font-bold">Embedded Linux</h3>
                          <p className="text-2xs text-text-muted">Yocto &amp; Buildroot</p>
                        </div>
                      </div>
                      <p className="text-xs text-text-secondary leading-relaxed">
                        Include the ShellHub agent in your embedded Linux image. Ship devices pre-configured for remote access from the factory floor.
                      </p>
                    </div>

                    {/* Build config mockup */}
                    <div className="mx-4 mb-4">
                      <div className="bg-surface/60 border border-border rounded-lg overflow-hidden">
                        <div className="flex items-center gap-2 px-4 py-2 border-b border-border">
                          <span className="text-2xs text-text-muted font-mono">local.conf</span>
                          <span className="ml-auto px-1.5 py-0.5 text-2xs font-mono bg-accent-yellow/10 text-accent-yellow rounded">Yocto</span>
                        </div>
                        <div className="p-4 font-mono text-2xs leading-[1.7]">
                          <Comment># Enable ShellHub agent in the image</Comment>
                          <Ln><Kw>IMAGE_INSTALL</Kw>:append = <Str>" shellhub-agent"</Str></Ln>
                          <Ln color={C.textMuted}>&nbsp;</Ln>
                          <Comment># Configure server address</Comment>
                          <Ln><Kw>SHELLHUB_SERVER</Kw> = <Str>"https://cloud.shellhub.io"</Str></Ln>
                          <Ln><Kw>SHELLHUB_TENANT_ID</Kw> = <Str>"your-tenant-id"</Str></Ln>
                          <Ln color={C.textMuted}>&nbsp;</Ln>
                          <Comment># Auto-register on first boot</Comment>
                          <Ln><Kw>SHELLHUB_AUTO_REGISTER</Kw> = <Str>"true"</Str></Ln>
                        </div>
                      </div>
                    </div>

                    {/* Buildroot config */}
                    <div className="mx-4 mb-5">
                      <div className="bg-surface/60 border border-border rounded-lg overflow-hidden">
                        <div className="flex items-center gap-2 px-4 py-2 border-b border-border">
                          <span className="text-2xs text-text-muted font-mono">.config</span>
                          <span className="ml-auto px-1.5 py-0.5 text-2xs font-mono bg-accent-yellow/10 text-accent-yellow rounded">Buildroot</span>
                        </div>
                        <div className="p-4 font-mono text-2xs leading-[1.7]">
                          <Ln><Kw>BR2_PACKAGE_SHELLHUB_AGENT</Kw>=<Str>y</Str></Ln>
                          <Ln><Kw>BR2_PACKAGE_SHELLHUB_AGENT_SERVER</Kw>=<Str>"https://cloud.shellhub.io"</Str></Ln>
                          <Ln><Kw>BR2_PACKAGE_SHELLHUB_AGENT_TENANT_ID</Kw>=<Str>"your-tenant-id"</Str></Ln>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>
          </div>
        </div>
      </section>

      {/* ─────────── Docker & Containers ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <div className="grid lg:grid-cols-2 gap-12 items-center">
            {/* Left: diagram */}
            <Reveal>
              <ShimmerCard>
                <div className="bg-card border border-border rounded-xl p-8 overflow-hidden">
                  <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-6">Connection Flow</p>

                  {/* Diagram */}
                  <div className="flex flex-col items-center gap-0">
                    {/* Developer */}
                    <div className="flex items-center gap-3 px-5 py-3 bg-surface border border-border rounded-lg w-full max-w-xs">
                      <div className="w-8 h-8 rounded-full bg-primary/15 flex items-center justify-center">
                        <svg className="w-4 h-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
                        </svg>
                      </div>
                      <div>
                        <p className="text-xs font-medium">Developer</p>
                        <p className="text-2xs text-text-muted font-mono">ssh user@container</p>
                      </div>
                    </div>

                    {/* Arrow */}
                    <div className="flex flex-col items-center py-2">
                      <div className="w-px h-6 bg-primary/40" />
                      <svg className="w-3 h-3 text-primary/60" viewBox="0 0 12 12"><path d="M6 0v10M2 6l4 5 4-5" fill="none" stroke="currentColor" strokeWidth={1.5} /></svg>
                    </div>

                    {/* ShellHub Cloud */}
                    <div className="flex items-center gap-3 px-5 py-3 bg-primary/[0.06] border border-primary/25 rounded-lg w-full max-w-xs">
                      <div className="w-8 h-8 rounded-full bg-primary/15 flex items-center justify-center">
                        <svg className="w-4 h-4 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 15a4.5 4.5 0 0 0 4.5 4.5H18a3.75 3.75 0 0 0 1.332-7.257 3 3 0 0 0-3.758-3.848 5.25 5.25 0 0 0-10.233 2.33A4.502 4.502 0 0 0 2.25 15Z" />
                        </svg>
                      </div>
                      <div>
                        <p className="text-xs font-medium text-primary">ShellHub</p>
                        <p className="text-2xs text-text-muted font-mono">SSH tunnel &amp; auth</p>
                      </div>
                    </div>

                    {/* Arrow */}
                    <div className="flex flex-col items-center py-2">
                      <div className="w-px h-6 bg-accent-cyan/40" />
                      <svg className="w-3 h-3 text-accent-cyan/60" viewBox="0 0 12 12"><path d="M6 0v10M2 6l4 5 4-5" fill="none" stroke="currentColor" strokeWidth={1.5} /></svg>
                    </div>

                    {/* Docker Host */}
                    <div className="w-full max-w-xs border border-border rounded-lg overflow-hidden">
                      <div className="px-4 py-2 bg-surface/60 border-b border-border flex items-center gap-2">
                        <svg className="w-4 h-4 text-accent-cyan" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M5.25 14.25h13.5m-13.5 0a3 3 0 0 1-3-3m3 3a3 3 0 1 0 0 6h13.5a3 3 0 1 0 0-6m-16.5-3a3 3 0 0 1 3-3h13.5a3 3 0 0 1 3 3m-19.5 0a4.5 4.5 0 0 1 .9-2.7L5.737 5.1a3.375 3.375 0 0 1 2.7-1.35h7.126c1.062 0 2.062.5 2.7 1.35l2.587 3.45a4.5 4.5 0 0 1 .9 2.7m0 0a3 3 0 0 1-3 3m0 3h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Zm-3 6h.008v.008h-.008v-.008Zm0-6h.008v.008h-.008v-.008Z" />
                        </svg>
                        <span className="text-2xs font-mono text-text-muted">Docker Host</span>
                      </div>
                      <div className="p-3 bg-card space-y-2">
                        {[
                          { name: "web-app", port: "8080", color: C.green },
                          { name: "api-server", port: "3000", color: C.cyan },
                          { name: "worker", port: "--", color: C.yellow },
                        ].map((c) => (
                          <div key={c.name} className="flex items-center gap-2 px-3 py-1.5 bg-surface rounded border border-border">
                            <div className="w-2 h-2 rounded-full" style={{ background: c.color }} />
                            <span className="text-2xs font-mono text-text-secondary flex-1">{c.name}</span>
                            <span className="text-2xs font-mono text-text-muted">:{c.port}</span>
                          </div>
                        ))}
                      </div>
                    </div>
                  </div>
                </div>
              </ShimmerCard>
            </Reveal>

            {/* Right: copy */}
            <div>
              <Reveal>
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Docker &amp; Containers
                </p>
                <h2 className="text-[clamp(1.75rem,4vw,2.5rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  SSH into containers, not just hosts
                </h2>
                <p className="text-sm text-text-secondary leading-relaxed mb-8">
                  SSH directly into Docker containers without exposing ports or running docker exec. ShellHub routes connections to individual containers through the agent on the host.
                </p>
              </Reveal>

              <Reveal delay={0.05}>
                <TerminalChrome title="terminal">
                  <Comment># Connect to a specific container</Comment>
                  <Ln><Val>$</Val> ssh <Cyn>root@web-app.host01.ns.shellhub.io</Cyn></Ln>
                  <Ln color={C.textSec}>Connected to web-app (container: a1b2c3d4)</Ln>
                  <Ln color={C.textMuted}>&nbsp;</Ln>
                  <Ln><Val>root@web-app:~#</Val> <Dim>curl localhost:8080/health</Dim></Ln>
                  <Ln><Str>{`{"status":"healthy","uptime":"14d 6h"}`}</Str></Ln>
                </TerminalChrome>
              </Reveal>
            </div>
          </div>
        </div>
      </section>

      {/* ─────────── API-First ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
              API-First
            </p>
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Automate everything with the REST API
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              Every feature in ShellHub is accessible through a well-documented REST API. Build custom dashboards, automate workflows, or integrate with your internal tools.
            </p>
          </Reveal>

          <Reveal delay={0.1}>
            <ShimmerCard className="max-w-4xl mx-auto">
              <div className="bg-card border border-border rounded-xl overflow-hidden">
                {/* API tabs */}
                <div className="flex items-center gap-1 px-5 py-3 border-b border-border bg-surface/60">
                  <span className="px-3 py-1 text-2xs font-mono font-semibold bg-accent-green/10 text-accent-green border border-accent-green/20 rounded-full">
                    GET
                  </span>
                  <span className="text-2xs font-mono text-text-secondary ml-2">/api/devices</span>
                  <div className="ml-auto flex items-center gap-2">
                    <span className="px-2 py-0.5 text-2xs font-mono text-accent-green bg-accent-green/10 rounded">200 OK</span>
                  </div>
                </div>

                <div className="grid md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-border">
                  {/* Request */}
                  <div className="p-5">
                    <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-3">Request</p>
                    <div className="bg-surface/60 border border-border rounded-lg p-4 font-mono text-2xs leading-[1.7]">
                      <Ln><Val>$</Val> curl -s \</Ln>
                      <Ln>  -H <Str>"Authorization: Bearer $TOKEN"</Str> \</Ln>
                      <Ln>  -H <Str>"Content-Type: application/json"</Str> \</Ln>
                      <Ln>  <Cyn>https://cloud.shellhub.io/api/devices</Cyn></Ln>
                    </div>
                    <div className="mt-4 space-y-2">
                      <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-2">Headers</p>
                      {[
                        { key: "Authorization", val: "Bearer eyJhb..." },
                        { key: "Content-Type", val: "application/json" },
                        { key: "X-Tenant-ID", val: "d0a1c2e4..." },
                      ].map((h) => (
                        <div key={h.key} className="flex items-center gap-2 text-2xs font-mono">
                          <span className="text-primary">{h.key}:</span>
                          <span className="text-text-muted truncate">{h.val}</span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Response */}
                  <div className="p-5">
                    <p className="text-2xs font-mono text-text-muted uppercase tracking-wider mb-3">Response</p>
                    <div className="bg-surface/60 border border-border rounded-lg p-4 font-mono text-2xs leading-[1.7]">
                      <Ln>{`[`}</Ln>
                      <Ln>{"  "}{`{`}</Ln>
                      <Ln>{"    "}<Kw>"uid"</Kw>: <Str>"a1b2c3d4e5f6"</Str>,</Ln>
                      <Ln>{"    "}<Kw>"name"</Kw>: <Str>"gateway-east-01"</Str>,</Ln>
                      <Ln>{"    "}<Kw>"identity"</Kw>: {`{`}</Ln>
                      <Ln>{"      "}<Kw>"mac"</Kw>: <Str>"00:1a:2b:3c:4d:5e"</Str></Ln>
                      <Ln>{"    "}{`}`},</Ln>
                      <Ln>{"    "}<Kw>"info"</Kw>: {`{`}</Ln>
                      <Ln>{"      "}<Kw>"version"</Kw>: <Str>"0.15.1"</Str>,</Ln>
                      <Ln>{"      "}<Kw>"arch"</Kw>: <Str>"amd64"</Str></Ln>
                      <Ln>{"    "}{`}`},</Ln>
                      <Ln>{"    "}<Kw>"status"</Kw>: <Str>"online"</Str>,</Ln>
                      <Ln>{"    "}<Kw>"last_seen"</Kw>: <Str>"2026-02-14T..."</Str></Ln>
                      <Ln>{"  "}{`}`}</Ln>
                      <Ln>{`]`}</Ln>
                    </div>
                  </div>
                </div>

                {/* More endpoints */}
                <div className="px-5 py-3 border-t border-border bg-surface/40 flex flex-wrap gap-2">
                  {[
                    { method: "POST", path: "/api/sessions", color: C.yellow },
                    { method: "GET", path: "/api/stats", color: C.green },
                    { method: "PUT", path: "/api/devices/{uid}", color: C.blue },
                    { method: "DELETE", path: "/api/sessions/{uid}", color: C.red },
                    { method: "GET", path: "/api/namespaces", color: C.green },
                  ].map((ep) => (
                    <span key={ep.path} className="inline-flex items-center gap-1.5 px-2 py-1 bg-white/[0.03] border border-border rounded text-2xs font-mono text-text-muted hover:border-border-light transition-colors">
                      <span style={{ color: ep.color }} className="font-semibold">{ep.method}</span>
                      {ep.path}
                    </span>
                  ))}
                </div>
              </div>
            </ShimmerCard>
          </Reveal>
        </div>
      </section>

      {/* ─────────── Standard SSH callout ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal className="text-center mb-14">
            <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
              Standard SSH. Zero plugins.
            </h2>
            <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">
              ShellHub works with any tool that supports SSH. No proprietary clients, no custom APIs, no vendor lock-in.
            </p>
          </Reveal>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {[
              {
                color: C.primary,
                title: "SSH as transport",
                desc: "Any tool that uses SSH for remote execution works with ShellHub out of the box. Ansible, rsync, scp, Git over SSH.",
                icon: (
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke={C.primary} strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" />
                  </svg>
                ),
              },
              {
                color: C.cyan,
                title: "No agent changes",
                desc: "The ShellHub agent runs independently on your devices. Install once and use it with every tool in your stack.",
                icon: (
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke={C.cyan} strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" />
                  </svg>
                ),
              },
              {
                color: C.green,
                title: "API-first design",
                desc: "Programmatic access to devices, sessions, and configurations. Build custom integrations with the REST API.",
                icon: (
                  <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke={C.green} strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M17.25 6.75 22.5 12l-5.25 5.25m-10.5 0L1.5 12l5.25-5.25m7.5-3-4.5 16.5" />
                  </svg>
                ),
              },
            ].map((b, i) => (
              <Reveal key={i} delay={i * 0.06}>
                <ShimmerCard className="h-full">
                  <div className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                    <div
                      className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border"
                      style={{ background: `${b.color}15`, borderColor: `${b.color}25` }}
                    >
                      {b.icon}
                    </div>
                    <h4 className="text-sm font-semibold mb-2">{b.title}</h4>
                    <p className="text-xs text-text-secondary leading-relaxed">{b.desc}</p>
                  </div>
                </ShimmerCard>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ─────────── CTA ─────────── */}
      <section className="py-24 border-t border-border">
        <div className="max-w-7xl mx-auto px-8">
          <Reveal>
            <div className="relative bg-card border border-border rounded-2xl p-12 text-center overflow-hidden">
              <ConnectionGrid />
              <div className="absolute inset-0 bg-gradient-to-br from-accent-cyan/[0.06] via-transparent to-primary/[0.04] pointer-events-none" />

              <div className="relative z-10">
                <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">
                  Ready to integrate?
                </p>
                <h2 className="text-[clamp(1.5rem,3vw,2.25rem)] font-bold tracking-[-0.03em] leading-tight mb-4">
                  Start integrating today
                </h2>
                <p className="text-sm text-text-secondary max-w-md mx-auto leading-relaxed mb-8">
                  Get ShellHub running and connect it to your existing workflow in minutes. Standard SSH means zero learning curve.
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
                    href="/v2/docs/"
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
