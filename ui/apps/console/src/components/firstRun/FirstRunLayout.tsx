import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { XMarkIcon } from "@heroicons/react/24/outline";
import SessionMenu from "@/components/layout/SessionMenu";
import FramedShell from "@/components/layout/FramedShell";
import ScreenIntro from "@/components/layout/ScreenIntro";
import { useIsDesktop } from "@/hooks/useIsDesktop";

interface FirstRunLayoutProps {
  eyebrow: string;
  title?: string;
  lead?: string;
  signedIn: boolean;
  inConsole: boolean;
  children: ReactNode;
}

/**
 * The focused screen the first run happens on: the framed shell without the sidebar, since every
 * page it leads to is empty until a device exists. A signed-in user gets the account menu; once
 * they are inside a namespace (`inConsole`), the page also offers to skip into the console for this
 * visit. Nothing remembers the skip: the trail comes back until the namespace has a device.
 */
export default function FirstRunLayout({
  eyebrow,
  title = "Get your first shell",
  lead = "From nothing to a shell on one of your devices. This page follows along, and picks up where you left off if you leave.",
  signedIn,
  inConsole,
  children,
}: FirstRunLayoutProps) {
  const isDesktop = useIsDesktop();

  return (
    <FramedShell
      width="2xl"
      trailing={
        signedIn && (
          <SessionMenu placement={isDesktop ? "tabStrip" : "tabStripAvatar"} />
        )
      }
    >
      <div className="page-enter">
        {inConsole && (
          <Link
            to="/devices"
            aria-label="Skip for now"
            title="Skip for now"
            className="absolute top-10 right-2 sm:right-6 w-8 h-8 rounded-md flex items-center justify-center text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
          >
            <XMarkIcon className="w-4 h-4" strokeWidth={2} />
          </Link>
        )}
        <ScreenIntro eyebrow={eyebrow} title={title} lead={lead} />
        {children}
      </div>
    </FramedShell>
  );
}
