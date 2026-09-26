import { ReactNode, useId } from "react";
import type { Palette } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import BaseDialog, { type DialogSize } from "@/components/common/BaseDialog";
import ScrollFades from "@/components/common/ScrollFades";
import { useScrollEdges } from "@/hooks/useScrollEdges";
import DialogHeader, {
  type DialogLayout,
} from "@/components/common/DialogHeader";

/**
 * Props of Modal. icon and description are required: a dialog opens by saying what it does.
 * canClose refuses Escape and a backdrop click while it returns false; the close button and
 * onClose itself are the caller's to guard. footerStart sits at the footer's left, across from the
 * actions, for a secondary link such as the docs or another way to do the same thing. The center layout stacks the header and turns the actions into
 * full-width buttons, the last one on top; it has no room for footerStart.
 */
export interface ModalProps {
  layout?: DialogLayout;
  open: boolean;
  onClose: () => void;
  canClose?: () => boolean;
  icon: ReactNode;
  iconColor?: Palette;
  title: string;
  description: ReactNode;
  size?: Exclude<DialogSize, "full">;
  children?: ReactNode;
  footer?: ReactNode;
  footerStart?: ReactNode;
  bodyClassName?: string;
}

/**
 * The dialog for detail and edit flows: the DialogHeader, a body and an optional footer of
 * actions. Only the body scrolls, so the header and the actions stay in view however long the
 * content runs. The panel is the `modal` container, which the fields laid out with container
 * queries read.
 */
export default function Modal({
  layout = "default",
  open,
  onClose,
  canClose,
  icon,
  iconColor,
  title,
  description,
  size = "md",
  children,
  footer,
  footerStart,
  bodyClassName,
}: ModalProps) {
  const headingId = useId();
  const descriptionId = useId();
  const {
    ref: bodyRef,
    moreAbove,
    moreBelow,
  } = useScrollEdges<HTMLDivElement>();

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      canClose={canClose}
      size={size}
      aria-labelledby={headingId}
      aria-describedby={descriptionId}
      className="overflow-hidden @container/modal"
    >
      <DialogHeader
        layout={layout}
        icon={icon}
        iconColor={iconColor}
        title={title}
        description={description}
        titleId={headingId}
        descriptionId={descriptionId}
        onClose={onClose}
      />
      {children != null && (
        <div
          className={cn(
            "relative flex-auto min-h-0 flex flex-col",
            layout === "default" && "border-t border-border",
          )}
        >
          <div
            ref={bodyRef}
            className={cn(
              "flex-auto min-h-0 overflow-y-auto overscroll-contain",
              bodyClassName ??
                (layout === "center" ? "px-8 pb-2" : "px-6 py-5"),
            )}
          >
            {children}
          </div>
          <ScrollFades moreAbove={moreAbove} moreBelow={moreBelow} />
        </div>
      )}
      {layout === "center"
        ? footer && (
            <div className="px-8 pt-4 pb-7 shrink-0 flex flex-col-reverse gap-2 [&>*]:w-full [&>*]:justify-center">
              {footer}
            </div>
          )
        : (footer || footerStart) && (
            <div className="px-6 py-4 border-t border-border shrink-0 flex flex-col-reverse gap-3 sm:flex-row sm:items-center">
              {footerStart && (
                <div className="min-w-0 text-xs text-text-muted sm:mr-auto">
                  {footerStart}
                </div>
              )}
              <div className="flex flex-col-reverse gap-2 sm:flex-row sm:items-center sm:ml-auto">
                {footer}
              </div>
            </div>
          )}
    </BaseDialog>
  );
}
