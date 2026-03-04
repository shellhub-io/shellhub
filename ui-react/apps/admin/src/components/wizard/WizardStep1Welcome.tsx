import {
  ArrowsRightLeftIcon,
  ShieldCheckIcon,
  BoltIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";

const features = [
  {
    icon: <ArrowsRightLeftIcon className="w-5 h-5" />,
    title: "Remote Access",
    description: "Reach any Linux device from anywhere via CLI or web — no VPN needed.",
  },
  {
    icon: <ShieldCheckIcon className="w-5 h-5" />,
    title: "Secure Connection",
    description: "Encrypted tunnels bypass firewalls and NAT automatically.",
  },
  {
    icon: <BoltIcon className="w-5 h-5" />,
    title: "Easy Setup",
    description: "One command on the device. Automated agent management from there.",
  },
];

export default function WizardStep1Welcome() {
  const name = useAuthStore((s) => s.name || s.username);

  return (
    <div className="py-2">
      <div className="mb-6">
        <h2 className="text-xl font-mono font-bold text-text-primary mb-1">
          Welcome{name ? `, ${name}` : ""}.
        </h2>
        <p className="text-sm text-text-muted leading-relaxed">
          ShellHub gives you a single encrypted gateway to every Linux device
          you manage. Let&apos;s connect your first one.
        </p>
      </div>

      <ul className="flex flex-col sm:flex-row gap-3">
        {features.map((f) => (
          <li
            key={f.title}
            className="flex-1 bg-background border border-border rounded-xl p-4 flex flex-col gap-3 hover:border-primary/30 transition-colors duration-200"
          >
            <div className="w-9 h-9 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
              {f.icon}
            </div>
            <div>
              <h3 className="text-sm font-semibold text-text-primary mb-1">
                {f.title}
              </h3>
              <p className="text-xs text-text-muted leading-relaxed">
                {f.description}
              </p>
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
