import { useState, useId, type FormEvent } from "react";
import { isSdkError } from "@/api/errors";
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
import {
  NAMESPACE_NAME_MIN_LENGTH,
  validateNamespaceName,
} from "@/utils/validation";
import { getConfig } from "@/env";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";

const FORM_ID = "create-namespace-form";

function CloudForm({
  inputId,
  name,
  setName,
  displayError,
  resetError,
  onSubmit,
}: {
  inputId: string;
  name: string;
  setName: (v: string) => void;
  displayError: string | null;
  resetError: () => void;
  onSubmit: (e: FormEvent) => void;
}) {
  return (
    <form id={FORM_ID} onSubmit={onSubmit}>
      <NamespaceNameField
        id={inputId}
        value={name}
        onChange={(v) => {
          setName(v);
          resetError();
        }}
        error={displayError}
      />
    </form>
  );
}

interface CreateNamespaceDialogProps {
  open: boolean;
  onClose: () => void;
}

export default function CreateNamespaceDialog({
  open,
  onClose,
}: CreateNamespaceDialogProps) {
  const autoId = useId();
  const titleId = `create-ns-title-${autoId}`;
  const inputId = `create-ns-input-${autoId}`;
  // Namespace creation is a premium (Cloud/Enterprise) feature. Community is single-namespace,
  // so this dialog never renders there — the selector shows the upsell instead.
  const isCloud = getConfig().cloud || getConfig().enterprise;

  const [name, setName] = useState("");
  const [validationError, setValidationError] = useState<string | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const createNs = useCreateNamespace();

  const displayError = validationError ?? submitError ?? null;

  const resetError = () => {
    setValidationError(null);
    setSubmitError(null);
    createNs.reset();
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const err = validateNamespaceName(name);
    if (err) {
      setValidationError(err);
      return;
    }
    setValidationError(null);
    setSubmitError(null);
    try {
      await createNs.mutateAsync(name);
      onClose();
    } catch (caught) {
      if (isSdkError(caught)) {
        if (caught.status === 409) {
          setSubmitError("A namespace with this name already exists.");
        } else if (caught.status === 403) {
          setSubmitError(
            "You have reached the namespace limit or do not have permission.",
          );
        } else if (caught.status === 400) {
          setSubmitError("The namespace name is invalid.");
        } else {
          setSubmitError("An unexpected error occurred. Please try again.");
        }
      } else {
        setSubmitError("An unexpected error occurred. Please try again.");
      }
    }
  };

  if (!isCloud) return null;

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
        <CloudForm
          inputId={inputId}
          name={name}
          setName={setName}
          displayError={displayError}
          resetError={resetError}
          onSubmit={(e) => void handleSubmit(e)}
        />
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
            loading={createNs.isPending}
            disabled={name.length < NAMESPACE_NAME_MIN_LENGTH}
          >
            Create
          </Button>
        </div>
      </footer>
    </BaseDialog>
  );
}
