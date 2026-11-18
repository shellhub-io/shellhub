import { RectangleStackIcon } from "@heroicons/react/24/outline";
import UpsellDialog from "@/components/common/UpsellDialog";

const HIGHLIGHTS = [
  "Separate namespaces for teams, environments, or customers",
  "Members and roles scoped to each namespace",
  "Devices and access policies isolated per namespace",
];

interface NamespaceUpsellDialogProps {
  open: boolean;
  onClose: () => void;
}

/**
 * Explains why another namespace needs a paid edition, shown instead of failing the create.
 */
export default function NamespaceUpsellDialog({
  open,
  onClose,
}: NamespaceUpsellDialogProps) {
  return (
    <UpsellDialog
      open={open}
      onClose={onClose}
      icon={<RectangleStackIcon />}
      title="Run more than one namespace"
      description="This instance runs as a single namespace. Upgrade to keep separate workspaces, each with its own devices, members, and access policies."
      highlights={HIGHLIGHTS}
    />
  );
}
