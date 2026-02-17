import { Reveal, ShimmerCard } from "./components";
import { C } from "./constants";

export function HowItWorks() {
  return (
    <section className="py-24 border-t border-border">
      <div className="max-w-7xl mx-auto px-8">
        <Reveal className="text-center mb-14">
          <p className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-[#7B8EDB] mb-3">How It Works</p>
          <h2 className="text-[clamp(1.75rem,4vw,3rem)] font-bold tracking-[-0.03em] leading-tight mb-4">From anywhere to any device.</h2>
          <p className="text-sm text-text-secondary max-w-lg mx-auto leading-relaxed">Instead of juggling VPNs, public IPs, and firewall rules, connect to all your devices through one secure gateway.</p>
        </Reveal>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {/* Card 1: Remote SSH Access */}
          <Reveal>
            <ShimmerCard className="bg-card border border-border rounded-xl p-6 lg:p-8 hover:border-primary/30 transition-all duration-300">
              <div className="flex items-center gap-3 mb-5">
                <span className="text-2xs font-mono font-semibold text-[#7B8EDB] border border-primary/20 px-2.5 py-1 rounded-md uppercase tracking-[0.1em]">Remote Access</span>
                <span className="text-2xs font-mono font-bold text-text-secondary">01</span>
              </div>
              <h3 className="text-lg font-bold tracking-[-0.02em] mb-5">SSH into any device, from anywhere</h3>
              <svg viewBox="0 0 520 200" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto">
                <defs><marker id="arrow-pri" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.primary}/></marker></defs>
                <rect x="10" y="65" width="80" height="70" rx="10" fill={C.card} stroke={C.border}/>
                <circle cx="50" cy="85" r="12" stroke={C.primary} strokeWidth="1.5" fill="none"/>
                <path d="M50 80 L50 83 M46 82 L54 82" stroke={C.primary} strokeWidth="1.5" strokeLinecap="round"/>
                <circle cx="50" cy="78" r="3" fill={C.primary}/>
                <text x="50" y="125" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">User</text>
                <line x1="95" y1="100" x2="148" y2="100" stroke={C.primary} strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#arrow-pri)"/>
                <text x="122" y="92" fontFamily="IBM Plex Mono" fontSize="9" fill={C.primaryGlow} textAnchor="middle" letterSpacing=".05em">SSH</text>
                <rect x="152" y="65" width="90" height="70" rx="10" fill={C.card} stroke={C.border}/>
                <rect x="165" y="78" width="64" height="40" rx="4" fill={C.surface} stroke={C.border}/>
                <text x="197" y="95" fontFamily="IBM Plex Mono" fontSize="8" fill={C.primary} textAnchor="middle">$ ssh</text>
                <text x="197" y="107" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textMuted} textAnchor="middle">user@device</text>
                <text x="197" y="125" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">Terminal</text>
                <line x1="247" y1="100" x2="290" y2="100" stroke={C.primary} strokeWidth="1.5" markerEnd="url(#arrow-pri)"/>
                <rect x="294" y="50" width="100" height="100" rx="12" fill={C.primaryDim} stroke={C.primary} strokeWidth="1.5"/>
                <rect x="322" y="72" width="44" height="28" rx="6" fill={`${C.primary}30`}/>
                <text x="344" y="89" fontFamily="IBM Plex Mono" fontSize="9" fill={C.primary} textAnchor="middle" fontWeight="600">SH</text>
                <text x="344" y="115" fontFamily="IBM Plex Sans" fontSize="11" fill={C.primary} textAnchor="middle" fontWeight="600">ShellHub</text>
                <text x="344" y="130" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Cloud</text>
                <line x1="399" y1="100" x2="428" y2="100" stroke={C.primary} strokeWidth="1.5" markerEnd="url(#arrow-pri)"/>
                <rect x="432" y="60" width="8" height="80" rx="2" fill={C.border}/><rect x="432" y="60" width="8" height="80" rx="2" fill={`${C.primary}15`}/>
                <text x="436" y="55" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textMuted} textAnchor="middle" letterSpacing=".05em">NAT</text>
                <line x1="444" y1="100" x2="460" y2="100" stroke={C.primary} strokeWidth="1.5" markerEnd="url(#arrow-pri)"/>
                <rect x="464" y="65" width="50" height="70" rx="10" fill={C.card} stroke={C.border}/>
                <rect x="474" y="78" width="30" height="22" rx="3" fill={C.surface} stroke={C.primaryGlow}/>
                <rect x="480" y="104" width="18" height="4" rx="2" fill={C.border}/><circle cx="489" cy="111" r="1.5" fill={C.primary}/>
                <text x="489" y="125" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">Device</text>
              </svg>
            </ShimmerCard>
          </Reveal>

          {/* Card 2: Session Recording */}
          <Reveal delay={0.08}>
            <ShimmerCard className="bg-card border border-border rounded-xl p-6 lg:p-8 hover:border-accent-cyan/30 transition-all duration-300">
              <div className="flex items-center gap-3 mb-5">
                <span className="text-2xs font-mono font-semibold text-accent-cyan border border-accent-cyan/20 px-2.5 py-1 rounded-md uppercase tracking-[0.1em]">Audit</span>
                <span className="text-2xs font-mono font-bold text-text-secondary">02</span>
              </div>
              <h3 className="text-lg font-bold tracking-[-0.02em] mb-5">Record and replay every session</h3>
              <svg viewBox="0 0 520 200" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto">
                <defs><marker id="arrow-cyan" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.cyan}/></marker></defs>
                <rect x="10" y="65" width="100" height="70" rx="10" fill={C.card} stroke={C.border}/>
                <rect x="22" y="78" width="76" height="36" rx="4" fill={C.surface} stroke={C.border}/>
                <text x="60" y="94" fontFamily="IBM Plex Mono" fontSize="7" fill={C.cyan} textAnchor="middle">session_01</text>
                <text x="60" y="106" fontFamily="IBM Plex Mono" fontSize="6" fill={C.textMuted} textAnchor="middle">recording...</text>
                <circle cx="90" cy="82" r="4" fill={C.cyan} opacity=".6"><animate attributeName="opacity" values=".3;1;.3" dur="1.5s" repeatCount="indefinite"/></circle>
                <text x="60" y="125" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">SSH Session</text>
                <line x1="115" y1="100" x2="165" y2="100" stroke={C.cyan} strokeWidth="1.5" markerEnd="url(#arrow-cyan)"/>
                <rect x="170" y="55" width="100" height="90" rx="12" fill={C.cyanDim} stroke={C.cyan} strokeWidth="1.5"/>
                <rect x="196" y="72" width="48" height="28" rx="6" fill={`${C.cyan}30`}/>
                <text x="220" y="89" fontFamily="IBM Plex Mono" fontSize="9" fill={C.cyan} textAnchor="middle" fontWeight="600">SH</text>
                <text x="220" y="115" fontFamily="IBM Plex Sans" fontSize="11" fill={C.cyan} textAnchor="middle" fontWeight="600">ShellHub</text>
                <text x="220" y="130" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Gateway</text>
                <line x1="275" y1="85" x2="340" y2="70" stroke={C.cyan} strokeWidth="1.5" markerEnd="url(#arrow-cyan)"/>
                <line x1="275" y1="115" x2="340" y2="135" stroke={C.cyan} strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#arrow-cyan)"/>
                <rect x="345" y="38" width="100" height="64" rx="10" fill={C.card} stroke={C.border}/>
                <ellipse cx="395" cy="55" rx="20" ry="8" fill="none" stroke={C.cyan} strokeWidth="1.2"/>
                <path d="M375 55 L375 72 Q375 80 395 80 Q415 80 415 72 L415 55" fill="none" stroke={C.cyan} strokeWidth="1.2"/>
                <text x="395" y="93" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">Storage</text>
                <line x1="450" y1="70" x2="470" y2="70" stroke={C.cyan} strokeWidth="1.2" strokeDasharray="3 2" markerEnd="url(#arrow-cyan)"/>
                <rect x="474" y="50" width="40" height="40" rx="8" fill={`${C.cyan}20`} stroke={C.cyan} strokeWidth="1"/>
                <polygon points="488,63 488,77 500,70" fill={C.cyan}/>
                <text x="494" y="103" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Replay</text>
                <rect x="345" y="115" width="100" height="50" rx="10" fill={C.card} stroke={C.border}/>
                <rect x="358" y="126" width="74" height="26" rx="4" fill={C.surface}/>
                <text x="395" y="137" fontFamily="IBM Plex Mono" fontSize="7" fill={C.cyan} textAnchor="middle">$ ls -la</text>
                <text x="395" y="147" fontFamily="IBM Plex Mono" fontSize="6" fill={C.textMuted} textAnchor="middle">drwxr-xr-x</text>
                <text x="395" y="175" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">Live View</text>
              </svg>
            </ShimmerCard>
          </Reveal>

          {/* Card 3: Firewall Rules */}
          <Reveal delay={0.16}>
            <ShimmerCard className="bg-card border border-border rounded-xl p-6 lg:p-8 hover:border-accent-yellow/30 transition-all duration-300">
              <div className="flex items-center gap-3 mb-5">
                <span className="text-2xs font-mono font-semibold text-accent-yellow border border-accent-yellow/20 px-2.5 py-1 rounded-md uppercase tracking-[0.1em]">Security</span>
                <span className="text-2xs font-mono font-bold text-text-secondary">03</span>
              </div>
              <h3 className="text-lg font-bold tracking-[-0.02em] mb-5">Control access with firewall rules</h3>
              <svg viewBox="0 0 520 200" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto">
                <defs>
                  <marker id="arrow-yel" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.yellow}/></marker>
                  <marker id="arrow-grn" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.green}/></marker>
                  <marker id="arrow-red" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.red}/></marker>
                </defs>
                <rect x="10" y="55" width="100" height="90" rx="10" fill={C.card} stroke={C.border}/>
                <text x="60" y="80" fontFamily="IBM Plex Sans" fontSize="11" fill={C.text} textAnchor="middle" fontWeight="600">Connection</text>
                <text x="60" y="95" fontFamily="IBM Plex Sans" fontSize="11" fill={C.text} textAnchor="middle" fontWeight="600">Requests</text>
                <rect x="22" y="108" width="76" height="10" rx="3" fill={C.yellowDim}/><circle cx="30" cy="113" r="3" fill={C.green}/><text x="56" y="116" fontFamily="IBM Plex Mono" fontSize="6" fill={C.textMuted}>192.168.1.10</text>
                <rect x="22" y="122" width="76" height="10" rx="3" fill={C.yellowDim}/><circle cx="30" cy="127" r="3" fill={C.red}/><text x="56" y="130" fontFamily="IBM Plex Mono" fontSize="6" fill={C.textMuted}>10.0.0.55</text>
                <line x1="115" y1="100" x2="178" y2="100" stroke={C.yellow} strokeWidth="1.5" markerEnd="url(#arrow-yel)"/>
                <rect x="182" y="40" width="140" height="120" rx="12" fill={C.yellowDim} stroke={C.yellow} strokeWidth="1.5"/>
                <text x="252" y="62" fontFamily="IBM Plex Mono" fontSize="9" fill={C.yellow} textAnchor="middle" letterSpacing=".05em">RULES ENGINE</text>
                <rect x="196" y="72" width="112" height="22" rx="5" fill={C.card} stroke={C.border}/>
                <circle cx="210" cy="83" r="5" fill={`${C.green}30`} stroke={C.green}/><text x="210" y="86" fontSize="8" fill={C.green} textAnchor="middle">&#10003;</text>
                <text x="252" y="86" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textSec} textAnchor="middle">Allow 192.168.*</text>
                <rect x="196" y="100" width="112" height="22" rx="5" fill={C.card} stroke={C.border}/>
                <circle cx="210" cy="111" r="5" fill={C.redDim} stroke={C.red}/><text x="210" y="114" fontSize="8" fill={C.red} textAnchor="middle">&#10005;</text>
                <text x="252" y="114" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textSec} textAnchor="middle">Deny 10.0.0.*</text>
                <rect x="196" y="128" width="112" height="22" rx="5" fill={C.card} stroke={C.border}/>
                <circle cx="210" cy="139" r="5" fill={`${C.green}30`} stroke={C.green}/><text x="210" y="142" fontSize="8" fill={C.green} textAnchor="middle">&#10003;</text>
                <text x="252" y="142" fontFamily="IBM Plex Mono" fontSize="7" fill={C.textSec} textAnchor="middle">Allow port 22</text>
                <line x1="327" y1="80" x2="380" y2="60" stroke={C.green} strokeWidth="1.5" markerEnd="url(#arrow-grn)"/>
                <line x1="327" y1="120" x2="380" y2="150" stroke={C.red} strokeWidth="1.5" strokeDasharray="4 3" markerEnd="url(#arrow-red)"/>
                <rect x="384" y="35" width="80" height="55" rx="10" fill={C.greenDim} stroke={C.green} strokeWidth="1"/>
                <text x="424" y="57" fontFamily="IBM Plex Sans" fontSize="11" fill={C.green} textAnchor="middle" fontWeight="600">Allowed</text>
                <rect x="400" y="67" width="48" height="14" rx="4" fill={C.card} stroke={C.border}/>
                <circle cx="408" cy="74" r="2" fill={C.green}/><text x="430" y="77" fontSize="6" fill={C.textMuted} textAnchor="middle">&rarr; Device</text>
                <rect x="384" y="125" width="80" height="55" rx="10" fill={C.redDim} stroke={C.red} strokeWidth="1"/>
                <text x="424" y="148" fontFamily="IBM Plex Sans" fontSize="11" fill={C.red} textAnchor="middle" fontWeight="600">Denied</text>
                <line x1="410" y1="160" x2="438" y2="170" stroke={C.red} strokeWidth="1.5"/>
                <line x1="438" y1="160" x2="410" y2="170" stroke={C.red} strokeWidth="1.5"/>
              </svg>
            </ShimmerCard>
          </Reveal>

          {/* Card 4: Web Terminal */}
          <Reveal delay={0.24}>
            <ShimmerCard className="bg-card border border-border rounded-xl p-6 lg:p-8 hover:border-accent-green/30 transition-all duration-300">
              <div className="flex items-center gap-3 mb-5">
                <span className="text-2xs font-mono font-semibold text-accent-green border border-accent-green/20 px-2.5 py-1 rounded-md uppercase tracking-[0.1em]">Web Terminal</span>
                <span className="text-2xs font-mono font-bold text-text-secondary">04</span>
              </div>
              <h3 className="text-lg font-bold tracking-[-0.02em] mb-5">Access devices from your browser</h3>
              <svg viewBox="0 0 520 200" fill="none" xmlns="http://www.w3.org/2000/svg" className="w-full h-auto">
                <defs><marker id="arrow-green" markerWidth="8" markerHeight="6" refX="8" refY="3" orient="auto"><path d="M0,0 L8,3 L0,6" fill={C.green}/></marker></defs>
                <rect x="10" y="50" width="110" height="100" rx="10" fill={C.card} stroke={C.border}/>
                <rect x="10" y="50" width="110" height="24" rx="10" fill={C.surface}/><rect x="10" y="64" width="110" height="10" fill={C.surface}/>
                <circle cx="25" cy="62" r="4" fill={C.red} opacity=".7"/><circle cx="37" cy="62" r="4" fill={C.yellow} opacity=".7"/><circle cx="49" cy="62" r="4" fill={C.green} opacity=".7"/>
                <rect x="60" y="58" width="50" height="8" rx="4" fill={C.bg}/>
                <rect x="20" y="82" width="90" height="56" rx="3" fill={C.bg}/>
                <text x="65" y="98" fontFamily="IBM Plex Mono" fontSize="7" fill={C.green} textAnchor="middle">shellhub.io</text>
                <text x="65" y="112" fontFamily="IBM Plex Mono" fontSize="6" fill={C.textMuted} textAnchor="middle">Web Terminal</text>
                <rect x="30" y="118" width="60" height="3" rx="1" fill={`${C.green}20`}/><rect x="30" y="124" width="40" height="3" rx="1" fill={`${C.green}10`}/>
                <text x="65" y="157" fontFamily="IBM Plex Sans" fontSize="11" fill={C.textSec} textAnchor="middle">Browser</text>
                <line x1="125" y1="100" x2="172" y2="100" stroke={C.green} strokeWidth="1.5" markerEnd="url(#arrow-green)"/>
                <text x="148" y="92" fontFamily="IBM Plex Mono" fontSize="9" fill={`${C.green}60`} textAnchor="middle" letterSpacing=".05em">HTTPS</text>
                <rect x="176" y="50" width="110" height="100" rx="12" fill={C.greenDim} stroke={C.green} strokeWidth="1.5"/>
                <rect x="204" y="68" width="54" height="28" rx="6" fill={`${C.green}30`}/>
                <text x="231" y="85" fontFamily="IBM Plex Mono" fontSize="9" fill={C.green} textAnchor="middle" fontWeight="600">SH</text>
                <text x="231" y="115" fontFamily="IBM Plex Sans" fontSize="11" fill={C.green} textAnchor="middle" fontWeight="600">ShellHub</text>
                <text x="231" y="130" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Gateway</text>
                <line x1="291" y1="75" x2="340" y2="45" stroke={C.green} strokeWidth="1.2" markerEnd="url(#arrow-green)"/>
                <line x1="291" y1="100" x2="340" y2="100" stroke={C.green} strokeWidth="1.2" markerEnd="url(#arrow-green)"/>
                <line x1="291" y1="125" x2="340" y2="155" stroke={C.green} strokeWidth="1.2" markerEnd="url(#arrow-green)"/>
                <text x="318" y="68" fontFamily="IBM Plex Mono" fontSize="7" fill={`${C.green}40`} textAnchor="middle">SSH</text>
                <text x="318" y="94" fontFamily="IBM Plex Mono" fontSize="7" fill={`${C.green}40`} textAnchor="middle">SSH</text>
                <text x="318" y="142" fontFamily="IBM Plex Mono" fontSize="7" fill={`${C.green}40`} textAnchor="middle">SSH</text>
                <rect x="344" y="20" width="100" height="50" rx="8" fill={C.card} stroke={C.border}/>
                <rect x="356" y="30" width="20" height="14" rx="3" fill={`${C.green}20`} stroke={C.green} strokeWidth=".8"/><text x="366" y="40" fontSize="8" fill={C.green} textAnchor="middle">Pi</text>
                <text x="415" y="42" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Raspberry Pi</text>
                <text x="415" y="55" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textMuted} textAnchor="middle" letterSpacing=".05em">192.168.1.10</text>
                <rect x="344" y="78" width="100" height="50" rx="8" fill={C.card} stroke={C.border}/>
                <rect x="356" y="88" width="20" height="14" rx="2" fill={`${C.green}20`} stroke={C.green} strokeWidth=".8"/><rect x="360" y="92" width="12" height="6" rx="1" fill={`${C.green}40`}/>
                <text x="415" y="100" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">Linux Server</text>
                <text x="415" y="113" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textMuted} textAnchor="middle" letterSpacing=".05em">10.0.0.5</text>
                <rect x="344" y="136" width="100" height="50" rx="8" fill={C.card} stroke={C.border}/>
                <circle cx="366" cy="155" r="8" fill="none" stroke={C.green} strokeWidth=".8"/><circle cx="366" cy="155" r="3" fill={`${C.green}40`}/>
                <text x="415" y="158" fontFamily="IBM Plex Sans" fontSize="9" fill={C.textSec} textAnchor="middle">IoT Device</text>
                <text x="415" y="171" fontFamily="IBM Plex Mono" fontSize="9" fill={C.textMuted} textAnchor="middle" letterSpacing=".05em">172.16.0.8</text>
              </svg>
            </ShimmerCard>
          </Reveal>
        </div>
      </div>
    </section>
  );
}
