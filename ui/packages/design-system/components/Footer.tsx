import {
  DiscordIcon,
  GithubIcon,
  ShellHubLogo,
  TwitterXIcon,
} from "../primitives/icons";

export function Footer() {
  return (
    <footer className="border-t border-border pt-14 pb-8">
      <div className="max-w-7xl mx-auto px-8">
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-[2fr_1fr_1fr_1fr_1fr] gap-10 mb-10">
          <div>
            <a href="#" className="inline-block mb-4">
              <ShellHubLogo className="h-8" />
            </a>
            <p className="text-xs text-text-secondary max-w-[220px] leading-relaxed">
              The open source SSH gateway for remote access to Linux devices.
            </p>
          </div>
          {[
            {
              title: "Product",
              links: [
                "Features",
                "Pricing",
                "Cloud",
                "Self-Hosted",
                "Enterprise",
              ],
            },
            {
              title: "Resources",
              links: [
                "Documentation",
                "Getting Started",
                "API Reference",
                "Blog",
                "Changelog",
              ],
            },
            {
              title: "Company",
              links: ["About", "Careers", "Contact", "Partners"],
            },
            {
              title: "Legal",
              links: ["Privacy Policy", "Terms of Service", "Security"],
            },
          ].map((col) => (
            <div key={col.title}>
              <h5 className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-secondary mb-3">
                {col.title}
              </h5>
              <ul className="space-y-1.5">
                {col.links.map((l) => (
                  <li key={l}>
                    <a
                      href="#"
                      className="text-xs text-text-secondary hover:text-text-primary transition-colors"
                    >
                      {l}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="flex flex-col sm:flex-row justify-between items-center pt-6 border-t border-border gap-4">
          <span className="text-2xs font-mono text-text-secondary">
            &copy; 2026 ShellHub. All rights reserved.
          </span>
          <div className="flex gap-3">
            {[GithubIcon, TwitterXIcon, DiscordIcon].map((Icon, i) => (
              <a
                key={i}
                href="#"
                className="text-text-secondary hover:text-text-primary transition-colors"
              >
                <Icon className="w-4 h-4" />
              </a>
            ))}
          </div>
        </div>
      </div>
    </footer>
  );
}
