import { MinusIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { IconButton } from "@shellhub/design-system/primitives";
import { desktopWindow } from "@/utils/desktopWindow";

function report(action: Promise<void>) {
  action.catch((error: unknown) => {
    console.error("Window control failed", error);
  });
}

interface WindowControlsProps {
  sidebarPinned?: boolean;
  sidebarMode?: "pin" | "drawer";
  onToggleSidebar?: () => void;
}

/**
 * The top-right corner of the app chrome: the sidebar toggle and, inside the desktop app, the
 * window buttons. The browser has its own, so there they are left out.
 */
export default function WindowControls({
  sidebarPinned = false,
  sidebarMode = "pin",
  onToggleSidebar,
}: WindowControlsProps) {
  const win = desktopWindow();
  const pinMode = sidebarMode === "pin";
  const sidebarLabel = pinMode
    ? sidebarPinned
      ? "Unpin sidebar"
      : "Pin sidebar"
    : "Open navigation menu";

  return (
    <div className="flex items-center gap-0.5">
      {onToggleSidebar && (
        <IconButton
          onClick={onToggleSidebar}
          aria-label={sidebarLabel}
          title={sidebarLabel}
          className={cn(
            pinMode && sidebarPinned && "text-primary bg-primary/10",
          )}
        >
          <svg
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth={1.6}
            className="w-4 h-4"
            aria-hidden="true"
          >
            <rect x="3.5" y="4.5" width="17" height="15" rx="2" />
            <path d="M9 4.5v15" />
          </svg>
        </IconButton>
      )}
      {win && (
        <>
          <IconButton
            aria-label="Minimize window"
            title="Minimize"
            onClick={() => report(win.minimize())}
          >
            <MinusIcon className="w-4 h-4" />
          </IconButton>
          <IconButton
            aria-label="Maximize window"
            title="Maximize"
            onClick={() => report(win.toggleMaximize())}
          >
            <svg
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth={1.6}
              className="w-3.5 h-3.5"
              aria-hidden="true"
            >
              <rect x="4.5" y="4.5" width="15" height="15" rx="1.5" />
            </svg>
          </IconButton>
          <IconButton
            aria-label="Close window"
            title="Close"
            onClick={() => report(win.close())}
            className="hover:bg-accent-red/15 hover:text-accent-red"
          >
            <XMarkIcon className="w-4 h-4" />
          </IconButton>
        </>
      )}
    </div>
  );
}
