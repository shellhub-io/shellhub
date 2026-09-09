import { useId } from "react";
import {
  XMarkIcon,
  BookOpenIcon,
  FolderPlusIcon,
} from "@heroicons/react/24/outline";
import {
  Button,
  IconBadge,
  IconButton,
} from "@shellhub/design-system/primitives";
import BaseDialog from "./BaseDialog";
import NamespaceNameField from "./fields/NamespaceNameField";
import { NAMESPACE_NAME_MIN_LENGTH } from "@/utils/validation";
import { isEnterpriseOrCloud } from "@/env";
import { useNamespaceCreateForm } from "@/hooks/useNamespaceCreateForm";

const FORM_ID = "create-namespace-form";

interface CreateNamespaceDialogProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Creates a namespace without leaving the current page, then switches into it.
 */
export default function CreateNamespaceDialog({
  open,
  onClose,
}: CreateNamespaceDialogProps) {
  const autoId = useId();
  const titleId = `create-ns-title-${autoId}`;
  const inputId = `create-ns-input-${autoId}`;
  const isPremium = isEnterpriseOrCloud();
  const form = useNamespaceCreateForm(onClose);

  if (!isPremium) return null;

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="lg"
      aria-labelledby={titleId}
    >
      {/* Header */}
      <header className="flex items-center justify-between px-6 pt-5 pb-4 border-b border-border shrink-0">
        <div className="flex items-center gap-3">
          <IconBadge size="sm">
            <FolderPlusIcon className="w-4 h-4" />
          </IconBadge>
          <h2 id={titleId} className="text-sm font-semibold text-text-primary">
            Create a Namespace
          </h2>
        </div>

        <IconButton onClick={onClose} aria-label="Close dialog">
          <XMarkIcon className="w-4 h-4" />
        </IconButton>
      </header>

      {/* Body */}
      <div className="px-6 py-5 space-y-5">
        <form id={FORM_ID} onSubmit={(e) => void form.submit(e)}>
          <NamespaceNameField
            id={inputId}
            value={form.name}
            onChange={form.changeName}
            error={form.error}
          />
        </form>
      </div>

      {/* Footer */}
      <footer className="px-6 py-4 border-t border-border flex items-center justify-between shrink-0">
        <a
          href="https://docs.shellhub.io/self-hosted/administration"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1.5 text-xs text-text-muted hover:text-text-secondary transition-colors"
        >
          <BookOpenIcon className="w-3.5 h-3.5" />
          Administration Guide
        </a>

        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            type="submit"
            form={FORM_ID}
            loading={form.isPending}
            disabled={form.name.length < NAMESPACE_NAME_MIN_LENGTH}
          >
            Create
          </Button>
        </div>
      </footer>
    </BaseDialog>
  );
}
