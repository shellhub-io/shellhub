import { Outlet } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { ShellHubLogo } from "@shellhub/design-system/primitives";
import { isCommunity } from "@/env";
import AuthFooterLinks from "@/components/common/AuthFooterLinks";
import { columnWidths, type ColumnWidth } from "@/components/layout/columnWidths";

interface AuthLayoutProps {
  width?: ColumnWidth;
}

/**
 * The shell of the screens outside the console: the logo and a card holding the screen. The
 * community edition gets the docs, version and community line under it; the paid editions end
 * at the card. The column is the form's width unless a screen asks for the wide one.
 */
export default function AuthLayout({ width = "md" }: AuthLayoutProps) {
  return (
    <div className="relative min-h-screen bg-background grid-bg">
      <main
        className={cn(
          "relative w-full mx-auto px-4 sm:px-6 pt-16 pb-16 flex flex-col",
          columnWidths[width],
        )}
      >
        <ShellHubLogo className="h-10 mb-10" />
        <div className="bg-surface border border-border rounded-2xl p-8">
          <Outlet />
        </div>
        {isCommunity() && <AuthFooterLinks />}
      </main>
    </div>
  );
}
