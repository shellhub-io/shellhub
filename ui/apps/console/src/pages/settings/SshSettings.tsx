import { useState } from "react";
import { useController } from "react-hook-form";
import { Link } from "react-router-dom";
import {
  CheckIcon,
  CommandLineIcon,
  KeyIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  VideoCameraIcon,
} from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useNamespace, type Namespace } from "@/hooks/useNamespaces";
import { useAccessPolicies } from "@/hooks/useAccessPolicies";
import {
  useEditNamespace,
  useSetSshAccessMode,
} from "@/hooks/useNamespaceMutations";
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
import FormModal from "@/components/common/FormModal";
import Modal from "@/components/common/Modal";
import PageLoader from "@/components/common/PageLoader";
import RecBadge from "@/components/sessions/RecBadge";
import SettingsSection from "@/components/settings/SettingsSection";
import SavedMark from "@/components/settings/SavedMark";
import SettingsSwitchCard from "@/components/settings/SettingsSwitchCard";
import {
  BANNER_MAX_LENGTH,
  bannerSchema,
  type BannerFormValues,
} from "./bannerSchema";

type AccessMode = "legacy" | "identity";

const MODES = {
  identity: {
    icon: ShieldCheckIcon,
    name: "Identity",
    how: "Each login is approved in the browser by whoever connects, and an access policy has to allow it.",
  },
  legacy: {
    icon: KeyIcon,
    name: "Legacy",
    how: "Logins are allowed by the public keys added to this namespace and their filters. There is no browser approval.",
  },
} as const;

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

function AccessModeCard({ ns }: { ns: Namespace }) {
  const setSshAccessMode = useSetSshAccessMode();
  const { policies, isLoading: policiesLoading } = useAccessPolicies();
  const canChange = useHasPermission("namespace:updateSshAccessMode");
  const { error, saved, run } = useSettingSave(
    "Couldn't change the SSH access mode. Try again.",
  );
  const [pending, setPending] = useState<AccessMode | null>(null);

  const mode: AccessMode = ns.settings?.ssh_access_mode ?? "identity";
  const other: AccessMode = mode === "identity" ? "legacy" : "identity";
  const legacyAllowed = ns.settings?.ssh_legacy_allowed ?? false;
  const noPolicies = !policiesLoading && policies.length === 0;
  const { icon: Icon, name, how } = MODES[mode];

  const apply = (next: AccessMode) =>
    run(() =>
      setSshAccessMode.mutateAsync({
        path: { tenant: ns.tenant_id },
        body: { ssh_access_mode: next },
      }),
    );

  return (
    <>
      <div role="group" aria-label="SSH access mode">
        <div className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 px-5 py-4 rounded-xl border border-border bg-card">
          <div className="flex gap-3.5 min-w-0 flex-1 basis-72">
            <span
              aria-hidden="true"
              className={cn(
                "grid place-items-center w-9 h-9 shrink-0 rounded-lg border",
                mode === "identity"
                  ? "bg-primary/10 border-primary/20 text-primary"
                  : "bg-hover-medium border-border-light text-text-secondary",
              )}
            >
              <Icon className="w-[18px] h-[18px]" />
            </span>
            <div className="min-w-0">
              <p className="text-2xs font-mono uppercase tracking-label text-text-muted">
                SSH access mode
              </p>
              <p className="mt-0.5 flex items-center gap-2 text-sm font-semibold text-text-primary">
                {name}
                {saved && <SavedMark />}
              </p>
              <p className="mt-1 max-w-md text-xs text-text-muted leading-relaxed">
                {how}
              </p>
              {mode === "identity" && <NoPoliciesWarning />}
            </div>
          </div>
          {legacyAllowed && canChange && (
            <Button
              size="sm"
              variant={other === "identity" ? "primary" : "secondary"}
              disabled={setSshAccessMode.isPending}
              onClick={() => setPending(other)}
            >
              Switch to {MODES[other].name.toLowerCase()}
            </Button>
          )}
        </div>
        {error && (
          <p role="alert" className="mt-2 text-2xs text-accent-red">
            {error}
          </p>
        )}
      </div>

      <ConfirmDialog
        open={pending !== null}
        onClose={() => setPending(null)}
        onConfirm={async () => {
          const next = pending;
          setPending(null);
          if (next) await apply(next);
        }}
        icon={<ShieldCheckIcon />}
        variant="warning"
        title={
          pending === "identity"
            ? "Switch to identity access?"
            : "Switch to legacy access?"
        }
        description={
          pending === "identity"
            ? noPolicies
              ? "Every login will need browser approval and an access policy that allows it. There are no policies yet, so every SSH login will be denied until you add one."
              : "Every login will need browser approval and an access policy that allows it. Public keys stop granting access on their own."
            : "Logins will be authorized by public keys and their filters. Access policies and browser approval stop applying."
        }
        confirmLabel={
          pending === "identity" ? "Switch to identity" : "Switch to legacy"
        }
      />
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
      <span style={{ color: on([ansi("green"), ansi("brightGreen")]) }} className="font-bold">
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
        {current || <span className="italic opacity-60">No banner is set.</span>}
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
 * The SSH settings: how logins are authorized, whether sessions are recorded on the server, and
 * the banner shown at login. Recording applies at once, the access mode after a confirmation
 * since it can lock every login out, and the banner is edited in its own dialog, or read in one
 * by a role that may not edit it.
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
      <AccessModeCard ns={ns} />
      {isEnterpriseOrCloud() && <RecordingCard ns={ns} />}
      <BannerCard ns={ns} />
    </SettingsSection>
  );
}
