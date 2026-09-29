import { cn } from "@shellhub/design-system/cn";
import { ShellHubLogo } from "@shellhub/design-system/primitives";

/**
 * The ShellHub logo, clipped to its cloud unless full. It is one drawing either way, so switching
 * full slides the lettering out or in beside a cloud that does not move. Decorative: whatever
 * holds it carries the name. The widths are the drawing's own at h-7, so they change with it. Both
 * places that show it, the sidebar and the tab strip, set its bottom 5px above the 48px strip's
 * floor, which centres it 29px down, on the same line as the tabs' contents and the strip's
 * buttons: TabStrip puts the new-tab button 5px above the floor and centres its trailing group
 * under a 10px top padding, both landing on that same centre. The sidebar's 12.7px left padding
 * centres the cloud in the rail. Moving any of those means moving the others.
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
