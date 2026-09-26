import type { ReactNode } from "react";
import { XMarkIcon } from "@heroicons/react/24/outline";
import {
  IconBadge,
  IconButton,
  type Palette,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

/**
 * How a dialog lays out its parts: default puts the icon beside the text, center stacks a larger
 * icon above centred text, for success, recovery codes and upgrade prompts.
 */
export type DialogLayout = "default" | "center";

/**
 * Props of DialogHeader. titleId and descriptionId are the ids the dialog points aria-labelledby
 * and aria-describedby at. onClose adds the close button; without it the header has none.
 */
export interface DialogHeaderProps {
  layout?: DialogLayout;
  icon: ReactNode;
  iconColor?: Palette;
  title: string;
  description: ReactNode;
  titleId: string;
  descriptionId: string;
  onClose?: () => void;
}

/**
 * The top of every dialog: a tinted icon whose colour carries the dialog's gravity, the title,
 * and one line on what the dialog does. The header is the window's drag region inside the desktop
 * app, since an open dialog leaves the app's own chrome inert.
 */
export default function DialogHeader({
  layout = "default",
  icon,
  iconColor = "primary",
  title,
  description,
  titleId,
  descriptionId,
  onClose,
}: DialogHeaderProps) {
  const centered = layout === "center";

  return (
    <div
      data-tauri-drag-region
      className={cn(
        "relative flex gap-4 px-6 pt-6 pb-5 shrink-0",
        centered
          ? "flex-col items-center text-center pt-8 px-10"
          : "items-start",
        onClose && !centered && "pr-14",
      )}
    >
      <IconBadge
        color={iconColor}
        size={centered ? "lg" : "md"}
        className={
          centered
            ? "[&>svg]:w-6 [&>svg]:h-6"
            : "mt-0.5 [&>svg]:w-5 [&>svg]:h-5"
        }
      >
        {icon}
      </IconBadge>
      <div
        data-tauri-drag-region
        className={cn("min-w-0", !centered && "flex-1")}
      >
        <h2
          data-tauri-drag-region
          id={titleId}
          className="text-lg font-semibold leading-snug text-text-primary"
        >
          {title}
        </h2>
        <div
          data-tauri-drag-region
          id={descriptionId}
          className="mt-1 text-sm text-text-muted"
        >
          {description}
        </div>
      </div>
      {onClose && (
        <IconButton
          variant="ghost"
          aria-label="Close"
          data-dismiss
          onClick={onClose}
          className="absolute top-4 right-4"
        >
          <XMarkIcon className="w-5 h-5" />
        </IconButton>
      )}
    </div>
  );
}
