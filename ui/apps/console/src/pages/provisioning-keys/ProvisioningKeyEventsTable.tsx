import { useMemo, useState } from "react";
import { useProvisioningKeyEvents } from "@/hooks/useProvisioningKeyEvents";
import { useActionDialog } from "@/hooks/useActionDialog";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import DataTable from "@/components/common/DataTable";
import ActionDialog from "@/components/common/ActionDialog";
import { useDeviceActionRunner } from "@/hooks/useDeviceActionRunner";
import { pageCount } from "@/utils/pagination";
import { getProvisioningKeyEventColumns } from "./provisioningKeyEventColumns";

const PER_PAGE = 15;

const EMPTY_MESSAGE =
  "No registrations yet. Devices that register with this key will appear here.";

/**
 * What a provisioning key has been used for, page by page.
 */
export default function ProvisioningKeyEventsTable({ id }: { id: string }) {
  const [page, setPage] = useState(1);
  const refreshHistory = useInvalidateByIds("provisioningKeyHistory");
  const deviceActions = useActionDialog({
    onSuccess: () => void refreshHistory(),
  });
  const runDeviceAction = useDeviceActionRunner();
  const columns = useMemo(
    () => getProvisioningKeyEventColumns(deviceActions.requestAction),
    [deviceActions.requestAction],
  );
  const { events, totalCount, isLoading, error } = useProvisioningKeyEvents({
    id,
    page,
    perPage: PER_PAGE,
  });

  if (error) {
    return (
      <div
        role="alert"
        className="text-xs text-accent-red bg-accent-red/[0.06] border border-accent-red/20 rounded-lg px-3 py-2.5"
      >
        Could not load registration activity. Check your connection and try
        again.
      </div>
    );
  }

  return (
    <>
      <DataTable
        label="Registration Activity"
        columns={columns}
        data={events}
        rowKey={(event) => event.id}
        rowClassName={() => "[&>td]:py-4"}
        isLoading={isLoading}
        loadingMessage="Loading activity..."
        emptyMessage={EMPTY_MESSAGE}
        page={page}
        totalPages={pageCount(totalCount, PER_PAGE)}
        totalCount={totalCount}
        itemLabel="registration"
        onPageChange={setPage}
      />
      {deviceActions.action && (
        <ActionDialog
          key={deviceActions.actionKey}
          action={deviceActions.action}
          onClose={deviceActions.close}
          onSuccess={deviceActions.handleSuccess}
          entityType="device"
          runAction={runDeviceAction}
        />
      )}
    </>
  );
}
