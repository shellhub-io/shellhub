import {
  useId,
  useState,
  type KeyboardEvent,
  type ReactElement,
  type ReactNode,
} from "react";
import {
  Button,
  Dropdown,
  IconBadge,
  type Palette,
} from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";

/**
 * The compact header of an anchored dialog: a small tinted icon, the title and one line under it.
 * titleId and descriptionId are what the panel's aria-labelledby and aria-describedby point at.
 */
export function ContextualDialogHeader({
  icon,
  iconColor = "primary",
  title,
  description,
  titleId,
  descriptionId,
}: {
  icon: ReactNode;
  iconColor?: Palette;
  title: string;
  description: ReactNode;
  titleId: string;
  descriptionId: string;
}) {
  return (
    <div className="flex items-center gap-2.5 px-3.5 pt-3">
      <IconBadge
        size="sm"
        color={iconColor}
        className="w-7 h-7 rounded-md [&>svg]:w-4 [&>svg]:h-4"
      >
        {icon}
      </IconBadge>
      <div className="min-w-0">
        <h2
          id={titleId}
          className="text-sm font-semibold text-text-primary leading-tight"
        >
          {title}
        </h2>
        <p id={descriptionId} className="text-2xs text-text-muted leading-snug">
          {description}
        </p>
      </div>
    </div>
  );
}

/**
 * The keyboard hint in an anchored dialog's footer: what Enter does, and that Escape closes.
 */
export function KeyHint({ action }: { action: string }) {
  return (
    <span className="text-2xs text-text-muted font-mono">
      <kbd className="px-1 border border-border-light rounded">↵</kbd> {action}
      <kbd className="ml-2 px-1 border border-border-light rounded">
        esc
      </kbd>{" "}
      close
    </span>
  );
}

/**
 * Props of ContextualDialog. onSubmit may be async; a rejection keeps the dialog open and shows
 * the error's message, so a caller throws an Error worded for the user.
 */
export interface ContextualDialogProps {
  trigger: ReactElement;
  icon: ReactNode;
  iconColor?: Palette;
  title: string;
  description: ReactNode;
  submitLabel: string;
  submitDisabled?: boolean;
  width?: "sm" | "md";
  onSubmit: () => Promise<void> | void;
  onOpenChange?: (open: boolean) => void;
  children: ReactNode;
}

function keepInside(e: KeyboardEvent) {
  if (e.key !== "Escape") e.stopPropagation();
}

/**
 * A small dialog anchored to the control that opened it, for a one-field change that shouldn't
 * take the user off the page: the same header as every dialog, scaled down, then the field and
 * the action. Enter submits, Escape and a click outside cancel, except while a submit is pending,
 * so a failure still has a dialog to show in. Focus starts in the field and goes back to the
 * trigger on close. Clicks and keys on the trigger and inside the panel don't reach the page, so
 * the dialog works from a row that opens on click, Enter or Space; Escape still does.
 */
export default function ContextualDialog({
  trigger,
  icon,
  iconColor = "primary",
  title,
  description,
  submitLabel,
  submitDisabled,
  width = "sm",
  onSubmit,
  onOpenChange,
  children,
}: ContextualDialogProps) {
  const [open, setOpen] = useState(false);
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const titleId = useId();
  const descriptionId = useId();

  const changeOpen = (next: boolean) => {
    if (!next && pending) return;
    setOpen(next);
    if (next) setError(null);
    onOpenChange?.(next);
  };

  const submit = async () => {
    if (pending || submitDisabled) return;
    setPending(true);
    setError(null);
    try {
      await onSubmit();
      changeOpen(false);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Something went wrong.");
    } finally {
      setPending(false);
    }
  };

  return (
    <Dropdown
      mode="content"
      portal
      placement="bottom-start"
      open={open}
      onOpenChange={changeOpen}
    >
      <span
        role="presentation"
        className="inline-flex"
        onClick={(e) => e.stopPropagation()}
        onKeyDown={keepInside}
      >
        <Dropdown.Trigger>{trigger}</Dropdown.Trigger>
      </span>
      <Dropdown.Panel
        aria-labelledby={titleId}
        aria-describedby={descriptionId}
        className={cn(
          "bg-card border-border-light",
          width === "sm" ? "w-80" : "w-[26rem]",
        )}
        onClick={(e) => e.stopPropagation()}
        onKeyDown={keepInside}
      >
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void submit();
          }}
        >
          <ContextualDialogHeader
            icon={icon}
            iconColor={iconColor}
            title={title}
            description={description}
            titleId={titleId}
            descriptionId={descriptionId}
          />
          <div className="px-3.5 py-3 space-y-2">
            {children}
            {error && (
              <p role="alert" className="text-2xs text-accent-red">
                {error}
              </p>
            )}
          </div>
          <div className="flex items-center justify-between gap-3 px-3.5 py-2.5 border-t border-border">
            <KeyHint action="save" />
            <Button
              type="submit"
              size="sm"
              loading={pending}
              disabled={submitDisabled}
            >
              {submitLabel}
            </Button>
          </div>
        </form>
      </Dropdown.Panel>
    </Dropdown>
  );
}
