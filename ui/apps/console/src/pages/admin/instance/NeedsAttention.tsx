import { useId, useState, type ReactNode } from "react";
import { Link } from "react-router-dom";
import {
  ArrowRightIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  UserPlusIcon,
} from "@heroicons/react/24/outline";
import { Badge, Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import type { UserAdminResponse } from "@/client";
import { useAdminUsers } from "@/hooks/useAdminUsers";
import { useAwaitingMemberNamespaces } from "@/hooks/useAdminNamespaces";
import { useDeviceCapacity } from "@/hooks/useDeviceCapacity";
import { formatCount } from "@/utils/count";
import { formatRelative } from "@/utils/date";
import {
  formatLicenseTimestamp,
  getLicenseAlertConfig,
  type DeviceCapacity,
} from "@/utils/license";
import ApproveAccountDialog from "../users/ApproveAccountDialog";
import RejectAccountDialog from "../users/RejectAccountDialog";
import { Panel, PanelError } from "./Panel";

const PENDING_PREVIEW = 5;
const NAMESPACE_SCAN = 10;

const NOTICE_COLOR = {
  warning: "text-accent-yellow",
  error: "text-accent-red",
  info: "text-accent-blue",
};

function Notice({
  tone,
  children,
}: {
  tone: keyof typeof NOTICE_COLOR;
  children: ReactNode;
}) {
  return (
    <li className="flex items-center gap-3 px-5 py-3 border-t border-border">
      <ExclamationTriangleIcon
        className={cn("w-5 h-5 shrink-0", NOTICE_COLOR[tone])}
      />
      <p className="flex-1 text-sm text-text-primary">{children}</p>
      <Button variant="ghost" size="sm" as={Link} to="/admin/license">
        License
      </Button>
    </li>
  );
}

function startNotice(license: { starts_at: number }) {
  if (license.starts_at * 1000 <= Date.now()) return null;
  return {
    tone: "warning" as const,
    message: `This license starts on ${formatLicenseTimestamp(license.starts_at)}. Devices can't be accepted until then.`,
  };
}

function capacityNotice(capacity: DeviceCapacity | null) {
  if (capacity?.state === "over" && capacity.used > capacity.limit) {
    return {
      tone: "error" as const,
      message: `${formatCount(capacity.used)} devices are accepted against a limit of ${formatCount(capacity.limit)}. Connections to devices are refused until the count is back within the license.`,
    };
  }
  if (capacity?.state === "over") {
    return {
      tone: "error" as const,
      message: `All ${formatCount(capacity.limit)} licensed devices are in use. New devices can't be accepted.`,
    };
  }
  if (capacity?.state === "approaching") {
    return {
      tone: "warning" as const,
      message: `${formatCount(capacity.used)} of ${formatCount(capacity.limit)} licensed devices are in use.`,
    };
  }
  return null;
}

function MemberRequest({
  user,
  namespaces,
  onApprove,
  onReject,
}: {
  user: UserAdminResponse;
  namespaces: string[];
  onApprove: () => void;
  onReject: () => void;
}) {
  const details = [
    user.email,
    namespaces.length > 0 && `joining ${namespaces.join(", ")}`,
    formatRelative(user.created_at),
  ].filter(Boolean);

  return (
    <li className="flex items-center gap-3 px-5 py-3 border-t border-border">
      <span
        aria-hidden="true"
        className="w-8 h-8 rounded-full bg-primary/10 border border-primary/20 text-primary text-xs font-semibold flex items-center justify-center shrink-0 uppercase"
      >
        {(user.name || user.email).charAt(0)}
      </span>
      <div className="flex-1 min-w-0">
        <Link
          to={`/admin/users/${user.id}`}
          className="block text-sm font-medium text-text-primary hover:text-primary truncate"
        >
          {user.name || user.username}
        </Link>
        <p className="text-xs text-text-muted truncate">
          {details.join(" · ")}
        </p>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        <Button
          variant="successSoft"
          size="sm"
          onClick={onApprove}
          aria-label={`Approve account for ${user.email}`}
        >
          Approve
        </Button>
        <Button
          variant="ghost"
          size="sm"
          onClick={onReject}
          aria-label={`Reject account for ${user.email}`}
        >
          Reject
        </Button>
      </div>
    </li>
  );
}

function MemberRequests({
  users,
  total,
  onApprove,
  onReject,
}: {
  users: UserAdminResponse[];
  total: number;
  onApprove: (user: UserAdminResponse) => void;
  onReject: (user: UserAdminResponse) => void;
}) {
  const joiningByUser = useAwaitingMemberNamespaces(NAMESPACE_SCAN);
  const headingId = useId();

  return (
    <div className="border-t border-border">
      <div className="flex items-center gap-2 px-5 pt-4 pb-2">
        <UserPlusIcon className="w-4 h-4 text-text-muted" />
        <h3
          id={headingId}
          className="text-xs font-semibold text-text-secondary"
        >
          New members awaiting approval
        </h3>
        <span className="text-xs font-mono text-text-muted">
          {formatCount(total)}
        </span>
      </div>
      <ul aria-labelledby={headingId}>
        {users.map((user) => (
          <MemberRequest
            key={user.id}
            user={user}
            namespaces={joiningByUser.get(user.id) ?? []}
            onApprove={() => onApprove(user)}
            onReject={() => onReject(user)}
          />
        ))}
      </ul>
    </div>
  );
}

/**
 * What waits on an Enterprise instance admin: members a namespace owner added who need approval,
 * and a licence that is expiring, not started yet or out of devices. It says nothing needs
 * attention only once every source has answered, and says which source failed otherwise.
 */
export default function NeedsAttention() {
  const pending = useAdminUsers({
    perPage: PENDING_PREVIEW,
    subset: "awaiting_approval",
  });
  const devices = useDeviceCapacity();
  const [approveTarget, setApproveTarget] = useState<UserAdminResponse | null>(
    null,
  );
  const [rejectTarget, setRejectTarget] = useState<UserAdminResponse | null>(
    null,
  );

  const licenseAlert = devices.license
    ? getLicenseAlertConfig(devices.license)
    : null;
  const notices = [
    licenseAlert && {
      tone: licenseAlert.variant,
      message: licenseAlert.message,
    },
    devices.license && startNotice(devices.license),
    capacityNotice(devices.capacity),
  ].filter((notice) => !!notice);
  const count = pending.totalCount + notices.length;
  const settled =
    !pending.isLoading &&
    !pending.isError &&
    !devices.isLoading &&
    !devices.isError;

  return (
    <Panel
      title="Needs your attention"
      action={
        count > 0 && (
          <Badge color="yellow" shape="pill">
            {formatCount(count)}
          </Badge>
        )
      }
    >
      {pending.isError && (
        <PanelError>Couldn&apos;t load member requests.</PanelError>
      )}
      {devices.isError && (
        <PanelError>
          Couldn&apos;t check the license and its device limit.
        </PanelError>
      )}

      {settled && count === 0 && (
        <p className="flex items-center gap-2 px-5 py-4 border-t border-border text-sm text-text-muted">
          <CheckCircleIcon className="w-5 h-5 text-accent-green" />
          Nothing needs your attention.
        </p>
      )}

      {notices.length > 0 && (
        <ul aria-label="License notices">
          {notices.map((notice) => (
            <Notice key={notice.message} tone={notice.tone}>
              {notice.message}
            </Notice>
          ))}
        </ul>
      )}

      {pending.users.length > 0 && (
        <MemberRequests
          users={pending.users}
          total={pending.totalCount}
          onApprove={setApproveTarget}
          onReject={setRejectTarget}
        />
      )}

      {pending.totalCount > 0 && (
        <Link
          to="/admin/users?subset=awaiting_approval"
          className="flex items-center justify-center gap-1.5 px-5 py-3 border-t border-border text-xs font-medium text-text-secondary hover:text-primary"
        >
          {pending.totalCount > pending.users.length
            ? `View all ${formatCount(pending.totalCount)} member requests`
            : "Open in the users list"}
          <ArrowRightIcon className="w-3.5 h-3.5" />
        </Link>
      )}

      <ApproveAccountDialog
        open={!!approveTarget}
        onClose={() => setApproveTarget(null)}
        user={approveTarget}
      />
      <RejectAccountDialog
        open={!!rejectTarget}
        onClose={() => setRejectTarget(null)}
        user={rejectTarget}
      />
    </Panel>
  );
}
