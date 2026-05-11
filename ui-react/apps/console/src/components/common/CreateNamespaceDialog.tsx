import { useState, useId, type FormEvent } from "react";
import {
  XMarkIcon,
  BookOpenIcon,
  FolderPlusIcon,
} from "@heroicons/react/24/outline";
import BaseDialog from "./BaseDialog";
import CopyButton from "./CopyButton";
import InputField from "./fields/InputField";
import { getConfig } from "@/env";
import { useCreateNamespace } from "@/hooks/useNamespaceMutations";

const CLI_COMMAND = "./bin/cli namespace create <namespace> <owner>";

const NAME_REGEX = /^[a-z0-9][a-z0-9-]*[a-z0-9]$/;

function validate(name: string): string | null {
  if (name.length < 3) return "Name must be at least 3 characters";
  if (name.length > 30) return "Name must be at most 30 characters";
  if (!NAME_REGEX.test(name))
    return "Only lowercase letters, numbers, and hyphens (cannot start or end with hyphen)";
  return null;
}

const rules = [
  "3–30 characters",
  "Lowercase letters, numbers, and hyphens only",
  "Cannot begin or end with a hyphen",
];

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
      <InputField
        id={inputId}
        label="Namespace Name"
        value={name}
        onChange={(v) => {
          setName(v.toLowerCase());
          resetError();
        }}
        placeholder="my-namespace"
        error={displayError ?? undefined}
        hint="3–30 characters · lowercase letters, numbers, and hyphens only"
        maxLength={30}
        autoFocus
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
      <div className="bg-card border border-border rounded-xl overflow-hidden">
        <div className="flex items-center justify-between px-4 py-2.5 border-b border-border bg-surface/50">
          <div className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-accent-red/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-accent-yellow/60" />
            <span className="w-2.5 h-2.5 rounded-full bg-accent-green/60" />
          </div>
          <span className="text-2xs font-mono text-text-muted/50">
            terminal
          </span>
          <CopyButton text={CLI_COMMAND} showLabel />
        </div>
        <div className="p-4 overflow-x-auto">
          <pre className="text-xs font-mono text-accent-cyan leading-relaxed whitespace-pre m-0">
            <span className="text-text-muted select-none">$ </span>
            {CLI_COMMAND}
          </pre>
        </div>
      </div>

      {/* Name rules */}
      <div>
        <p className="text-2xs font-mono font-semibold uppercase tracking-label text-text-muted mb-2.5">
          Naming rules
        </p>
        <ul className="space-y-1.5" aria-label="Namespace naming rules">
          {rules.map((rule) => (
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
  const createNs = useCreateNamespace();

  const displayError =
    validationError ??
    (createNs.error instanceof Error ? createNs.error.message : null);

  const resetError = () => {
    setValidationError(null);
    createNs.reset();
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    const err = validate(name);
    if (err) {
      setValidationError(err);
      return;
    }
    setValidationError(null);
    try {
      await createNs.mutateAsync(name);
      onClose();
    } catch {
      // error is surfaced via displayError
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
          <span className="w-8 h-8 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shrink-0">
            <FolderPlusIcon className="w-4 h-4" />
          </span>
          <h2 id={titleId} className="text-sm font-semibold text-text-primary">
            Create a Namespace
          </h2>
        </div>

        <button
          onClick={onClose}
          className="p-1.5 rounded-lg text-text-muted hover:text-text-primary hover:bg-hover-medium transition-all"
          aria-label="Close dialog"
        >
          <XMarkIcon className="w-4 h-4" />
        </button>
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
          <button
            onClick={onClose}
            className="px-4 py-2 rounded-lg text-xs font-medium text-text-secondary hover:text-text-primary hover:bg-hover-medium transition-all"
          >
            {isCloud ? "Cancel" : "Close"}
          </button>
          {isCloud && (
            <button
              type="submit"
              form={FORM_ID}
              disabled={createNs.isPending || name.length < 3}
              className="px-4 py-2 bg-primary hover:bg-primary/90 text-white rounded-lg text-xs font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
            >
              {createNs.isPending ? (
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin inline-block" />
              ) : (
                "Create"
              )}
            </button>
          )}
        </div>
      </footer>
    </BaseDialog>
  );
}
