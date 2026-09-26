import { ReactNode, useId, useState } from "react";
import { ExclamationTriangleIcon } from "@heroicons/react/24/outline";
import { ClipboardContext } from "@/hooks/useCopy";
import BaseDialog from "./BaseDialog";
import { Button } from "@shellhub/design-system/primitives";
import DialogHeader from "@/components/common/DialogHeader";

/**
 * Mounts a single clipboard-warning dialog for the whole app.
 * Wrap the app root with this once; every `useCopy()` call shares it.
 */
export function ClipboardProvider({ children }: { children: ReactNode }) {
  const [showDialog, setShowDialog] = useState(false);
  const titleId = useId();
  const descId = useId();

  const triggerWarning = () => setShowDialog(true);
  const handleClose = () => setShowDialog(false);
  const ctxValue = { triggerWarning };

  return (
    <ClipboardContext.Provider value={ctxValue}>
      {children}

      <BaseDialog
        open={showDialog}
        onClose={handleClose}
        size="sm"
        aria-labelledby={titleId}
        aria-describedby={descId}
      >
        <DialogHeader
          icon={<ExclamationTriangleIcon />}
          iconColor="yellow"
          title="Copying is not allowed"
          description="The clipboard only works on HTTPS or localhost. Serve this instance over HTTPS to copy from it."
          titleId={titleId}
          descriptionId={descId}
        />

        <div className="flex justify-end px-6 py-4 border-t border-border">
          <Button data-testid="copy-warning-ok-btn" onClick={handleClose}>
            OK
          </Button>
        </div>
      </BaseDialog>
    </ClipboardContext.Provider>
  );
}
