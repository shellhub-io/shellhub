import { cn } from "@shellhub/design-system/cn";
import { ShellHubLogo } from "@shellhub/design-system/primitives";

/**
 * The ShellHub logo, clipped to its cloud unless full. It is one drawing either way, so switching
 * full slides the lettering out or in beside a cloud that does not move. Decorative: whatever
 * holds it carries the name. The widths are the drawing's own at h-7, so they change with it. Both
 * places that show it, the sidebar and the tab strip, set its bottom 7px above the 48px strip's
 * floor, where the lettering meets the tabs' baseline, and the sidebar's 12.7px left padding
 * centres the cloud in the rail; moving one of those means moving the other.
 */
export default function LogoMark({
  full,
  className,
}: {
  full: boolean;
  className?: string;
}) {
  return (
    <span
      aria-hidden="true"
      className={cn(
        "block overflow-hidden transition-[width] duration-200 ease-in-out",
        full ? "w-[106.7px]" : "w-[33px]",
        className,
      )}
    >
      <ShellHubLogo className="h-7 max-w-none" />
    </span>
  );
}
