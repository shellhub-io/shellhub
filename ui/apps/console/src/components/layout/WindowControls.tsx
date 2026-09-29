import { MinusIcon, XMarkIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import { desktopWindow } from "@/utils/desktopWindow";

function report(action: Promise<void>) {
  action.catch((error: unknown) => {
    console.error("Window control failed", error);
  });
}

interface WindowControlsProps {
  onOpenNavigation?: () => void;
}

/**
 * The top-right corner of the app chrome: the button that opens the navigation drawer on a narrow
 * window, given onOpenNavigation, and, inside the desktop app, the window buttons. The browser has
 * its own, so there they are left out, and with neither there is nothing to render.
 */
export default function WindowControls({
  onOpenNavigation,
}: WindowControlsProps) {
  const win = desktopWindow();
  if (!onOpenNavigation && !win) return null;

  return (
    <div className="flex items-center gap-0.5">
      {onOpenNavigation && (
        <IconButton
          onClick={onOpenNavigation}
          aria-label="Open navigation menu"
          title="Open navigation menu"
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
