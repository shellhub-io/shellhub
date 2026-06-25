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
  WindowChrome,
} from "@shellhub/design-system/primitives";
import BaseDialog from "./BaseDialog";
import CopyButton from "./CopyButton";
import NamespaceNameField from "./fields/NamespaceNameField";
import {
  NAMESPACE_NAME_MIN_LENGTH,
  NAMESPACE_NAME_RULES,
  validateNamespaceName,
} from "@/utils/validation";
import { getConfig } from "@/env";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";
const CLI_COMMAND = "./bin/cli namespace create <namespace> <owner>";

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

function CeInstructions({ descriptionId }: { descriptionId: string }) {
  return (
    <>
      <p id={descriptionId} className="text-sm text-text-muted leading-relaxed">
        Community Edition uses the CLI to manage namespaces. Run this command on
        your server:
      </p>

      {/* Command block */}
      <WindowChrome
        variant="terminal"
        size="sm"
        titleBarSlot={<CopyButton text={CLI_COMMAND} showLabel />}
      >
        <pre className="text-accent-cyan whitespace-pre-wrap break-all m-0">
          <span className="text-text-muted select-none">$ </span>
          {CLI_COMMAND}
        </pre>
      </WindowChrome>

      {/* Name rules */}
      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
          Naming rules
        </p>
        <ul className="space-y-1.5" aria-label="Namespace naming rules">
          {NAMESPACE_NAME_RULES.map((rule) => (
            <li
              key={rule}
              className="flex items-start gap-2 text-xs text-text-muted"
            >
              <span
                className="w-1 h-1 rounded-full bg-border-light mt-1.5 shrink-0"
                aria-hidden="true"
              />
              {rule}
            </li>
          ))}
        </ul>
      </div>
    </>
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
  const descriptionId = `create-ns-description-${autoId}`;
  const inputId = `create-ns-input-${autoId}`;
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
          setSubmitError("You have reached the namespace limit or do not have permission.");
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

  return (
    <BaseDialog
      open={open}
      onClose={onClose}
      size="lg"
      aria-labelledby={titleId}
      aria-describedby={isCloud ? undefined : descriptionId}
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
        {isCloud ? (
          <CloudForm
            inputId={inputId}
            name={name}
            setName={setName}
            displayError={displayError}
            resetError={resetError}
            onSubmit={(e) => void handleSubmit(e)}
          />
        ) : (
          <CeInstructions descriptionId={descriptionId} />
        )}
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
            {isCloud ? "Cancel" : "Close"}
          </Button>
          {isCloud && (
            <Button
              type="submit"
              form={FORM_ID}
              loading={createNs.isPending}
              disabled={name.length < NAMESPACE_NAME_MIN_LENGTH}
            >
              Create
            </Button>
          )}
        </div>
      </footer>
    </BaseDialog>
  );
}
