import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { SiteHeader } from "@/components/SiteHeader";
import { docsUrl, websiteUrl } from "@/links";
import { Footer } from "@shellhub/design-system/components";

function RouterLink({
  href,
  className,
  children,
}: {
  href: string;
  className?: string;
  children: React.ReactNode;
}) {
  return (
    <Link to={href} className={className}>
      {children}
    </Link>
  );
}

/**
 * Page shell for every marketing route — full-height background, base type, and overflow-x
 * clipped so a decorative element that overhangs the viewport cannot scroll the page sideways.
 */
export function SiteLayout({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) {
  return (
    <div
      className={cn(
        "min-h-screen overflow-x-hidden bg-background font-sans text-text-primary antialiased",
        className,
      )}
    >
      <SiteHeader />
      {children}
      <Footer linkComponent={RouterLink} websiteUrl={websiteUrl} docsUrl={docsUrl} />
    </div>
  );
}
