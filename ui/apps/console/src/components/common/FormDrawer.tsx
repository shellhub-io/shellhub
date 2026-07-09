import type { ReactNode } from "react";
import type {
  FieldValues,
  SubmitHandler,
  UseFormReturn,
} from "react-hook-form";
import { Button } from "@shellhub/design-system/primitives";
import Drawer from "@/components/common/Drawer";
import FormRootError from "@/components/common/fields/FormRootError";

interface FormDrawerProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  onSubmit: SubmitHandler<T>;
  open: boolean;
  onClose: () => void;
  title: string;
  submitLabel: string;
  /** Label shown on the submit button while `isSubmitting`. */
  submittingLabel?: string;
  cancelLabel?: string;
  submitIcon?: ReactNode;
  /** Disable submit until the form is dirty (edit-in-place forms). */
  requireDirty?: boolean;
  /** Extra submit gate, ANDed with validity (e.g. "all fields filled"). */
  submitDisabled?: boolean;
  subtitle?: ReactNode;
  icon?: ReactNode;
  width?: "sm" | "md";
  bodyClassName?: string;
  /** Extra footer content rendered before Cancel/Submit. */
  footerExtra?: ReactNode;
  children: ReactNode;
}

/**
 * Drawer + form plumbing shared by every drawer form: it owns the `<form>`
 * element, the Cancel/Submit footer (wired through `handleSubmit`, disabled on
 * invalid/submitting), and the `root` error rendering. Callers only provide the
 * form object, the submit handler, and the fields as `children` — no drawer,
 * submit, or error boilerplate is re-implemented per form.
 */
export default function FormDrawer<T extends FieldValues>({
  form,
  onSubmit,
  open,
  onClose,
  title,
  submitLabel,
  submittingLabel = "Saving...",
  cancelLabel = "Cancel",
  submitIcon,
  requireDirty = false,
  submitDisabled = false,
  subtitle,
  icon,
  width,
  bodyClassName,
  footerExtra,
  children,
}: FormDrawerProps<T>) {
  const {
    handleSubmit,
    formState: { isValid, isSubmitting, isDirty, errors },
  } = form;

  const isSubmitDisabled =
    !isValid || isSubmitting || (requireDirty && !isDirty) || submitDisabled;

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={title}
      subtitle={subtitle}
      icon={icon}
      width={width}
      bodyClassName={bodyClassName}
      footer={
        <>
          {footerExtra}
          <Button variant="ghost" onClick={onClose}>
            {cancelLabel}
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit(onSubmit)()}
            disabled={isSubmitDisabled}
            loading={isSubmitting}
            icon={submitIcon}
          >
            {isSubmitting ? submittingLabel : submitLabel}
          </Button>
        </>
      }
    >
      <form
        onSubmit={(e) => void handleSubmit(onSubmit)(e)}
        className="space-y-5"
      >
        {children}
        <FormRootError message={errors.root?.message} />
      </form>
    </Drawer>
  );
}
