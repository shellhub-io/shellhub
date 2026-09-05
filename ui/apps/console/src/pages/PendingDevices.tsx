import { useNavigate, Link } from "react-router-dom";
import { ClockIcon, TicketIcon } from "@heroicons/react/24/outline";
import { Button, Callout } from "@shellhub/design-system/primitives";
import { useDevices, type NormalizedDevice } from "@/hooks/useDevices";
import { useEnrollmentSourceResolver } from "@/hooks/useEnrollmentSource";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import { useActionDialog } from "@/hooks/useActionDialog";
import { useDeviceActionRunner } from "@/hooks/useDeviceActionRunner";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import ActionDialog from "@/components/common/ActionDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import DistroIcon from "@/components/common/DistroIcon";
import KeyValueChip from "@/components/common/KeyValueChip";
import PlatformBadge from "@/components/common/PlatformBadge";
import { type EnrollmentSource } from "@/pages/install-keys/helpers";
import { apiErrorMessage } from "@/api/errors";
import { formatDateFull } from "@/utils/date";
import { PER_PAGE, pageCount } from "@/utils/pagination";

function EnrollmentSourceCell({
  installKeyId,
  enrollment,
}: {
  installKeyId?: string;
  enrollment: EnrollmentSource | null;
}) {
  if (!enrollment || !installKeyId) {
    return <span className="text-2xs text-text-muted">&mdash;</span>;
  }

  if (enrollment.kind === "key") {
    return (
      <Link
        to={`/install-keys/${encodeURIComponent(installKeyId)}/activity`}
        onClick={(e) => e.stopPropagation()}
        className="text-xs font-medium text-text-primary hover:text-primary hover:underline"
      >
        {enrollment.name}
      </Link>
    );
  }

  return (
    <span className="text-xs text-text-secondary">
      {enrollment.kind === "pairing"
        ? "Pairing code"
        : "Tenant-only registration"}
    </span>
  );
}

type PendingParams = { page: number };

const DEFAULTS: PendingParams = { page: 1 };

/**
 * The review queue: every device in the namespace still awaiting a decision, whichever install key
 * enrolled it, so a pending registration is found without opening each key's activity in turn.
 */
export default function PendingDevices() {
  const { params, setPage } = usePaginatedListState<PendingParams>({
    defaults: DEFAULTS,
  });
  const navigate = useNavigate();

  const { devices, totalCount, isLoading, error } = useDevices({
    page: params.page,
    perPage: PER_PAGE,
    status: "pending",
    sortBy: "created_at",
    orderBy: "desc",
  });

  const resolveEnrollment = useEnrollmentSourceResolver();
  const deviceActions = useActionDialog();
  const runDeviceAction = useDeviceActionRunner();

  const columns: Column<NormalizedDevice>[] = [
    {
      key: "name",
      header: "Device",
      render: (device) => (
        <div className="flex items-center gap-2.5">
          <DistroIcon
            id={device.info?.id ?? ""}
            className="text-[1.05rem] leading-none text-text-secondary shrink-0"
          />
          <div className="min-w-0">
            <div className="text-sm font-medium text-text-primary">
              {device.name}
            </div>
            {device.info?.pretty_name && (
              <div className="mt-0.5 text-2xs text-text-muted">
                {device.info.pretty_name}
              </div>
            )}
            <div className="mt-2 flex flex-wrap items-center gap-1.5">
              {device.identity?.mac && (
                <KeyValueChip label="MAC" value={device.identity.mac} />
              )}
              {device.remote_addr && (
                <KeyValueChip label="IP" value={device.remote_addr} />
              )}
              {device.info?.arch && (
                <KeyValueChip label="Arch" value={device.info.arch} />
              )}
              {device.info?.platform && (
                <PlatformBadge platform={device.info.platform} />
              )}
            </div>
          </div>
        </div>
      ),
    },
    {
      key: "created_at",
      header: "Registered",
      render: (device) => (
        <span className="text-2xs font-mono text-text-muted whitespace-nowrap">
          {formatDateFull(device.created_at)}
        </span>
      ),
    },
    {
      key: "install_key",
      header: "Install Key",
      render: (device) => (
        <EnrollmentSourceCell
          installKeyId={device.install_key_id}
          enrollment={resolveEnrollment(device.install_key_id)}
        />
      ),
    },
    {
      key: "actions",
      header: "Review",
      headerClassName: "text-right",
      render: (device) => (
        <div className="flex items-center justify-end gap-1.5">
          <RestrictedAction action="device:accept">
            <Button
              variant="successSoft"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                deviceActions.requestAction(device, "accept");
              }}
            >
              Accept
            </Button>
          </RestrictedAction>
          <RestrictedAction action="device:reject">
            <Button
              variant="warningSoft"
              size="sm"
              onClick={(e) => {
                e.stopPropagation();
                deviceActions.requestAction(device, "reject");
              }}
            >
              Reject
            </Button>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ClockIcon className="w-6 h-6" />}
        overline="Provisioning"
        title="Pending Devices"
        description="Devices that registered and are waiting to be accepted into the namespace."
      >
        <Button
          as={Link}
          to="/install-keys"
          variant="ghost"
          icon={<TicketIcon className="w-4 h-4" strokeWidth={2} />}
        >
          Install Keys
        </Button>
      </PageHeader>

      {error ? (
        <Callout variant="error">{apiErrorMessage(error)}</Callout>
      ) : (
        <DataTable
          columns={columns}
          data={devices}
          rowKey={(device) => device.uid}
          isLoading={isLoading}
          loadingMessage="Loading pending devices..."
          page={params.page}
          totalPages={pageCount(totalCount)}
          totalCount={totalCount}
          itemLabel="device"
          onPageChange={setPage}
          onRowClick={(device) => void navigate(`/devices/${device.uid}`)}
          rowClassName={() => "[&>td]:py-4"}
          emptyState={
            <div className="text-center">
              <ClockIcon
                className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
                strokeWidth={1}
              />
              <p className="text-xs font-mono text-text-muted">
                No devices are waiting for review
              </p>
            </div>
          }
        />
      )}

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
    </div>
  );
}
