import { useState, useLayoutEffect } from "react";
import { useForm, useController } from "react-hook-form";
import { Link } from "react-router-dom";
import {
  CheckIcon,
  ChatBubbleBottomCenterTextIcon,
  PencilSquareIcon,
  ChevronDownIcon,
  Cog6ToothIcon,
  PencilIcon,
  TagIcon,
  FingerPrintIcon,
  VideoCameraIcon,
  TrashIcon,
  ArrowRightStartOnRectangleIcon,
  DevicePhoneMobileIcon,
} from "@heroicons/react/24/outline";
import { useNamespace } from "../hooks/useNamespaces";
import {
  useEditNamespace,
  useDeleteNamespace,
  useLeaveNamespace,
  useSetDeviceAutoAccept,
} from "../hooks/useNamespaceMutations";
import { useAuthStore } from "../stores/authStore";
import { useHasPermission } from "../hooks/useHasPermission";
import PageHeader from "../components/common/PageHeader";
import CopyButton from "../components/common/CopyButton";
import Drawer from "../components/common/Drawer";
import ConfirmDialog from "../components/common/ConfirmDialog";
import BillingSection from "../components/billing/BillingSection";
import InputField from "@/components/common/fields/InputField";
import NamespaceNameField from "@/components/common/fields/NamespaceNameField";
import { validateNamespaceName } from "@/utils/validation";
import { getConfig } from "../env";
import { Button, IconButton } from "@shellhub/design-system/primitives";
import PageLoader from "@/components/common/PageLoader";
import SettingsCard from "@/components/common/SettingsCard";
import SettingsRow from "@/components/common/SettingsRow";

/* ─── Edit Name Drawer ─── */

type EditNameFormValues = { name: string };

function EditNameDrawer({
  open,
  onClose,
  currentName,
  tenantId,
}: {
  open: boolean;
  onClose: () => void;
  currentName: string;
  tenantId: string;
}) {
  const editNs = useEditNamespace();

  const { control, handleSubmit, reset, setError, clearErrors, formState } =
    useForm<EditNameFormValues>({
      mode: "onChange",
      defaultValues: { name: currentName },
    });

  const { isDirty, isValid, isSubmitting } = formState;

  useLayoutEffect(() => {
    reset({ name: currentName });
  }, [open, currentName, reset]);

  const { field, fieldState } = useController({
    name: "name",
    control,
    rules: { validate: (v) => validateNamespaceName(v) ?? true },
  });

  const onValid = async (values: EditNameFormValues) => {
    clearErrors("root");
    try {
      await editNs.mutateAsync({
        path: { tenant: tenantId },
        body: { name: values.name },
      });
      onClose();
    } catch {
      setError("root", {
        message: "Failed to rename namespace. The name may already be taken.",
      });
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title="Rename Namespace"
      bodyClassName="flex-1 overflow-y-auto px-6 py-5"
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            onClick={() => void handleSubmit(onValid)()}
            disabled={!isDirty || !isValid || isSubmitting}
            loading={isSubmitting}
            icon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Save
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(onValid)(e)} className="space-y-5">
        <NamespaceNameField
          id="edit-ns-name"
          value={field.value}
          onChange={field.onChange}
          error={fieldState.error?.message}
        />
        {formState.errors.root && (
          <p role="alert" className="text-2xs text-accent-red">
            {formState.errors.root.message}
          </p>
        )}
      </form>
    </Drawer>
  );
}

/* ─── Delete Namespace Dialog ─── */

function DeleteDialog({
  namespaceName,
  tenantId,
  onClose,
}: {
  namespaceName: string;
  tenantId: string;
  onClose: () => void;
}) {
  const deleteNs = useDeleteNamespace();
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");

  const canDelete = confirm === namespaceName;

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setError("");
        try {
          await deleteNs.mutateAsync(tenantId);
        } catch {
          setError("Failed to delete namespace.");
          throw new Error();
        }
      }}
      title="Delete Namespace"
      description={
        <>
          This action is{" "}
          <span className="font-medium text-accent-red">permanent</span> and
          cannot be undone. All devices, sessions, and data will be lost.
        </>
      }
      confirmLabel="Delete Namespace"
      confirmDisabled={!canDelete}
    >
      <div className="mb-4">
        <InputField
          id="delete-ns-confirm"
          label={`Type "${namespaceName}" to confirm`}
          value={confirm}
          onChange={setConfirm}
          placeholder={namespaceName}
        />
      </div>
      {error && <p className="text-2xs text-accent-red mb-3">{error}</p>}
    </ConfirmDialog>
  );
}

