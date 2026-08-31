import { GithubIcon, ShellHubLogo } from "../primitives/icons";

const GITHUB_URL = "https://github.com/shellhub-io/shellhub";
const WEBSITE_URL = "http://website.localhost";
const DOCS_URL = "http://docs.localhost";
const CURRENT_YEAR = new Date().getFullYear();

type FooterApp = "website" | "docs";

const APP_ORIGINS: Record<FooterApp, string> = {
  website: WEBSITE_URL,
  docs: DOCS_URL,
};

const COLUMNS = [
  {
    title: "Product",
    links: [
      { label: "Features", href: `${WEBSITE_URL}/features` },
      { label: "Pricing", href: `${WEBSITE_URL}/pricing` },
      { label: "Enterprise", href: `${WEBSITE_URL}/enterprise` },
    ],
  },
  {
    title: "Resources",
    links: [
      { label: "Documentation", href: DOCS_URL },
      { label: "Getting Started", href: `${WEBSITE_URL}/getting-started` },
    ],
  },
];

interface FooterProps {
  app?: FooterApp;
  linkComponent?: React.ElementType;
}

function isSameApp(href: string, origin: string) {
  return href.startsWith(origin);
}

/**
 * Site footer shared by the website and the docs. app names which of the two is rendering it,
 * so a link that stays inside that app can be routed by linkComponent while a link crossing
 * to the other origin stays a plain anchor and reloads.
 */
export function Footer({
  app = "website",
  linkComponent: Link = "a",
}: FooterProps) {
  const origin = APP_ORIGINS[app];
  const linkCls = "text-text-secondary hover:text-text-primary transition-colors";

  return (
    <footer className="border-t border-border pt-14 pb-8">
      <div className="max-w-7xl mx-auto px-8">
        <div className="flex flex-wrap gap-10 mb-10">
          <div className="w-full md:flex-[2]">
            <Link href={WEBSITE_URL} className="inline-block mb-4">
              <ShellHubLogo className="h-8" />
            </Link>
            <p className="text-xs text-text-secondary max-w-[220px] leading-relaxed">
              The open source SSH gateway for remote access to Linux devices.
            </p>
          </div>
          {COLUMNS.map((col) => (
            <div key={col.title} className="flex-1 min-w-[140px] text-xs">
              <h5 className="text-2xs font-mono font-semibold uppercase tracking-[0.15em] text-text-secondary mb-3">
                {col.title}
              </h5>
              <ul className="space-y-1.5">
                {col.links.map((l) => (
                  <li key={l.label}>
                    {isSameApp(l.href, origin) ? (
                      <Link href={l.href} className={linkCls}>
                        {l.label}
                      </Link>
                    ) : (
                      <a
                        href={l.href}
                        className={linkCls}
                        target="_blank"
                        rel="noopener noreferrer"
                      >
                        {l.label}
                      </a>
                    )}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        <div className="flex flex-col sm:flex-row justify-between items-center pt-6 border-t border-border gap-4">
          <span className="text-2xs font-mono text-text-secondary">
            &copy; {CURRENT_YEAR} ShellHub. All rights reserved.
          </span>
          <div className="flex gap-3">
            <a
              href={GITHUB_URL}
              className={linkCls}
              aria-label="GitHub"
              target="_blank"
              rel="noopener noreferrer"
            >
              <GithubIcon className="w-4 h-4" />
            </a>
          </div>
        </div>
      </div>
    </footer>
  );
}
