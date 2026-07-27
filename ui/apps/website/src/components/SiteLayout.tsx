import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { SiteHeader } from "@/components/SiteHeader";
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
      <Footer linkComponent={RouterLink} />
    </div>
  );
}
