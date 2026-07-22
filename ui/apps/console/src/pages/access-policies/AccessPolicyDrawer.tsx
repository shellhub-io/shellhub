import { useState, useRef, FormEvent, ReactNode } from "react";
import {
  UsersIcon,
  UserIcon,
  CpuChipIcon,
  ShieldCheckIcon,
  TagIcon,
  CommandLineIcon,
  ClipboardDocumentListIcon,
  ExclamationCircleIcon,
  CheckIcon,
  ChevronDownIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon } from "@shellhub/design-system/primitives";
import { cn } from "@shellhub/design-system/cn";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useClickOutside } from "@/hooks/useClickOutside";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace, type NamespaceMember } from "@/hooks/useNamespaces";
import { useServiceAccounts } from "@/hooks/useServiceAccounts";
import { useTags } from "@/hooks/useTags";
import {
  useCreateAccessPolicy,
  useUpdateAccessPolicy,
} from "@/hooks/useAccessPolicyMutations";
import type { AccessPolicy, AccessPolicyRequest } from "@/client";
import { ROLES } from "@/pages/team/helpers";
import ChipInput from "@/components/common/fields/ChipInput";
import SourceIpInput from "@/components/common/fields/SourceIpInput";
import InputField from "@/components/common/fields/InputField";
import Drawer from "@/components/common/Drawer";
import { LABEL } from "@/utils/styles";
import { Button } from "@shellhub/design-system/primitives";

type SubjectType = "all-members" | "role" | "user" | "service-account";
type FilterOption = "all" | "hostname" | "tags";
type LoginsOption = "any" | "specific";

/* A field label above a control. */
function Label({ children }: { children: ReactNode }) {
  return <span className={LABEL}>{children}</span>;
}

/* Selector box that opens an inline dropdown below it, closing on outside click.
   `trigger` renders the current selection as pills; `children` is the dropdown body. */
function PickerBox({
  trigger,
  empty,
  active,
  children,
}: {
  trigger: ReactNode;
  empty: boolean;
  active?: boolean;
  children: (close: () => void) => ReactNode;
}) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);
  useClickOutside(ref, () => setOpen(false));

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        onClick={() => setOpen((o) => !o)}
        className={cn(
          "w-full min-h-[44px] flex flex-wrap items-center gap-1.5 px-3 py-2 bg-card border rounded-lg text-left transition-colors",
          open || active
            ? "border-primary/60"
            : "border-border hover:border-border-light",
        )}
      >
        {trigger}
        <ChevronDownIcon
          className="w-4 h-4 text-text-muted ml-auto shrink-0"
          strokeWidth={2}
        />
      </button>
      {open && (
        <div className="absolute z-40 mt-1.5 w-full bg-card border border-border-light rounded-xl shadow-2xl overflow-hidden">
          {children(() => setOpen(false))}
        </div>
      )}
      {empty && !open ? null : null}
    </div>
  );
}

/* One selectable row inside a picker dropdown, with an optional inventory count. */
function Row({
  icon,
  label,
  sub,
  meta,
  selected,
  onClick,
}: {
  icon: ReactNode;
  label: ReactNode;
  sub?: string;
  meta?: ReactNode;
  selected: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "w-full flex items-center gap-2.5 px-2 py-2 rounded-lg text-left transition-colors",
        selected ? "bg-primary/10" : "hover:bg-primary/10",
      )}
    >
      <span className="grid place-items-center w-6 h-6 rounded-md bg-card text-text-secondary shrink-0">
        {icon}
      </span>
      <span className="min-w-0">
        <span className="block text-sm text-text-primary truncate">
          {label}
        </span>
        {sub && (
          <span className="block text-xs text-text-muted truncate">{sub}</span>
        )}
      </span>
      {meta && (
        <span className="ml-auto text-xs text-text-muted flex items-center gap-1.5">
          {meta}
        </span>
      )}
      <CheckIcon
        className={cn(
          "w-4 h-4 text-primary shrink-0",
          selected ? "opacity-100" : "opacity-0",
          meta ? "" : "ml-auto",
        )}
        strokeWidth={2.5}
      />
    </button>
  );
}

