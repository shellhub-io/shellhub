import {
  CommandLineIcon,
  CubeIcon,
  DocumentArrowUpIcon,
  FingerPrintIcon,
  GlobeAltIcon,
  KeyIcon,
  PlayCircleIcon,
  ShieldCheckIcon,
  WindowIcon,
} from "@heroicons/react/24/outline";
import { Section, SectionHeader, InfoCard } from "@/components";
import { C } from "@shellhub/design-system/constants";

const FEATURES = [
  {
    icon: CommandLineIcon,
    color: C.primary,
    title: "Native SSH",
    desc: "Use the SSH client you already have. No plugin, no wrapper, no proprietary protocol.",
  },
  {
    icon: WindowIcon,
    color: C.cyan,
    title: "Web Terminal",
    desc: "Open a shell on any device from the browser, with nothing installed on your machine.",
  },
  {
    icon: ShieldCheckIcon,
    color: C.yellow,
    title: "Access Policies",
    desc: "Default deny. Grant members, roles or API keys a set of devices, down to the unix login.",
  },
  {
    icon: FingerPrintIcon,
    color: C.green,
    title: "SSH Identities",
    desc: "Every key belongs to a person or an API key. A key never seen before waits for a member to approve it.",
  },
  {
    icon: PlayCircleIcon,
    color: C.red,
    title: "Session Recording",
    desc: "Replay what happened in a session. Cloud and Enterprise keep the recordings on the server.",
  },
  {
    icon: DocumentArrowUpIcon,
    color: C.blue,
    title: "SCP and SFTP",
    desc: "Move files to and from remote devices with the tools you already script against.",
  },
  {
    icon: CubeIcon,
    color: C.primary,
    title: "Container Access",
    desc: "SSH straight into containers running on remote hosts. No docker exec on the host.",
  },
  {
    icon: GlobeAltIcon,
    color: C.cyan,
    title: "Web Endpoints",
    desc: "Reach an HTTP service on a device at a URL, without opening a port. Cloud and Enterprise.",
  },
  {
    icon: KeyIcon,
    color: C.yellow,
    title: "MFA and Single Sign-On",
    desc: "TOTP for console logins on Cloud and Enterprise, SAML single sign-on on Enterprise.",
  },
];

/**
 * Landing section: the feature grid, three across. A feature that is not in every edition says
 * which editions have it, so the page never promises Community something it lacks.
 */
export function FeatureGrid() {
  return (
    <Section bordered={false}>
      <SectionHeader
        eyebrow="Features"
        title="Everything you need to manage remote devices."
      />

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {FEATURES.map((f, i) => (
          <InfoCard
            key={f.title}
            icon={f.icon}
            color={f.color}
            title={f.title}
            description={f.desc}
            delay={(i % 3) * 0.06}
          />
        ))}
      </div>
    </Section>
  );
}
