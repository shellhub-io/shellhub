import { useId, useState, type ReactNode } from "react";
import { PencilSquareIcon } from "@heroicons/react/24/outline";
import { IconButton } from "@shellhub/design-system/primitives";
import ContextualDialog from "@/components/common/ContextualDialog";
import type { RenameSectionProps } from "@/components/common/RenameSection";
import InputField from "@/components/common/fields/InputField";
import { renameErrorMessage, renameFieldLabel } from "@/utils/rename";

/**
 * Props of RenameDialog. rename is passed in, so the same dialog serves devices and containers.
 * preview shows what the new name turns into as it is typed, such as the SSHID.
 */
export interface RenameDialogProps {
  uid: string;
  currentName: string;
  entityLabel: string;
  description: ReactNode;
  rename: RenameSectionProps["rename"];
  preview?: (name: string) => ReactNode;
}

/**
 * Renames a device or container from its row in a list, through a ContextualDialog anchored to a
 * pencil button. An empty or unchanged name closes without a request.
 */
export default function RenameDialog({
  uid,
  currentName,
  entityLabel,
  description,
  rename,
  preview,
}: RenameDialogProps) {
  const [name, setName] = useState(currentName);
  const inputId = useId();
  const next = name.trim();

  return (
    <ContextualDialog
      trigger={
        <IconButton
          variant="ghost"
          size="sm"
          aria-label={`Rename ${entityLabel}`}
          title={`Rename ${entityLabel}`}
        >
          <PencilSquareIcon className="w-4 h-4" />
        </IconButton>
      }
      onOpenChange={(open) => {
        if (open) setName(currentName);
      }}
      icon={<PencilSquareIcon />}
      title={`Rename ${entityLabel}`}
      description={description}
      submitLabel="Rename"
      onSubmit={async () => {
        if (!next || next === currentName) return;
        try {
          await rename({ path: { uid }, body: { name: next } });
        } catch (err) {
          throw new Error(renameErrorMessage(err, entityLabel));
        }
      }}
    >
      <InputField
        id={inputId}
        label={renameFieldLabel(entityLabel)}
        hideLabel
        value={name}
        onChange={setName}
        variant="mono"
      />
      {preview && (
        <p className="text-2xs font-mono text-text-muted truncate">
          {preview(next || currentName)}
        </p>
      )}
    </ContextualDialog>
  );
}
