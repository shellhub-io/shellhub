import { isEnterpriseOrCloud } from "@/env";
import { useCreateNamespaceStore } from "@/stores/createNamespaceStore";
import CreateNamespaceDialog from "../common/CreateNamespaceDialog";
import NamespaceUpsellDialog from "../common/NamespaceUpsellDialog";

/**
 * Where the palette's Create namespace lands. It lives in the layout because the palette closes
 * and unmounts as it opens it; the Community edition gets the upgrade dialog instead.
 */
export default function CreateNamespaceHost() {
  const open = useCreateNamespaceStore((s) => s.open);
  const close = useCreateNamespaceStore((s) => s.closeDialog);

  return isEnterpriseOrCloud() ? (
    <CreateNamespaceDialog open={open} onClose={close} />
  ) : (
    <NamespaceUpsellDialog open={open} onClose={close} />
  );
}
