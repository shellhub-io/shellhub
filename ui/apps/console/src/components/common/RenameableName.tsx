import RenameDialog, {
  type RenameDialogProps,
} from "@/components/common/RenameDialog";
import { buildSshid } from "@/utils/sshid";

/**
 * A device or container name in a list row, with the rename button beside it. The button shows
 * while the row is hovered or focused, and stays while its dialog is open. nsName builds the SSHID
 * preview; when it is empty the preview is the bare name.
 */
export default function RenameableName({
  uid,
  name,
  entityLabel,
  rename,
  canRename,
  nsName,
}: {
  uid: string;
  name: string;
  entityLabel: string;
  rename: RenameDialogProps["rename"];
  canRename: boolean;
  nsName: string;
}) {
  return (
    <span className="flex items-center gap-1.5">
      <span className="text-sm font-medium text-text-primary group-hover:text-primary transition-colors">
        {name}
      </span>
      {canRename && (
        <span className="opacity-0 group-hover:opacity-100 focus-within:opacity-100 has-[[aria-expanded=true]]:opacity-100 transition-opacity">
          <RenameDialog
            uid={uid}
            currentName={name}
            entityLabel={entityLabel}
            description="Its SSHID changes with the name."
            rename={rename}
            preview={(next) => (nsName ? buildSshid(nsName, next) : next)}
          />
        </span>
      )}
    </span>
  );
}
