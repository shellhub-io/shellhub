import { useMutation } from "@tanstack/react-query";
import { useState } from "react";
import { useController } from "react-hook-form";
import { Link } from "react-router-dom";
import {
  CheckIcon,
  CommandLineIcon,
  KeyIcon,
  ExclamationTriangleIcon,
  LockClosedIcon,
  LockOpenIcon,
  ShieldCheckIcon,
  VideoCameraIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import {
  type Namespace,
  useEditNamespace,
  useNamespace,
} from "@/hooks/useNamespaceRequests";
import { isLegacyNamespace } from "@/hooks/useNamespaceLegacy";
import { useAccessPolicies } from "@/hooks/useAccessPolicyRequests";
import { useUnverifiedDeviceCount } from "@/hooks/useDevices";
import { formatCount } from "@/pages/provisioning-keys/helpers";
import { useHasPermission } from "@/hooks/useHasPermission";
import { useSettingSave } from "@/hooks/useSettingSave";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import { useAuthStore } from "@/stores/authStore";
import {
  ansiColor,
  useTerminalThemeStore,
  type AnsiColorName,
} from "@/stores/terminalThemeStore";
import { isEnterpriseOrCloud } from "@/env";
import { readableOn } from "@/utils/color";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import InputField from "@/components/common/fields/InputField";
import FormModal from "@/components/common/FormModal";
import Modal from "@/components/common/Modal";
import PageLoader from "@/components/common/PageLoader";
import RecBadge from "@/components/sessions/RecBadge";
import StatePill from "@/components/common/StatePill";
import SettingsSection from "@/components/settings/SettingsSection";
import SavedMark from "@/components/settings/SavedMark";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import { upgradeNamespaceMutation } from "@/client";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import {
  BANNER_MAX_LENGTH,
  bannerSchema,
  type BannerFormValues,
} from "./bannerSchema";

const MODELS = {
  new: {
    icon: ShieldCheckIcon,
    name: "New",
    how: "Each login is approved in the browser by whoever connects, and an access policy has to allow it.",
  },
  legacy: {
    icon: KeyIcon,
    name: "Legacy",
    how: "Logins are allowed by the public keys added to this namespace and their filters. There is no browser approval.",
  },
} as const;

const UPGRADE_CHANGES = [
  "The owner keeps SSH access through a starter access policy, unless the namespace already has policies, which stay as they are. Everyone else needs a policy that allows them.",
  "Password and public-key logins stop authenticating, in the web terminal too.",
  "Each person approves an SSH identity on their first login.",
  "New devices enroll by pairing or a provisioning key. Devices already enrolled keep working.",
  "There is no way back to legacy access.",
];

function NoPoliciesWarning() {
  const { policies, isLoading } = useAccessPolicies();
  if (isLoading || policies.length > 0) return null;
  return (
    <p className="mt-2 flex items-center gap-1.5 text-xs text-accent-yellow">
      <ExclamationTriangleIcon
        aria-hidden="true"
        className="w-3.5 h-3.5 shrink-0"
        strokeWidth={2}
      />
      No access policies yet, so every SSH login is denied.
      <Link to="/access-policies" className="underline hover:no-underline">
        Add one
      </Link>
    </p>
  );
}

function UpgradeDialog({
  ns,
  onClose,
  onConfirm,
}: {
  ns: Namespace;
  onClose: () => void;
  onConfirm: () => Promise<void>;
}) {
  const [confirm, setConfirm] = useState("");

  return (
    <ConfirmDialog
      open
      onClose={onClose}
      onConfirm={onConfirm}
      icon={<ShieldCheckIcon />}
      variant="warning"
      title="Upgrade this namespace?"
      description="This changes how everyone reaches this namespace's devices."
      confirmLabel="Upgrade namespace"
      confirmDisabled={confirm !== ns.name}
    >
      <ul className="mb-4 space-y-1.5 pl-4 text-xs text-text-secondary list-disc">
        {UPGRADE_CHANGES.map((change) => (
          <li key={change}>{change}</li>
        ))}
      </ul>
      <div className="mb-4">
        <InputField
          id="upgrade-ns-confirm"
          label={`Type "${ns.name}" to confirm`}
          value={confirm}
          onChange={setConfirm}
          placeholder={ns.name}
        />
      </div>
    </ConfirmDialog>
  );
}

function NamespaceModelCard({ ns }: { ns: Namespace }) {
  const invalidateUpgraded = useInvalidateByIds(
    "getNamespaces",
    "getNamespace",
    "listAccessPolicies",
  );
  const upgrade = useMutation({
    ...upgradeNamespaceMutation(),
    onSuccess: invalidateUpgraded,
  });
  const canUpgrade = useHasPermission("namespace:upgrade");
  const { error, saved, run } = useSettingSave(
    "Couldn't upgrade the namespace. Try again.",
  );
  const [open, setOpen] = useState(false);

  const legacy = isLegacyNamespace(ns);
  const { icon: Icon, name, how } = legacy ? MODELS.legacy : MODELS.new;

  return (
    <>
      <div role="group" aria-label="Namespace">
        <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 px-5 py-4 rounded-xl border border-border bg-card">
          <div className="flex gap-3.5 min-w-0 flex-1 basis-72">
            <span
              aria-hidden="true"
              className={cn(
                "grid place-items-center w-9 h-9 shrink-0 rounded-lg border",
                legacy
                  ? "bg-hover-medium border-border-light text-text-secondary"
                  : "bg-primary/10 border-primary/20 text-primary",
              )}
            >
              <Icon className="w-[18px] h-[18px]" />
            </span>
            <div className="min-w-0">
              <p className="text-2xs font-mono uppercase tracking-label text-text-muted">
                Namespace
              </p>
              <p className="mt-0.5 flex items-center gap-2 text-sm font-semibold text-text-primary">
                {name}
                {saved && <SavedMark />}
              </p>
              <p className="mt-1 max-w-md text-xs text-text-muted leading-relaxed">
                {how}
              </p>
              {!legacy && <NoPoliciesWarning />}
            </div>
          </div>
          {legacy && canUpgrade && (
            <Button
              size="sm"
              variant="primary"
              disabled={upgrade.isPending}
              onClick={() => setOpen(true)}
            >
              Upgrade
            </Button>
          )}
        </div>
        {error && (
          <p role="alert" className="mt-2 text-2xs text-accent-red">
            {error}
          </p>
        )}
      </div>

      {open && (
        <UpgradeDialog
          ns={ns}
          onClose={() => setOpen(false)}
          onConfirm={async () => {
            setOpen(false);
            await run(() =>
              upgrade.mutateAsync({ path: { tenant: ns.tenant_id } }),
            );
          }}
        />
      )}
    </>
  );
}

function RecordingCard({ ns }: { ns: Namespace }) {
  const editNs = useEditNamespace();
  const canChange = useHasPermission("namespace:updateSessionRecording");
  const { error, saved, run } = useSettingSave(
    "Couldn't change session recording. Try again.",
  );
  const recording = ns.settings?.session_record ?? false;

  return (
    <SettingsSwitchCard
      icon={<VideoCameraIcon />}
      title="Session recording"
      description="Record every SSH session on the server, for audit and playback."
      error={error}
      saved={saved}
      control={
        <RecBadge
          on={recording}
          label="Session recording"
          disabled={!canChange || editNs.isPending}
          onToggle={() =>
            void run(() =>
              editNs.mutateAsync({
                path: { tenant: ns.tenant_id },
                body: { settings: { session_record: !recording } },
              }),
            )
          }
        />
      }
    />
  );
}

function UnsignedWarning({
  unsigned,
  total,
}: {
  unsigned: number;
  total: number;
}) {
  return (
    <p className="flex items-center gap-2 px-3 py-2 rounded-lg border border-accent-yellow/20 bg-accent-yellow/8 text-xs text-accent-yellow">
      <ExclamationTriangleIcon
        aria-hidden="true"
        className="w-4 h-4 shrink-0"
        strokeWidth={2}
      />
      {unsigned === 1
        ? `1 of your ${formatCount(total)} devices doesn't support signatures yet. Update it before you enforce them here.`
        : `${formatCount(unsigned)} of your ${formatCount(total)} devices don't support signatures yet. Update them before you enforce them here.`}
    </p>
  );
}

function EnforcedBadge({
  enforced,
  disabled,
  onChange,
}: {
  enforced: boolean;
  disabled: boolean;
  onChange: (enforced: boolean) => void;
}) {
  const Icon = enforced ? LockClosedIcon : LockOpenIcon;

  return (
    <StatePill
      on={enforced}
      onClassName="border-accent-green/40 bg-accent-green/10 text-accent-green"
      onToggle={() => onChange(!enforced)}
      label="Device signature enforcement"
      disabled={disabled}
    >
      <Icon aria-hidden="true" className="w-3.5 h-3.5" strokeWidth={2} />
      <span className="font-mono text-[10px] font-bold tracking-wider">
        {enforced ? "ON" : "OFF"}
      </span>
    </StatePill>
  );
}

function EnforceSignaturesCard({ ns }: { ns: Namespace }) {
  const editNs = useEditNamespace();
  const canChange = useHasPermission("namespace:requireVerifiedDeviceKeys");
  const { error, saved, run } = useSettingSave(
    "Couldn't change device signature enforcement. Try again.",
  );
  const [confirming, setConfirming] = useState(false);
  const enforced = ns.settings?.require_verified_device_key ?? false;
  const unverified = useUnverifiedDeviceCount({ enabled: !enforced });
  const unsigned = unverified.count ?? 0;
  const total = Math.max(
    (ns.devices_accepted_count ?? 0) + (ns.devices_pending_count ?? 0),
    unsigned,
  );
  const counted = !unverified.isLoading && !unverified.error;
  const nobodyCutOff = counted && unsigned === 0;

  const save = (next: boolean) =>
    run(() =>
      editNs.mutateAsync({
        path: { tenant: ns.tenant_id },
        body: { settings: { require_verified_device_key: next } },
      }),
    );

  const disabled =
    !canChange || editNs.isPending || (!enforced && unverified.isLoading);
  const change = (next: boolean) => {
    if (disabled) return;
    if (!next || nobodyCutOff) void save(next);
    else setConfirming(true);
  };

  return (
    <>
      <SettingsSwitchCard
        icon={<LockClosedIcon />}
        title="Device signature enforcement"
        description="Refuse agents that can't prove they hold the device's key."
        error={error}
        saved={saved}
        control={
          <EnforcedBadge
            enforced={enforced}
            disabled={disabled}
            onChange={change}
          />
        }
        footer={
          !enforced &&
          counted &&
          unsigned > 0 && <UnsignedWarning unsigned={unsigned} total={total} />
        }
      />
      <ConfirmDialog
        open={confirming}
        onClose={() => setConfirming(false)}
        onConfirm={async () => {
          setConfirming(false);
          await save(true);
        }}
        icon={<LockClosedIcon />}
        variant="warning"
        title="Enforce device signatures?"
        description={
          counted
            ? `${formatCount(unsigned)} of ${formatCount(total)} devices will stop connecting.`
            : "Devices whose agent can't sign will stop connecting."
        }
        confirmLabel="Enforce anyway"
      >
        <p className="mb-4 text-xs text-text-secondary leading-relaxed">
          {unsigned === 1
            ? "ShellHub will refuse its agent until it's updated, and you won't be able to reach it in the meantime."
            : "ShellHub will refuse their agents until they're updated, and you won't be able to reach them in the meantime."}
        </p>
      </ConfirmDialog>
    </>
  );
}

function useTerminalLook() {
  const colors = useTerminalThemeStore((st) => st.theme.colors);
  const fontFamily = useTerminalThemeStore((st) => st.fontFamilyWithFallback);
  const fontSize = useTerminalThemeStore((st) => st.fontSize);
  return {
    style: {
      background: colors.background,
      color: colors.foreground,
      fontFamily,
      fontSize,
    },
    ansi: (name: AnsiColorName) => ansiColor(colors, name),
  };
}

function LoginPrompt({ ns }: { ns: Namespace }) {
  const { style, ansi } = useTerminalLook();
  const on = (candidates: string[]) =>
    readableOn(style.background, candidates, style.color);
  return (
    <>
      <span
        style={{ color: on([ansi("green"), ansi("brightGreen")]) }}
        className="font-bold"
      >
        $
      </span>{" "}
      ssh{" "}
      <span
        style={{
          color: on([
            ansi("cyan"),
            ansi("brightCyan"),
            ansi("blue"),
            ansi("brightBlue"),
          ]),
        }}
      >
        {`root@${ns.name}.device@${window.location.hostname}`}
      </span>
    </>
  );
}

function BannerModal({
  open,
  onClose,
  ns,
}: {
  open: boolean;
  onClose: () => void;
  ns: Namespace;
}) {
  const editNs = useEditNamespace();
  const current = ns.settings?.connection_announcement ?? "";
  const form = useDrawerForm(open, bannerSchema, { banner: current });
  const {
    control,
    setError,
    clearErrors,
    formState: { errors },
  } = form;
  const { field } = useController({ name: "banner", control });
  const { style } = useTerminalLook();
  const overLimit = field.value.length > BANNER_MAX_LENGTH;

  const onValid = async (values: BannerFormValues) => {
    clearErrors("root");
    try {
      await editNs.mutateAsync({
        path: { tenant: ns.tenant_id },
        body: { settings: { connection_announcement: values.banner } },
      });
      onClose();
    } catch {
      setError("root", { message: "Couldn't save the banner. Try again." });
    }
  };

  return (
    <FormModal
      size="lg"
      form={form}
      onSubmit={onValid}
      open={open}
      onClose={onClose}
      icon={<CommandLineIcon />}
      title="Login banner"
      description="Plain text shown to anyone who connects to a device in this namespace over SSH."
      submitLabel="Save banner"
      requireDirty
      bodyClassName="p-0"
      footerStart={
        <span className="flex flex-wrap items-center gap-x-3 gap-y-1">
          <span
            className={cn(
              "text-2xs font-mono",
              overLimit ? "text-accent-red font-semibold" : "text-text-muted",
            )}
          >
            {field.value.length.toLocaleString()}/
            {BANNER_MAX_LENGTH.toLocaleString()}
          </span>
          {errors.banner?.message && (
            <span role="alert" className="text-2xs text-accent-red">
              {errors.banner.message}
            </span>
          )}
        </span>
      }
      submitIcon={<CheckIcon className="w-4 h-4" strokeWidth={2} />}
    >
      <div style={{ background: style.background }}>
        <pre
          style={style}
          className="px-6 pt-4 leading-relaxed whitespace-pre-wrap break-all"
        >
          <LoginPrompt ns={ns} />
        </pre>
        <textarea
          {...field}
          data-autofocus
          aria-label="Login banner"
          rows={14}
          placeholder="Authorized use only."
          style={style}
          className="block w-full px-6 pt-1 pb-4 placeholder:opacity-40 focus:outline-none resize-none leading-relaxed"
        />
      </div>
    </FormModal>
  );
}

function BannerViewer({
  open,
  onClose,
  ns,
}: {
  open: boolean;
  onClose: () => void;
  ns: Namespace;
}) {
  const current = ns.settings?.connection_announcement ?? "";
  const { style } = useTerminalLook();

  return (
    <Modal
      size="lg"
      open={open}
      onClose={onClose}
      icon={<CommandLineIcon />}
      title="Login banner"
      description="What anyone who connects to a device in this namespace over SSH sees first."
      bodyClassName="p-0"
      footer={
        <Button variant="ghost" onClick={onClose}>
          Close
        </Button>
      }
    >
      <pre
        style={style}
        className="px-6 py-4 leading-relaxed whitespace-pre-wrap break-words"
      >
        <LoginPrompt ns={ns} />
        {"\n"}
        {current || (
          <span className="italic opacity-60">No banner is set.</span>
        )}
      </pre>
    </Modal>
  );
}

function BannerCard({ ns }: { ns: Namespace }) {
  const canEdit = useHasPermission("namespace:editBanner");
  const [open, setOpen] = useState(false);

  return (
    <>
      <SettingsSwitchCard
        icon={<CommandLineIcon />}
        title="Login banner"
        description="Shown to anyone who connects over SSH, such as a legal notice."
        control={
          <Button size="sm" variant="secondary" onClick={() => setOpen(true)}>
            {canEdit ? "Edit" : "View"}
          </Button>
        }
      />
      {canEdit ? (
        <BannerModal open={open} onClose={() => setOpen(false)} ns={ns} />
      ) : (
        <BannerViewer open={open} onClose={() => setOpen(false)} ns={ns} />
      )}
    </>
  );
}

/**
 * The SSH settings: how logins are authorized, whether sessions are recorded on the server, whether
 * agents must sign their connections with the device key, and the banner shown at login. Recording
 * applies at once. Enforcing device signatures applies at once when no accepted or pending device
 * would stop connecting, and otherwise asks first, showing how many would; it also asks when that
 * count cannot be read. Lifting it applies at once. The upgrade of a legacy namespace asks for the
 * namespace's name after spelling out what changes, since it cannot be undone. The banner is edited
 * in its own dialog, or read in one by a role that may not edit it.
 */
export default function SshSettings() {
  const { tenant: tenantId } = useAuthStore();
  const { namespace: ns } = useNamespace(tenantId ?? "");

  if (!ns) return <PageLoader label="Loading settings" padding="lg" />;

  return (
    <SettingsSection
      title="SSH"
      description="How logins to this namespace's devices are authorized and recorded."
    >
      <NamespaceModelCard ns={ns} />
      {isEnterpriseOrCloud() && <RecordingCard ns={ns} />}
      <EnforceSignaturesCard ns={ns} />
      <BannerCard ns={ns} />
    </SettingsSection>
  );
}