const TABBTN = (on: boolean) =>
  cn(
    "px-2.5 py-1.5 text-xs rounded-md border font-medium transition-colors",
    on
      ? "bg-primary/12 text-primary border-primary/30"
      : "bg-transparent text-text-secondary border-transparent hover:text-text-primary",
  );

/* Small pill used inside a selector box trigger. */
function Pill({
  icon,
  children,
  count,
}: {
  icon: ReactNode;
  children: ReactNode;
  count?: number;
}) {
  return (
    <span className="inline-flex items-center gap-1.5 px-2 py-0.5 text-xs rounded-md font-medium bg-primary/10 text-primary">
      {icon}
      {children}
      {count !== undefined && <span className="text-text-muted">{count}</span>}
    </span>
  );
}

function AccessPolicyDrawer({
  open,
  editPolicy,
  onClose,
}: {
  open: boolean;
  editPolicy: AccessPolicy | null;
  onClose: () => void;
}) {
  const { tenant: tenantId } = useAuthStore();
  const { namespace } = useNamespace(tenantId ?? "");
  const { tags: allTagObjects } = useTags();
  const allTags = allTagObjects.map((t) => t.name);
  const createPolicy = useCreateAccessPolicy();
  const updatePolicy = useUpdateAccessPolicy();
  const isEdit = !!editPolicy;

  // Service accounts share the namespace membership but are not human members, so keep them
  // out of the member picker (they carry the "service" role); reach one via its own tab.
  const members = (namespace?.members ?? []).filter(
    (m): m is NamespaceMember =>
      !!m.id && !!m.role && !!m.email && String(m.role) !== "service",
  );
  const { serviceAccounts } = useServiceAccounts();
  const roleMemberCount = (role: string) =>
    members.filter((m) => String(m.role) === role).length;

  const [name, setName] = useState("");
  const [effect, setEffect] = useState<"allow" | "deny">("allow");
  const [subjectType, setSubjectType] = useState<SubjectType>("all-members");
  const [roleValue, setRoleValue] = useState<string>("administrator");
  const [userValue, setUserValue] = useState<string>("");
  const [saValue, setSaValue] = useState<string>("");
  const [filterOption, setFilterOption] = useState<FilterOption>("all");
  const [hostname, setHostname] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [loginsOption, setLoginsOption] = useState<LoginsOption>("any");
  const [logins, setLogins] = useState<string[]>([]);
  const [sourceIP, setSourceIP] = useState<string[]>([]);
  const [requireStepUp, setRequireStepUp] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [whoTab, setWhoTab] = useState<"role" | "user" | "service-account">(
    "role",
  );
  const [devTab, setDevTab] = useState<FilterOption>("all");

  useResetOnOpen(open, () => {
    const filterInit: FilterOption = editPolicy
      ? editPolicy.filter.tags.length > 0
        ? "tags"
        : editPolicy.filter.hostname && editPolicy.filter.hostname !== ".*"
          ? "hostname"
          : "all"
      : "all";
    const loginsInit: LoginsOption =
      editPolicy &&
      !(editPolicy.logins.length === 1 && editPolicy.logins[0] === "*")
        ? "specific"
        : "any";

    const editValue = editPolicy?.subject.value ?? "";
    const editIsServiceAccount =
      editPolicy?.subject.type === "user" &&
      serviceAccounts.some((sa) => sa.id === editValue);

    const subjInit: SubjectType = editIsServiceAccount
      ? "service-account"
      : (editPolicy?.subject.type ?? "all-members");

    setName(editPolicy?.name ?? "");
    setEffect(editPolicy?.effect ?? "allow");
    setSubjectType(subjInit);
    setRoleValue(
      editPolicy?.subject.type === "role"
        ? editPolicy.subject.value
        : "administrator",
    );
    setUserValue(
      editPolicy?.subject.type === "user" && !editIsServiceAccount
        ? editValue
        : "",
    );
    setSaValue(editIsServiceAccount ? editValue : "");
    setWhoTab(subjInit === "all-members" ? "role" : subjInit);
    setFilterOption(filterInit);
    setDevTab(filterInit);
    setHostname(
      editPolicy && filterInit === "hostname"
        ? (editPolicy.filter.hostname ?? "")
        : "",
    );
    setSelectedTags(
      editPolicy && filterInit === "tags"
        ? editPolicy.filter.tags.map((t) => t.name)
        : [],
    );
    setLoginsOption(loginsInit);
    setLogins(loginsInit === "specific" ? (editPolicy?.logins ?? []) : []);
    setSourceIP(editPolicy?.source_ip ?? []);
    setRequireStepUp(editPolicy?.require_step_up ?? false);
    setSubmitting(false);
    setError(null);
  });

  const buildSubject = (): AccessPolicyRequest["subject"] => {
    if (subjectType === "role") return { type: "role", value: roleValue };
    if (subjectType === "user") return { type: "user", value: userValue };
    // A service account is targeted through a user subject bound to its id.
    if (subjectType === "service-account")
      return { type: "user", value: saValue };
    return { type: "all-members", value: "" };
  };
  const buildFilter = (): AccessPolicyRequest["filter"] => {
    if (filterOption === "hostname" && hostname) return { hostname };
    if (filterOption === "tags" && selectedTags.length > 0)
      return { tags: selectedTags };
    return { hostname: ".*" };
  };
  const buildLogins = (): string[] => (loginsOption === "any" ? ["*"] : logins);

  const confirmDisabled =
    !name.trim() ||
    (subjectType === "user" && !userValue) ||
    (subjectType === "service-account" && !saValue) ||
    (filterOption === "hostname" && !hostname.trim()) ||
    (filterOption === "tags" &&
      (selectedTags.length === 0 || selectedTags.length > 3)) ||
    (loginsOption === "specific" && logins.length === 0);

  /* ---- summaries for the trigger pills + consequence callout ---- */
  const memberById = (id: string) => members.find((m) => m.id === id);
  const saById = (id: string) => serviceAccounts.find((s) => s.id === id);

  const subjectLabel = (): string => {
    if (subjectType === "role") return `the ${roleValue} role`;
    if (subjectType === "user")
      return memberById(userValue)?.email ?? "a member";
    if (subjectType === "service-account")
      return saById(saValue)?.name ?? "a service account";
    return "all members";
  };
  const deviceLabel = (): string => {
    if (filterOption === "tags")
      return selectedTags.length
        ? `devices tagged ${selectedTags.join(", ")}`
        : "devices";
    if (filterOption === "hostname")
      return `devices matching /${hostname || "…"}/`;
    return "all devices";
  };
  const loginLabel = (): string =>
    loginsOption === "any" ? "any login (incl. root)" : logins.join(", ");

  const isBroad =
    effect === "allow" &&
    subjectType === "all-members" &&
    loginsOption === "any" &&
    filterOption === "all";

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    const body: AccessPolicyRequest = {
      name: name.trim(),
      effect,
      subject: buildSubject(),
      filter: buildFilter(),
      logins: buildLogins(),
      source_ip: sourceIP,
      require_step_up: requireStepUp,
    };
    try {
      if (isEdit && editPolicy) {
        await updatePolicy.mutateAsync({ path: { id: editPolicy.id }, body });
      } else {
        await createPolicy.mutateAsync({ body });
      }
      onClose();
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : `Failed to ${isEdit ? "update" : "create"} access policy`,
      );
    } finally {
      setSubmitting(false);
    }
  };

  /* ---- Who trigger ---- */
  const whoTrigger =
    subjectType === "all-members" ? (
      <Pill icon={<UsersIcon className="w-3.5 h-3.5" />}>All members</Pill>
    ) : subjectType === "role" ? (
      <Pill
        icon={<ShieldCheckIcon className="w-3.5 h-3.5" />}
        count={roleMemberCount(roleValue)}
      >
        {roleValue}
      </Pill>
    ) : subjectType === "service-account" ? (
      <Pill icon={<CpuChipIcon className="w-3.5 h-3.5" />}>
        {saById(saValue)?.name ?? "select…"}
      </Pill>
    ) : userValue ? (
      <Pill icon={<UserIcon className="w-3.5 h-3.5" />}>
        {memberById(userValue)?.email}
      </Pill>
    ) : (
      <span className="text-sm text-text-muted">
        Select a role, member, or service account…
      </span>
    );

  /* ---- Devices trigger ---- */
  const devTrigger =
    filterOption === "all" ? (
      <Pill icon={<DevicesIcon className="w-3.5 h-3.5" />}>All devices</Pill>
    ) : filterOption === "tags" ? (
      selectedTags.length ? (
        <>
          {selectedTags.map((t) => (
            <Pill key={t} icon={<TagIcon className="w-3.5 h-3.5" />}>
              {t}
            </Pill>
          ))}
        </>
      ) : (
        <span className="text-sm text-text-muted">Pick tags…</span>
      )
    ) : (
      <Pill icon={<ClipboardDocumentListIcon className="w-3.5 h-3.5" />}>
        <span className="font-mono">/{hostname || "…"}/</span>
      </Pill>
    );

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Access Policy" : "New Access Policy"}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button
            variant="primary"
            type="submit"
            onClick={() => void handleSubmit()}
            disabled={submitting || confirmDisabled}
            loading={submitting}
          >
            {submitting
              ? "Saving..."
              : isEdit
                ? "Save Changes"
                : "Create Policy"}
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {/* Name */}
        <InputField
          id="access-policy-name"
          label="Name"
          value={name}
          onChange={setName}
          placeholder="e.g. Operators to prod"
        />

        {/* Effect */}
        <div>
          <Label>Effect</Label>
          <div className="inline-flex bg-card border border-border rounded-lg p-0.5 gap-0.5">
            {(["allow", "deny"] as const).map((e) => (
              <button
                key={e}
                type="button"
                onClick={() => setEffect(e)}
                className={cn(
                  "flex items-center justify-center gap-2 px-4 py-1.5 rounded-md text-sm font-medium transition-colors",
                  effect === e
                    ? e === "allow"
                      ? "bg-accent-green/15 text-accent-green"
                      : "bg-accent-red/15 text-accent-red"
                    : "text-text-secondary hover:text-text-primary",
                )}
              >
                <span
                  className={cn(
                    "w-1.5 h-1.5 rounded-full",
                    effect === e ? "bg-current" : "bg-text-muted",
                  )}
                />
                {e === "allow" ? "Allow" : "Deny"}
              </button>
            ))}
          </div>
          {effect === "deny" && (
            <p className="mt-1.5 text-xs text-accent-red flex items-center gap-1.5">
              <ShieldCheckIcon className="w-3.5 h-3.5" strokeWidth={2} />
              Deny wins — evaluated before every allow.
            </p>
          )}
        </div>

        {/* Who */}
        <div>
          <Label>{effect === "deny" ? "Block access for" : "Who"}</Label>
          <PickerBox
            trigger={whoTrigger}
            empty={subjectType === "user" && !userValue}
          >
            {(close) => (
              <div className="p-2">
                <Row
                  icon={<UsersIcon className="w-4 h-4" />}
                  label="All members"
                  meta={<span className="text-text-muted">everyone</span>}
                  selected={subjectType === "all-members"}
                  onClick={() => {
                    setSubjectType("all-members");
                    close();
                  }}
                />
                <div className="flex gap-1 px-1 py-2">
                  <button
                    type="button"
                    className={TABBTN(whoTab === "role")}
                    onClick={() => setWhoTab("role")}
                  >
                    Roles
                  </button>
                  <button
                    type="button"
                    className={TABBTN(whoTab === "user")}
                    onClick={() => setWhoTab("user")}
                  >
                    Members
                  </button>
                  <button
                    type="button"
                    className={TABBTN(whoTab === "service-account")}
                    onClick={() => setWhoTab("service-account")}
                  >
                    Service accounts
                  </button>
                </div>
                <div className="max-h-56 overflow-y-auto">
                  {whoTab === "role" &&
                    ROLES.map((role) => (
                      <Row
                        key={role}
                        icon={<ShieldCheckIcon className="w-4 h-4" />}
                        label={role}
                        meta={
                          <>
                            <UsersIcon className="w-3.5 h-3.5" />{" "}
                            {roleMemberCount(role)}
                          </>
                        }
                        selected={subjectType === "role" && roleValue === role}
                        onClick={() => {
                          setSubjectType("role");
                          setRoleValue(role);
                          close();
                        }}
                      />
                    ))}
                  {whoTab === "user" &&
                    members.map((m) => (
                      <Row
                        key={m.id}
                        icon={<UserIcon className="w-4 h-4" />}
                        label={m.email}
                        selected={subjectType === "user" && userValue === m.id}
                        onClick={() => {
                          setSubjectType("user");
                          setUserValue(m.id);
                          close();
                        }}
                      />
                    ))}
                  {whoTab === "service-account" &&
                    (serviceAccounts.length ? (
                      serviceAccounts.map((sa) => (
                        <Row
                          key={sa.id}
                          icon={<CpuChipIcon className="w-4 h-4" />}
                          label={sa.name}
                          sub="service account"
                          selected={
                            subjectType === "service-account" &&
                            saValue === sa.id
                          }
                          onClick={() => {
                            setSubjectType("service-account");
                            setSaValue(sa.id);
                            close();
                          }}
                        />
                      ))
                    ) : (
                      <p className="px-2 py-3 text-xs text-text-muted">
                        No service accounts yet.
                      </p>
                    ))}
                </div>
              </div>
            )}
          </PickerBox>
        </div>

        {/* connector */}
        <div className="flex items-center gap-2 -my-1 pl-1 text-xs text-text-muted">
          <ChevronDownIcon
            className="w-3.5 h-3.5 text-border-light"
            strokeWidth={2}
          />
          {effect === "deny" ? "is blocked from" : "can SSH into"}
        </div>

        {/* Devices */}
        <div>
          <Label>Devices</Label>
          <PickerBox
            trigger={devTrigger}
            empty={filterOption === "tags" && !selectedTags.length}
          >
            {(close) => (
              <div className="p-2">
                <div className="flex gap-1 px-1 pb-2">
                  {(["all", "tags", "hostname"] as const).map((o) => (
                    <button
                      key={o}
                      type="button"
                      className={TABBTN(devTab === o)}
                      onClick={() => setDevTab(o)}
                    >
                      {o === "all" ? "All" : o === "tags" ? "Tags" : "Hostname"}
                    </button>
                  ))}
                </div>
                {devTab === "all" && (
                  <Row
                    icon={<DevicesIcon className="w-4 h-4" />}
                    label="All devices"
                    meta={<span className="text-text-muted">every device</span>}
                    selected={filterOption === "all"}
                    onClick={() => {
                      setFilterOption("all");
                      close();
                    }}
                  />
                )}
                {devTab === "tags" && (
                  <div className="max-h-56 overflow-y-auto">
                    {allTags.length ? (
                      allTags.map((t) => (
                        <Row
                          key={t}
                          icon={<TagIcon className="w-4 h-4" />}
                          label={<span className="font-mono">{t}</span>}
                          selected={
                            filterOption === "tags" && selectedTags.includes(t)
                          }
                          onClick={() => {
                            setFilterOption("tags");
                            setSelectedTags((prev) =>
                              prev.includes(t)
                                ? prev.filter((x) => x !== t)
                                : prev.length < 3
                                  ? [...prev, t]
                                  : prev,
                            );
                          }}
                        />
                      ))
                    ) : (
                      <p className="px-2 py-3 text-xs text-text-muted">
                        No tags in this namespace.
                      </p>
                    )}
                    <p className="px-2 pt-1 text-2xs text-text-muted">
                      Up to 3 tags · any match.
                    </p>
                  </div>
                )}
                {devTab === "hostname" && (
                  <div className="p-1">
                    <input
                      value={hostname}
                      onChange={(e) => {
                        setFilterOption("hostname");
                        setHostname(e.target.value);
                      }}
                      placeholder="^prod-.*$"
                      className="w-full px-3 py-2 bg-bg border border-border rounded-lg text-sm font-mono text-text-primary placeholder:text-text-muted outline-none focus:border-primary/60"
                    />
                    <p className="px-1 pt-2 text-2xs text-text-muted">
                      A regexp matched against device hostnames.
                    </p>
                  </div>
                )}
              </div>
            )}
          </PickerBox>
        </div>

        {/* Allowed logins */}
        <div>
          <Label>
            {effect === "deny" ? "Blocked logins" : "Allowed logins"}
          </Label>
          {loginsOption === "any" ? (
            <div className="flex items-center gap-2 min-h-[44px] px-3 py-2 bg-card border border-border rounded-lg">
              <span className="inline-flex items-center gap-1.5 px-2 py-0.5 text-xs rounded-md bg-primary/10 text-primary">
                <CommandLineIcon className="w-3.5 h-3.5" /> Any login
              </span>
              <button
                type="button"
                onClick={() => setLoginsOption("specific")}
                className="ml-auto text-xs text-primary hover:underline"
              >
                Restrict…
              </button>
            </div>
          ) : (
            <>
              <ChipInput
                id="access-policy-logins"
                label=""
                placeholder="type a unix login + Enter (e.g. deploy, root)"
                hint="Empty means any login."
                values={logins}
                onChange={(next) => {
                  setLogins(next);
                  if (next.length === 0) setLoginsOption("any");
                }}
              />
            </>
          )}
        </div>

        {/* Source IP */}
        <SourceIpInput
          id="access-policy-source-ip"
          label="Source IP"
          hint="Restrict to these CIDRs. A bare IP becomes a /32 host. Empty = any IP."
          values={sourceIP}
          onChange={setSourceIP}
        />

        {/* Require approval — toggle card */}
        <button
          type="button"
          onClick={() => setRequireStepUp((v) => !v)}
          className={cn(
            "w-full flex items-center gap-3 px-3.5 py-3 border rounded-xl text-left transition-colors",
            requireStepUp
              ? "border-primary/40 bg-primary/[0.06]"
              : "border-border bg-card",
          )}
        >
          <span
            className={cn(
              "grid place-items-center w-8 h-8 rounded-lg bg-surface shrink-0",
              requireStepUp ? "text-primary" : "text-text-secondary",
            )}
          >
            <ShieldCheckIcon className="w-4 h-4" strokeWidth={2} />
          </span>
          <span className="min-w-0">
            <span className="block text-sm font-semibold text-text-primary">
              Require browser approval
            </span>
            <span className="block text-xs text-text-muted">
              Even with an enrolled key, each session must be approved in the
              browser.
            </span>
          </span>
          <span
            className={cn(
              "relative ml-auto inline-flex h-5 w-9 items-center rounded-full transition-colors shrink-0",
              requireStepUp ? "bg-primary" : "bg-border-light",
            )}
          >
            <span
              className={cn(
                "inline-block h-4 w-4 transform rounded-full bg-white transition-transform",
                requireStepUp ? "translate-x-4" : "translate-x-0.5",
              )}
            />
          </span>
        </button>

        {/* Consequence callout */}
        <div className="flex gap-2.5 px-3.5 py-3 rounded-xl border border-border bg-card text-sm leading-relaxed">
          <CheckIcon
            className={cn(
              "w-4 h-4 shrink-0 mt-0.5",
              effect === "deny" ? "text-accent-red" : "text-accent-green",
            )}
            strokeWidth={2.5}
          />
          <p className="text-text-secondary">
            {effect === "deny" ? (
              <>
                <b className="text-text-primary">Denies</b> {subjectLabel()}{" "}
                from reaching{" "}
                <b className="text-text-primary">{deviceLabel()}</b>. Evaluated
                before allows.
              </>
            ) : (
              <>
                Lets <b className="text-text-primary">{subjectLabel()}</b> SSH
                into <b className="text-text-primary">{deviceLabel()}</b> as{" "}
                {loginLabel()}
                {sourceIP.length > 0 &&
                  `, from ${sourceIP.length} network${sourceIP.length > 1 ? "s" : ""}`}
                {requireStepUp && ", with browser approval"}.
                {isBroad && " This is the broadest grant possible."}
              </>
            )}
          </p>
        </div>

        {/* Error */}
        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}

export default AccessPolicyDrawer;
