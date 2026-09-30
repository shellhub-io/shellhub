import type { ReactNode } from "react";
import { cn } from "@shellhub/design-system/cn";
import LogoMark from "./LogoMark";
import WindowControls from "./WindowControls";
import { columnWidths, type ColumnWidth } from "./columnWidths";

interface FramedShellProps {
  trailing?: ReactNode;
  width?: ColumnWidth;
  children: ReactNode;
}

/**
 * The app's chrome without the sidebar: the logo strip on top and the framed panel under it,
 * for the screens that stand outside the console, before sign-in and before the first device.
 * The strip keeps the window controls and drag region the desktop app needs, and takes what a
 * screen wants beside them as `trailing`. The panel scrolls and centres the screen in a column
 * of the form's width, or the wide one when a screen asks for it.
 */
export default function FramedShell({
  trailing,
  width = "md",
  children,
}: FramedShellProps) {
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
          {trailing}
          <WindowControls />
        </div>
      </header>
      <div className="relative flex-1 min-h-0 mx-2 mb-2 overflow-hidden rounded-[10px] border border-border bg-surface">
        <div className="grid-bg scanline absolute inset-0 z-bg" />
        <main
          id="main-content"
          tabIndex={-1}
          className="absolute inset-0 overflow-y-auto outline-none"
        >
          <div
            className={cn(
              "relative w-full mx-auto px-4 sm:px-8 pt-12 pb-20",
              columnWidths[width],
            )}
          >
            {children}
          </div>
        </main>
      </div>
    </div>
  );
}
