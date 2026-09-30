import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { XMarkIcon } from "@heroicons/react/24/outline";
import SessionMenu from "@/components/layout/SessionMenu";
import { useIsDesktop } from "@/hooks/useIsDesktop";
import LogoMark from "@/components/layout/LogoMark";
import WindowControls from "@/components/layout/WindowControls";
import FirstRunIntro from "./FirstRunIntro";

interface FirstRunLayoutProps {
  eyebrow: string;
  title?: string;
  lead?: string;
  signedIn: boolean;
  inConsole: boolean;
  children: ReactNode;
}

/**
 * The focused screen the first run happens on: the app's chrome and framed panel, without the
 * sidebar, since every page it leads to is empty until a device exists. The chrome keeps the window
 * controls and drag region the desktop app needs. A signed-in user gets the account menu; once
 * they are inside a namespace (`inConsole`), the page also offers to skip into the console for this
 * visit. Nothing remembers the skip: the trail comes back until the namespace has a device.
 */
export default function FirstRunLayout({
  eyebrow,
  title,
  lead,
  signedIn,
  inConsole,
  children,
}: FirstRunLayoutProps) {
  const isDesktop = useIsDesktop();

  return (
    <div className="flex flex-col h-screen bg-background">
      <header
        data-tauri-drag-region
        className="h-12 shrink-0 flex items-end min-w-0 pl-[12.7px] pr-2"
      >
        <span className="sr-only">ShellHub</span>
        <LogoMark full className="mb-[5px] shrink-0" />
        <div
          data-tauri-drag-region
          className="ml-auto self-stretch pt-2.5 flex items-center gap-2 shrink-0"
        >
          {signedIn && (
            <SessionMenu
              placement={isDesktop ? "tabStrip" : "tabStripAvatar"}
            />
          )}
          <WindowControls />
        </div>
      </header>
      <div className="relative flex-1 min-h-0 mx-2 mb-2 overflow-hidden rounded-[10px] border border-border bg-surface">
        <div className="grid-bg scanline absolute inset-0 z-bg" />
        <main
          id="main-content"
          tabIndex={-1}
          className="page-enter absolute inset-0 overflow-y-auto outline-none"
        >
          <div className="relative w-full max-w-2xl mx-auto px-4 sm:px-8 pt-12 pb-20">
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
            <FirstRunIntro eyebrow={eyebrow} title={title} lead={lead} />
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