/* ─── Leave Namespace Dialog ─── */

function LeaveDialog({
  tenantId,
  onClose,
}: {
  tenantId: string;
  onClose: () => void;
}) {
  const leaveNs = useLeaveNamespace();
  const [error, setError] = useState("");

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={async () => {
        setError("");
        try {
          await leaveNs.mutateAsync(tenantId);
        } catch {
          setError("Failed to leave namespace.");
          throw new Error();
        }
      }}
      title="Leave Namespace"
      description="You will lose access to all devices and sessions. To rejoin, someone will need to invite you again."
      confirmLabel="Leave"
    >
      {error && <p className="text-2xs text-accent-red mb-3">{error}</p>}
    </ConfirmDialog>
  );
}

/* ─── Banner Collapsible ─── */

function BannerPreview({
  banner,
  canEdit,
}: {
  banner: string;
  canEdit: boolean;
}) {
  const [open, setOpen] = useState(false);

  if (!banner) {
    return (
      <SettingsRow
        icon={<ChatBubbleBottomCenterTextIcon className="w-4 h-4" />}
        title="SSH Banner"
        description="Message shown when users connect via SSH"
      >
        {canEdit && (
          <IconButton
            as={Link}
            to="/settings/banner"
            title="Edit"
            aria-label="Edit"
          >
            <PencilSquareIcon className="w-4 h-4" />
          </IconButton>
        )}
      </SettingsRow>
    );
  }

  return (
    <div className="px-5 py-4">
      {/* Header */}
      <div className="flex items-center justify-between gap-4">
        <button
          type="button"
          onClick={() => setOpen((v) => !v)}
          className="flex items-start gap-3 min-w-0 flex-1 text-left"
        >
          <span className="w-8 h-8 rounded-lg bg-hover-medium border border-border flex items-center justify-center text-text-muted shrink-0 mt-0.5">
            <ChatBubbleBottomCenterTextIcon className="w-4 h-4" />
          </span>
          <div className="min-w-0 pt-0.5">
            <p className="text-sm font-medium text-text-primary">SSH Banner</p>
            <p className="text-2xs text-text-muted mt-0.5">
              Message shown when users connect via SSH
            </p>
          </div>
        </button>
        <div className="flex items-center gap-1 shrink-0">
          {canEdit && (
            <IconButton
              as={Link}
              to="/settings/banner"
              title="Edit"
              aria-label="Edit"
            >
              <PencilSquareIcon className="w-4 h-4" />
            </IconButton>
          )}
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            className="inline-flex p-1.5 rounded-md text-text-muted hover:text-text-primary hover:bg-hover-medium transition-colors"
          >
            <ChevronDownIcon
              className={`w-4 h-4 transition-transform duration-200 ${open ? "rotate-180" : ""}`}
            />
          </button>
        </div>
      </div>

      {/* Collapsible content */}
      <div className="ml-11 mt-3">
        <div
          className={`relative overflow-hidden rounded-lg border border-border bg-card transition-all duration-200 ease-out ${open ? "max-h-[500px]" : "max-h-[120px]"}`}
        >
          <pre className="px-3 py-2.5 text-xs font-mono text-text-secondary leading-relaxed whitespace-pre-wrap break-words">
            {banner}
          </pre>
          {!open && (
            <>
              <div className="absolute inset-x-0 bottom-0 h-16 bg-gradient-to-t from-card via-card/70 to-transparent pointer-events-none" />
              <button
                type="button"
                onClick={() => setOpen(true)}
                className="absolute inset-x-0 bottom-0 flex items-center justify-center pb-2.5 pt-4"
              >
                <span className="inline-flex items-center gap-1 px-3 py-1 text-2xs font-semibold text-text-primary bg-surface border border-border-light rounded-full shadow-sm hover:bg-card hover:border-primary/30 transition-all">
                  Show more
                  <ChevronDownIcon
                    className="w-3 h-3 animate-bounce-subtle"
                    strokeWidth={2}
                  />
                </span>
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

/* ─── Page ─── */

export default function Settings() {
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");
  const editNs = useEditNamespace();
  const setDeviceAutoAccept = useSetDeviceAutoAccept();
  const [editNameOpen, setEditNameOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [leaveOpen, setLeaveOpen] = useState(false);
  const [togglingRecord, setTogglingRecord] = useState(false);
  const [togglingAutoAccept, setTogglingAutoAccept] = useState(false);

  const canRename = useHasPermission("namespace:rename");
  const canUpdateRecording = useHasPermission(
    "namespace:updateSessionRecording",
  );
  const canUpdateAutoAccept = useHasPermission(
    "namespace:updateDeviceAutoAccept",
  );
  const canEditBanner = useHasPermission("namespace:editBanner");
  const canDelete = useHasPermission("namespace:delete");

  const settings = ns?.settings;
  const sessionRecord = settings?.session_record ?? false;
  const deviceAutoAccept = settings?.device_auto_accept ?? false;
  const banner = settings?.connection_announcement ?? "";

  const handleToggleRecord = async () => {
    if (!tenantId || togglingRecord) return;
    setTogglingRecord(true);
    try {
      await editNs.mutateAsync({
        path: { tenant: tenantId },
        body: {
          settings: {
            session_record: !sessionRecord,
            connection_announcement: banner,
            device_auto_accept: deviceAutoAccept,
          },
        },
      });
    } catch {
      /* state didn't change */
    } finally {
      setTogglingRecord(false);
    }
  };

  const handleToggleAutoAccept = async () => {
    if (!tenantId || togglingAutoAccept) return;
    setTogglingAutoAccept(true);
    try {
      await setDeviceAutoAccept.mutateAsync({
        path: { tenant: tenantId },
        body: { device_auto_accept: !deviceAutoAccept },
      });
    } catch {
      /* state didn't change */
    } finally {
      setTogglingAutoAccept(false);
    }
  };

  if (!ns) {
    return <PageLoader label="Loading settings" padding="lg" />;
  }

  return (
    <div>
      <PageHeader
        icon={<Cog6ToothIcon className="w-6 h-6" />}
        overline="Namespace"
        title="Settings"
        description="Manage namespace configuration, SSH options, and access"
      />

      <div className="space-y-6 animate-fade-in">
        {/* ── General ── */}
        <SettingsCard title="General">
          <SettingsRow
            icon={<PencilIcon className="w-4 h-4" />}
            title="Name"
            description="The display name for this namespace"
          >
            <div className="flex items-center gap-2">
              <span className="text-sm font-mono text-text-secondary">
                {ns.name}
              </span>
              {canRename && (
                <IconButton
                  variant="primary"
                  type="button"
                  title="Rename"
                  aria-label="Rename namespace"
                  onClick={() => setEditNameOpen(true)}
                >
                  <PencilSquareIcon className="w-4 h-4" />
                </IconButton>
              )}
            </div>
          </SettingsRow>

          <SettingsRow
            icon={<TagIcon className="w-4 h-4" />}
            title="Type"
            description="Defines whether this namespace belongs to one user or a team"
          >
            <span
              className={`inline-flex items-center px-2.5 py-1 text-2xs font-mono font-semibold rounded border ${
                ns.type === "team"
                  ? "bg-primary/10 text-primary border-primary/20"
                  : "bg-accent-yellow/10 text-accent-yellow border-accent-yellow/20"
              }`}
            >
              {ns.type ?? "personal"}
            </span>
          </SettingsRow>

          <SettingsRow
            icon={<FingerPrintIcon className="w-4 h-4" />}
            title="Tenant ID"
            description="Use this identifier when integrating with the ShellHub API"
          >
            <div className="flex items-center gap-1">
              <code className="text-xs font-mono text-accent-cyan select-all">
                {ns.tenant_id}
              </code>
              <CopyButton text={ns.tenant_id} size="md" />
            </div>
          </SettingsRow>
        </SettingsCard>

        {/* ── SSH ── */}
        <SettingsCard title="SSH">
          {/* Session Recording (Cloud/Enterprise only) */}
          {(getConfig().cloud || getConfig().enterprise) && (
            <SettingsRow
              icon={<VideoCameraIcon className="w-4 h-4" />}
              title="Session Recording"
              description="Record SSH sessions for audit and playback"
            >
              <div
                className={`inline-flex items-center h-7 bg-card border border-border rounded-md p-0.5 ${!canUpdateRecording || togglingRecord ? "opacity-40 pointer-events-none" : ""}`}
              >
                <button
                  type="button"
                  onClick={() => {
                    if (sessionRecord) void handleToggleRecord();
                  }}
                  className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                    !sessionRecord
                      ? "bg-hover-strong text-text-secondary border border-border-light"
                      : "text-text-muted hover:text-text-secondary border border-transparent"
                  }`}
                >
                  Off
                </button>
                <button
                  type="button"
                  onClick={() => {
                    if (!sessionRecord) void handleToggleRecord();
                  }}
                  className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                    sessionRecord
                      ? "bg-primary/15 text-primary border border-primary/25"
                      : "text-text-muted hover:text-text-secondary border border-transparent"
                  }`}
                >
                  On
                </button>
              </div>
            </SettingsRow>
          )}

          {/* SSH Banner */}
          <BannerPreview banner={banner} canEdit={canEditBanner} />
        </SettingsCard>

        {/* ── Devices ── */}
        <SettingsCard title="Devices">
          <SettingsRow
            icon={<DevicePhoneMobileIcon className="w-4 h-4" />}
            title="Auto-Accept Devices"
            description="Automatically accept new devices when they connect for the first time"
          >
            <div
              className={`inline-flex items-center h-7 bg-card border border-border rounded-md p-0.5 ${!canUpdateAutoAccept || togglingAutoAccept ? "opacity-40 pointer-events-none" : ""}`}
            >
              <button
                type="button"
                onClick={() => {
                  if (deviceAutoAccept) void handleToggleAutoAccept();
                }}
                className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                  !deviceAutoAccept
                    ? "bg-hover-strong text-text-secondary border border-border-light"
                    : "text-text-muted hover:text-text-secondary border border-transparent"
                }`}
              >
                Off
              </button>
              <button
                type="button"
                onClick={() => {
                  if (!deviceAutoAccept) void handleToggleAutoAccept();
                }}
                className={`h-full px-2.5 text-2xs font-medium rounded transition-all duration-150 ${
                  deviceAutoAccept
                    ? "bg-primary/15 text-primary border border-primary/25"
                    : "text-text-muted hover:text-text-secondary border border-transparent"
                }`}
              >
                On
              </button>
            </div>
          </SettingsRow>
        </SettingsCard>

        {/* ── Billing (Cloud only) ── */}
        {getConfig().cloud && <BillingSection sectionId="billing" />}

        {/* ── Danger Zone ── */}
        <SettingsCard title="Danger Zone" danger>
          {canDelete ? (
            <SettingsRow
              icon={<TrashIcon className="w-4 h-4 text-accent-red" />}
              title="Delete Namespace"
              description="Permanently removes all devices, sessions, keys, and configuration. This cannot be undone."
            >
              <Button
                size="sm"
                variant="dangerSoft"
                onClick={() => setDeleteOpen(true)}
              >
                Delete
              </Button>
            </SettingsRow>
          ) : (
            <SettingsRow
              icon={
                <ArrowRightStartOnRectangleIcon className="w-4 h-4 text-accent-red" />
              }
              title="Leave Namespace"
              description="You will lose access immediately. To rejoin, someone will need to invite you again."
            >
              <Button
                size="sm"
                variant="dangerSoft"
                onClick={() => setLeaveOpen(true)}
              >
                Leave
              </Button>
            </SettingsRow>
          )}
        </SettingsCard>
      </div>

      {/* Drawers & Dialogs */}
      <EditNameDrawer
        open={editNameOpen}
        onClose={() => setEditNameOpen(false)}
        currentName={ns.name}
        tenantId={ns.tenant_id}
      />
      {deleteOpen && (
        <DeleteDialog
          namespaceName={ns.name}
          tenantId={ns.tenant_id}
          onClose={() => setDeleteOpen(false)}
        />
      )}
      {leaveOpen && (
        <LeaveDialog
          tenantId={ns.tenant_id}
          onClose={() => setLeaveOpen(false)}
        />
      )}
    </div>
  );
}
