import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { cn } from "@shellhub/design-system/cn";
import { ShellHubLogo } from "@shellhub/design-system/primitives";
import { Bars3Icon, XMarkIcon } from "@heroicons/react/24/outline";
import { ActionButton } from "@/components";
import { loginUrl, signupUrl } from "@/links";
import { navEntries } from "./navData";
import { Navbar, MobileDropdown } from "./Navbar";

const loginAction = { label: "Log In", href: loginUrl, external: true };
const signupAction = {
  label: "Sign Up Free",
  href: signupUrl,
  external: true,
};

function HeaderActions({ fullWidth = false }) {
  return (
    <>
      <ActionButton
        action={loginAction}
        variant="outline"
        size="md"
        fullWidth={fullWidth}
      />
      <ActionButton
        action={signupAction}
        size="md"
        glow={false}
        fullWidth={fullWidth}
        iconRight={null}
      />
    </>
  );
}

export function SiteHeader() {
  const [scrolled, setScrolled] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 50);
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const closeMobile = () => setMobileOpen(false);

  return (
    <header
      className={cn(
        "fixed inset-x-0 top-0 z-appbar h-14 border-b transition-all duration-300",
        scrolled || mobileOpen
          ? "border-border bg-background shadow-[0_1px_30px_rgba(0,0,0,.3)] backdrop-blur-xl backdrop-saturate-[1.8]"
          : "border-transparent bg-background/70",
      )}
    >
      <div className="mx-auto flex h-full max-w-7xl items-center justify-between px-8">
        <Link to="/" className="shrink-0">
          <ShellHubLogo className="h-8" />
        </Link>

        <button
          type="button"
          data-testid="mobile-nav-toggle"
          className="p-2 text-text-secondary lg:hidden"
          onClick={() => setMobileOpen((v) => !v)}
          aria-expanded={mobileOpen}
          aria-label={mobileOpen ? "Close menu" : "Open menu"}
        >
          {mobileOpen ? (
            <XMarkIcon className="size-5" aria-hidden="true" />
          ) : (
            <Bars3Icon className="size-5" aria-hidden="true" />
          )}
        </button>

        <Navbar />

        <div className="hidden items-center gap-2 lg:flex">
          <HeaderActions />
        </div>
      </div>

      {mobileOpen && (
        <nav
          data-testid="mobile-nav"
          className="flex flex-col gap-0.5 border-b border-border bg-background p-3 shadow-xl backdrop-blur-xl lg:hidden"
        >
          {navEntries.map((entry) =>
            entry.kind === "link" ? (
              <Link
                key={entry.label}
                to={entry.href}
                onClick={closeMobile}
                className="rounded-lg px-3 py-2 text-[13px] font-medium text-text-secondary"
              >
                {entry.label}
              </Link>
            ) : (
              <MobileDropdown
                key={entry.label}
                label={entry.label}
                items={entry.items}
                onNavigate={closeMobile}
              />
            ),
          )}
          <div className="mt-1 flex flex-col gap-2 border-t border-border pt-2">
            <HeaderActions fullWidth />
          </div>
        </nav>
      )}
    </header>
  );
}
