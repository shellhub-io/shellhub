import type { ReactNode } from "react";
import type {
  FieldValues,
  SubmitHandler,
  UseFormReturn,
} from "react-hook-form";
import { Button } from "@shellhub/design-system/primitives";
import Modal, { type ModalProps } from "@/components/common/Modal";
import FormRootError from "@/components/common/fields/FormRootError";
import { useDiscardGuard } from "@/hooks/useDiscardGuard";

interface FormModalProps<T extends FieldValues> {
  form: UseFormReturn<T>;
  onSubmit: SubmitHandler<T>;
  open: boolean;
  onClose: () => void;
  title: string;
  submitLabel: string;
  submittingLabel?: string;
  cancelLabel?: string;
  submitIcon?: ReactNode;
  requireDirty?: boolean;
  submitDisabled?: boolean;
  description: ReactNode;
  icon: ReactNode;
  size?: ModalProps["size"];
  bodyClassName?: string;
  footerExtra?: ReactNode;
  children: ReactNode;
}

/**
 * Modal + form plumbing shared by every modal form: it owns the `<form>`
 * element, the Cancel/Submit footer (wired through `handleSubmit`, disabled on
 * invalid/submitting), and the `root` error rendering. Callers only provide the
 * form object, the submit handler, and the fields as `children`, so no modal,
 * submit, or error boilerplate is re-implemented per form. Closing a dirty form
 * asks first, through useDiscardGuard.
 */
export default function FormModal<T extends FieldValues>({
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
  description,
  icon,
  size,
  bodyClassName,
  footerExtra,
  children,
}: FormModalProps<T>) {
  const {
    handleSubmit,
    formState: { isValid, isSubmitting, isDirty, errors },
  } = form;

  const { requestClose, prompt } = useDiscardGuard({
    open,
    dirty: isDirty && !isSubmitting,
    onClose,
  });

  const isSubmitDisabled =
    !isValid || isSubmitting || (requireDirty && !isDirty) || submitDisabled;

  return (
    <>
      <Modal
        open={open}
        onClose={requestClose}
        title={title}
        description={description}
        icon={icon}
        size={size}
        bodyClassName={bodyClassName}
        footer={
          <>
            {footerExtra}
            <Button variant="ghost" onClick={requestClose}>
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
      </Modal>
      {prompt}
    </>
  );
}
