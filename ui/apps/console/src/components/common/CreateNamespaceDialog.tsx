import { useId } from "react";
import { BookOpenIcon, FolderPlusIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import Modal from "./Modal";
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
  const inputId = `create-ns-input-${autoId}`;
  const isPremium = isEnterpriseOrCloud();
  const form = useNamespaceCreateForm(onClose);

  if (!isPremium) return null;

  return (
    <Modal
      open={open}
      onClose={onClose}
      icon={<FolderPlusIcon />}
      title="Create namespace"
      description="A namespace keeps its own devices, members and access policies apart from the rest."
      footerStart={
        <a
          href="https://docs.shellhub.io/self-hosted/administration"
          target="_blank"
          rel="noopener noreferrer"
          className="flex items-center gap-1.5 hover:text-text-secondary transition-colors"
        >
          <BookOpenIcon className="w-3.5 h-3.5" />
          Administration Guide
        </a>
      }
      footer={
        <>
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
        </>
      }
    >
      <form id={FORM_ID} onSubmit={(e) => void form.submit(e)}>
        <NamespaceNameField
          id={inputId}
          value={form.name}
          onChange={form.changeName}
          error={form.error}
        />
      </form>
    </Modal>
  );
}
