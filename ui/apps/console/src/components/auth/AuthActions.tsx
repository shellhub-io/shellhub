import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";

/**
 * A way out of a public screen: a route inside the app, a page outside it, or an action that
 * stays on the screen.
 */
export type AuthLink =
  | { label: string; to: string }
  | { label: string; href: string }
  | { label: string; onClick: () => void };

interface AuthActionsProps {
  primary?: ReactNode;
  links?: AuthLink[];
  align?: "start" | "center";
}

const linkClass =
  "text-xs text-text-muted hover:text-text-secondary transition-colors";

/**
 * The bottom of every screen outside the console: the one primary action, then the ways out of
 * the screen, in one row, in one style. Screens send the user elsewhere only from here, so moving
 * between them always happens in the same place.
 */
export default function AuthActions({
  primary,
  links = [],
  align = "start",
}: AuthActionsProps) {
  return (
    <div className="space-y-4">
      {primary}
      {links.length > 0 && (
        <div
          className={cn(
            "flex flex-wrap items-center gap-x-5 gap-y-2",
            align === "center" && "justify-center",
          )}
        >
          {links.map((link) => (
            <AuthLinkItem key={link.label} link={link} />
          ))}
        </div>
      )}
    </div>
  );
}

function AuthLinkItem({ link }: { link: AuthLink }) {
  if ("to" in link) {
    return (
      <Link to={link.to} className={linkClass}>
        {link.label}
      </Link>
    );
  }

  if ("href" in link) {
    return (
      <a
        href={link.href}
        target="_blank"
        rel="noopener noreferrer"
        className={linkClass}
      >
        {link.label}
      </a>
    );
  }

  return (
    <button type="button" onClick={link.onClick} className={linkClass}>
      {link.label}
    </button>
  );
}
