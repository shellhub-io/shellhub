import { Reveal, ShimmerCard } from "./components";
import { C } from "./constants";

export function FeatureGrid() {
  return (
    <section className="py-24">
      <div className="max-w-7xl mx-auto px-8">
        <Reveal className="text-center mb-14">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Features</p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight">Everything you need to manage remote devices.</h2>
        </Reveal>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {[
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round"><polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/></svg>, color: C.primary, title: "Native SSH Support", desc: "Use your standard SSH client. No proprietary tools or plugins required.", delay: 0 },
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.cyan} strokeWidth="1.5" strokeLinecap="round"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>, color: C.cyan, title: "SCP/SFTP File Transfer", desc: "Transfer files to and from remote devices with SCP and SFTP.", delay: 0.06 },
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.yellow} strokeWidth="1.5" strokeLinecap="round"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/><circle cx="12" cy="16" r="1"/></svg>, color: C.yellow, title: "Multi-Factor Auth", desc: "Require TOTP-based MFA for SSH connections. Works with any authenticator app.", delay: 0.12 },
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.green} strokeWidth="1.5" strokeLinecap="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 013 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>, color: C.green, title: "Audit Logging", desc: "Full audit trail of every connection, command, and session. Export logs for compliance.", delay: 0 },
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87"/><path d="M16 3.13a4 4 0 010 7.75"/></svg>, color: C.primary, title: "Team Management & RBAC", desc: "Invite team members, assign roles, and control who can access which devices.", delay: 0.06 },
            { icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={C.blue} strokeWidth="1.5" strokeLinecap="round"><rect x="2" y="6" width="20" height="12" rx="2"/><path d="M12 12h.01"/><path d="M17 12h.01"/><path d="M7 12h.01"/></svg>, color: C.blue, title: "Docker Container Access", desc: "SSH directly into Docker containers running on remote hosts. No docker exec needed.", delay: 0.12 },
          ].map((f, i) => (
            <Reveal key={i} delay={f.delay}>
              <ShimmerCard className="bg-card border border-border rounded-xl p-6 hover:border-border-light transition-all duration-300 h-full">
                <div className="w-10 h-10 rounded-lg flex items-center justify-center mb-4 border" style={{ background: `${f.color}15`, borderColor: `${f.color}25` }}>
                  {f.icon}
                </div>
                <h4 className="text-sm font-semibold mb-2 group-hover:text-primary transition-colors">{f.title}</h4>
                <p className="text-xs text-text-secondary leading-relaxed">{f.desc}</p>
              </ShimmerCard>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}
