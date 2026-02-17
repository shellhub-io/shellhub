import { Reveal } from "./components";
import { C } from "./constants";

export function Architecture() {
  return (
    <section className="py-24 bg-surface border-t border-b border-border">
      <div className="max-w-7xl mx-auto px-8">
        <Reveal className="text-center mb-4">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">Architecture</p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">Okay, but how does it actually work?</h2>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">Install a lightweight agent on your devices. The agent connects outbound to ShellHub&apos;s gateway. SSH in from anywhere.</p>
        </Reveal>

        <Reveal className="mt-12 bg-card border border-border rounded-xl p-6 lg:p-8 overflow-x-auto">
          <svg viewBox="0 0 900 320" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto min-w-[700px]">
            <defs>
              <marker id="aa" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.primary}/></marker>
              <marker id="aad" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={`${C.primary}60`}/></marker>
            </defs>
            <text x="100" y="30" fontFamily="IBM Plex Sans" fontSize="12" fill={C.textMuted} textAnchor="middle" letterSpacing=".1em">THE INTERNET</text>
            <circle cx="100" cy="100" r="35" fill="none" stroke={C.border} strokeWidth="1.5"/>
            <ellipse cx="100" cy="100" rx="35" ry="15" fill="none" stroke={C.border} strokeWidth="1"/>
            <ellipse cx="100" cy="100" rx="15" ry="35" fill="none" stroke={C.border} strokeWidth="1"/>
            <line x1="65" y1="100" x2="135" y2="100" stroke={C.border} strokeWidth="1"/>
            <circle cx="100" cy="100" r="35" fill="none" stroke={`${C.primary}30`} strokeWidth="1.5"/>
            <rect x="60" y="170" width="80" height="65" rx="10" fill={C.card} stroke={C.border}/>
            <circle cx="100" cy="190" r="10" stroke={C.primary} strokeWidth="1.2" fill="none"/>
            <circle cx="100" cy="187" r="3.5" fill={C.primary}/>
            <path d="M100 192 L100 195 M94 194 L106 194" stroke={C.primary} strokeWidth="1.2" strokeLinecap="round"/>
            <text x="100" y="225" fontFamily="IBM Plex Sans" fontSize="10" fill={C.textSec} textAnchor="middle">You</text>
            <line x1="145" y1="200" x2="285" y2="160" stroke={C.primary} strokeWidth="1.5" markerEnd="url(#aa)"/>
            <rect x="180" y="165" width="80" height="18" rx="4" fill={C.bg} stroke="none"/>
            <text x="220" y="178" fontFamily="IBM Plex Mono" fontSize="8" fill={C.primary} textAnchor="middle">Encrypted Tunnel</text>
            <rect x="290" y="40" width="300" height="240" rx="16" fill={`${C.primary}08`} stroke={C.primary} strokeWidth="1.5"/>
            <text x="440" y="30" fontFamily="IBM Plex Sans" fontSize="12" fill={C.primary} textAnchor="middle" fontWeight="600" letterSpacing=".1em">SHELLHUB CLOUD</text>
            <rect x="400" y="80" width="80" height="44" rx="10" fill={C.primaryDim} stroke={C.primary} strokeWidth="1"/>
            <text x="440" y="107" fontFamily="IBM Plex Mono" fontSize="14" fill={C.primary} textAnchor="middle" fontWeight="700">SH</text>
            {[
              { x: 310, y: 145, emoji: "\uD83D\uDD10", color: C.primary, label: "Authentication" },
              { x: 310, y: 182, emoji: "\uD83D\uDD12", color: C.primary, label: "Encryption" },
              { x: 310, y: 219, emoji: "\u23FA", color: C.cyan, label: "Session Rec." },
              { x: 460, y: 145, emoji: "\uD83D\uDEE1", color: C.yellow, label: "Firewall Rules" },
              { x: 460, y: 182, emoji: "\uD83D\uDCCB", color: C.green, label: "Audit Logging" },
              { x: 460, y: 219, emoji: "\uD83D\uDC65", color: C.primary, label: "Team RBAC" },
            ].map((b, i) => (
              <g key={i}>
                <rect x={b.x} y={b.y} width="110" height="28" rx="6" fill={C.card} stroke={C.border}/>
                <text x={b.x + 12} y={b.y + 17} fontSize="10" fill={b.color}>{b.emoji}</text>
                <text x={b.x + 30} y={b.y + 18} fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec}>{b.label}</text>
              </g>
            ))}
            <text x="750" y="30" fontFamily="IBM Plex Sans" fontSize="12" fill={C.textMuted} textAnchor="middle" letterSpacing=".1em">YOUR DEVICES</text>
            <rect x="630" y="50" width="6" height="225" rx="3" fill={C.border}/>
            <text x="633" y="44" fontFamily="IBM Plex Mono" fontSize="8" fill={C.textMuted} textAnchor="middle">NAT</text>
            <line x1="595" y1="105" x2="628" y2="105" stroke={C.primary} strokeWidth="1.5" markerEnd="url(#aa)"/>
            <line x1="638" y1="105" x2="660" y2="75" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#aad)"/>
            <line x1="638" y1="140" x2="660" y2="150" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#aad)"/>
            <line x1="638" y1="175" x2="660" y2="225" stroke={`${C.primary}60`} strokeWidth="1.2" strokeDasharray="4 3" markerEnd="url(#aad)"/>
            <text x="618" y="130" fontFamily="IBM Plex Mono" fontSize="7" fill={`${C.primary}60`} textAnchor="middle" transform="rotate(-90,618,130)">NAT Traversal</text>
            {[
              { y: 48, icon: "Pi", iconBg: C.green, label: "Raspberry Pi", begin: "0s" },
              { y: 110, icon: null, iconBg: C.primary, label: "Linux Server", begin: ".5s" },
              { y: 172, icon: "\uD83D\uDC33", iconBg: C.blue, label: "Docker Host", begin: "1s" },
              { y: 234, icon: null, iconBg: C.yellow, label: "IoT Gateway", begin: "1.5s", isIoT: true },
            ].map((d, i) => (
              <g key={i}>
                <rect x="664" y={d.y} width="140" height="52" rx="8" fill={C.card} stroke={C.border}/>
                {d.isIoT ? (
                  <><circle cx="692" cy={d.y + 20} r="10" fill="none" stroke={`${d.iconBg}50`} strokeWidth=".8"/><circle cx="692" cy={d.y + 20} r="4" fill={`${d.iconBg}30`}/></>
                ) : d.icon === "\uD83D\uDC33" ? (
                  <><rect x="678" y={d.y + 10} width="28" height="20" rx="4" fill={`${d.iconBg}15`} stroke={d.iconBg} strokeWidth=".8"/><text x="692" y={d.y + 24} fontSize="8" fill={d.iconBg} textAnchor="middle" fontWeight="600">{d.icon}</text></>
                ) : d.icon === "Pi" ? (
                  <><rect x="678" y={d.y + 10} width="28" height="20" rx="4" fill={`${d.iconBg}15`} stroke={d.iconBg} strokeWidth=".8"/><text x="692" y={d.y + 24} fontSize="9" fill={d.iconBg} textAnchor="middle" fontWeight="600">Pi</text></>
                ) : (
                  <><rect x="678" y={d.y + 10} width="28" height="20" rx="3" fill={`${d.iconBg}15`} stroke={d.iconBg} strokeWidth=".8"/><rect x="683" y={d.y + 15} width="18" height="10" rx="2" fill={`${d.iconBg}30`}/></>
                )}
                <text x="740" y={d.y + 24} fontFamily="IBM Plex Sans" fontSize="10" fill={C.text} textAnchor="start">{d.label}</text>
                <text x="740" y={d.y + 38} fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="start">agent running</text>
                <circle cx="794" cy={d.y + 8} r="3" fill={C.green}><animate attributeName="opacity" values="1;.3;1" dur="2s" repeatCount="indefinite" begin={d.begin}/></circle>
              </g>
            ))}
            <rect x="558" y="88" width="32" height="14" rx="3" fill={C.bg} stroke={`${C.primary}30`}/>
            <text x="574" y="98" fontFamily="IBM Plex Mono" fontSize="7" fill={C.primary} textAnchor="middle">SSH</text>
          </svg>
        </Reveal>

        <Reveal className="flex gap-4 mt-8 justify-center flex-wrap">
          {[
            { n: "1", t: "Install the ShellHub agent on your device" },
            { n: "2", t: "Agent connects outbound to ShellHub Cloud" },
            { n: "3", t: "Access your devices securely from anywhere" },
          ].map((s) => (
            <div key={s.n} className="flex items-center gap-3 px-4 py-2.5 bg-card border border-border rounded-lg">
              <span className="w-6 h-6 rounded-full bg-primary/10 border border-primary/20 text-primary flex items-center justify-center text-2xs font-bold font-mono">{s.n}</span>
              <span className="text-sm text-text-secondary">{s.t}</span>
            </div>
          ))}
        </Reveal>
      </div>
    </section>
  );
}
