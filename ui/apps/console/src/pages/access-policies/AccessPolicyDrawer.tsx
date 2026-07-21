import { useState, FormEvent } from "react";
import {
  UsersIcon,
  UserIcon,
  CpuChipIcon,
  ShieldCheckIcon,
  TagIcon,
  CommandLineIcon,
  ClipboardDocumentListIcon,
  ExclamationCircleIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon } from "@shellhub/design-system/primitives";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import { useAuthStore } from "@/stores/authStore";
import { useNamespace, type NamespaceMember } from "@/hooks/useNamespaces";
import { useServiceAccounts } from "@/hooks/useServiceAccounts";
import {
  useCreateAccessPolicy,
  useUpdateAccessPolicy,
} from "@/hooks/useAccessPolicyMutations";
import type { AccessPolicy, AccessPolicyRequest } from "@/client";
import { ROLES } from "@/pages/team/helpers";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import ChipInput from "@/components/common/fields/ChipInput";
import InputField from "@/components/common/fields/InputField";
import Drawer from "@/components/common/Drawer";
import { INPUT, LABEL } from "@/utils/styles";
import { Button } from "@shellhub/design-system/primitives";

type SubjectType = "all-members" | "role" | "user" | "service-account";
type FilterOption = "all" | "hostname" | "tags";
type LoginsOption = "any" | "specific";

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
  const createPolicy = useCreateAccessPolicy();
  const updatePolicy = useUpdateAccessPolicy();
  const isEdit = !!editPolicy;

  // Service accounts share the namespace membership but are not human members, so
  // keep them out of the member picker (they carry the "service" role). Target a
  // service account from a policy via the role=service subject instead.
  const members = (namespace?.members ?? []).filter(
    (m): m is NamespaceMember =>
      !!m.id && !!m.role && !!m.email && String(m.role) !== "service",
  );

  const { serviceAccounts } = useServiceAccounts();

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

    // A subject of type "user" whose value is a service account is shown as the
    // "service account" option, not the member picker (which excludes them).
    const editValue = editPolicy?.subject.value ?? "";
    const editIsServiceAccount =
      editPolicy?.subject.type === "user" &&
      serviceAccounts.some((sa) => sa.id === editValue);

    setName(editPolicy?.name ?? "");
    setEffect(editPolicy?.effect ?? "allow");
    setSubjectType(
      editIsServiceAccount
        ? "service-account"
        : (editPolicy?.subject.type ?? "all-members"),
    );
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
    setFilterOption(filterInit);
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

  const tagError =
    selectedTags.length > 3
      ? "You can select up to 3 tags"
      : filterOption === "tags" && selectedTags.length === 0
        ? "Select at least one tag"
        : undefined;

  const confirmDisabled =
    !name.trim() ||
    (subjectType === "user" && !userValue) ||
    (subjectType === "service-account" && !saValue) ||
    (filterOption === "hostname" && !hostname.trim()) ||
    (filterOption === "tags" &&
      (selectedTags.length === 0 || selectedTags.length > 3)) ||
    (loginsOption === "specific" && logins.length === 0);

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
          placeholder="Name used to identify the policy"
        />

        {/* Effect */}
        <RadioGroupField label="Effect" value={effect} onChange={setEffect}>
          <RadioCard
            value="allow"
            icon={<CheckCircleIcon className="w-4 h-4" />}
            label="Allow"
            description="Grant the matched access."
          />
          <RadioCard
            value="deny"
            icon={<NoSymbolIcon className="w-4 h-4" />}
            label="Deny"
            description="Block the matched access. Deny is evaluated first and wins over any allow."
          />
        </RadioGroupField>

        {/* Subject */}
        <div>
          <RadioGroupField
            label={effect === "deny" ? "Block access for" : "Grant access to"}
            value={subjectType}
            onChange={setSubjectType}
          >
            <RadioCard
              value="all-members"
              icon={<UsersIcon className="w-4 h-4" />}
              label="All members"
              description="Every member of this namespace is granted access."
            />
            <RadioCard
              value="role"
              icon={<ShieldCheckIcon className="w-4 h-4" />}
              label="Members with a role"
              description="Only members holding the selected role are granted access."
            />
            <RadioCard
              value="user"
              icon={<UserIcon className="w-4 h-4" />}
              label="A specific member"
              description="Only the selected member is granted access."
            />
            <RadioCard
              value="service-account"
              icon={<CpuChipIcon className="w-4 h-4" />}
              label="A service account"
              description="Only the selected service account is granted access."
            />
          </RadioGroupField>
          {subjectType === "role" && (
            <div className="mt-2">
              <span className={LABEL}>Role</span>
              <select
                id="access-policy-role"
                value={roleValue}
                onChange={(e) => setRoleValue(e.target.value)}
                className={INPUT}
              >
                {ROLES.map((role) => (
                  <option key={role} value={role}>
                    {role.charAt(0).toUpperCase() + role.slice(1)}
                  </option>
                ))}
              </select>
            </div>
          )}
          {subjectType === "user" && (
            <div className="mt-2">
              <span className={LABEL}>Member</span>
              <select
                id="access-policy-user"
                value={userValue}
                onChange={(e) => setUserValue(e.target.value)}
                className={INPUT}
              >
                <option value="" disabled>
                  Select a member...
                </option>
                {members.map((m) => (
                  <option key={m.id} value={m.id}>
                    {m.email}
                  </option>
                ))}
              </select>
            </div>
          )}
          {subjectType === "service-account" && (
            <div className="mt-2">
              <span className={LABEL}>Service account</span>
              <select
                id="access-policy-service-account"
                value={saValue}
                onChange={(e) => setSaValue(e.target.value)}
                className={INPUT}
              >
                <option value="" disabled>
                  {serviceAccounts.length === 0
                    ? "No service accounts yet"
                    : "Select a service account..."}
                </option>
                {serviceAccounts.map((sa) => (
                  <option key={sa.id} value={sa.id}>
                    {sa.name}
                  </option>
                ))}
              </select>
            </div>
          )}
        </div>

        {/* Device access */}
        <div>
          <RadioGroupField
            label="Devices"
            value={filterOption}
            onChange={setFilterOption}
          >
            <RadioCard
              value="all"
              icon={<DevicesIcon className="w-4 h-4" />}
              label="All devices"
              description="The policy applies to any device in the namespace."
            />
            <RadioCard
              value="hostname"
              icon={<ClipboardDocumentListIcon className="w-4 h-4" />}
              label="Filter by hostname"
              description="Restrict access using a regexp pattern for hostname."
            />
            <RadioCard
              value="tags"
              icon={<TagIcon className="w-4 h-4" />}
              label="Filter by tags"
              description="Restrict access to devices matching specific tags."
            />
          </RadioGroupField>
          {filterOption === "hostname" && (
            <div className="mt-2">
              <InputField
                id="access-policy-hostname"
                label="Hostname pattern"
                hideLabel
                value={hostname}
                onChange={setHostname}
                placeholder="e.g. .*"
                variant="mono"
              />
            </div>
          )}
          {filterOption === "tags" && (
            <div className="mt-2">
              <TagsSelector
                id="access-policy-filter-tags"
                label="Filter by tags"
                selected={selectedTags}
                onChange={setSelectedTags}
                error={tagError}
              />
            </div>
          )}
        </div>

        {/* Allowed logins */}
        <div>
          <RadioGroupField
            label={effect === "deny" ? "Blocked logins" : "Allowed logins"}
            value={loginsOption}
            onChange={setLoginsOption}
          >
            <RadioCard
              value="any"
              icon={<CommandLineIcon className="w-4 h-4" />}
              label="Any login"
              description="Connect as any unix login on the matched devices."
            />
            <RadioCard
              value="specific"
              icon={<UserIcon className="w-4 h-4" />}
              label="Specific logins"
              description="Restrict to an explicit list of unix logins."
            />
          </RadioGroupField>
          {loginsOption === "specific" && (
            <div className="mt-2">
              <ChipInput
                id="access-policy-logins"
                label="Logins"
                placeholder="e.g. deploy, ubuntu"
                hint="Press Enter or comma to add a login."
                values={logins}
                onChange={setLogins}
              />
            </div>
          )}
        </div>

        {/* Source IP */}
        <div>
          <ChipInput
            id="access-policy-source-ip"
            label="Source IP"
            placeholder="e.g. 10.0.0.0/8, 203.0.113.5/32"
            hint="Restrict to these CIDRs. Empty = any IP."
            values={sourceIP}
            onChange={setSourceIP}
          />
        </div>

        {/* Step-up */}
        <div>
          <span id="access-policy-stepup-label" className={LABEL}>
            Browser approval
          </span>
          <div className="flex items-center gap-2.5">
            <button
              type="button"
              role="switch"
              aria-checked={requireStepUp}
              aria-labelledby="access-policy-stepup-label"
              onClick={() => setRequireStepUp(!requireStepUp)}
              className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
                requireStepUp ? "bg-primary" : "bg-border"
              }`}
            >
              <span
                className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                  requireStepUp ? "translate-x-6" : "translate-x-1"
                }`}
              />
            </button>
            <span className="text-sm text-text-secondary">
              Require browser approval each session
            </span>
          </div>
          <p className="mt-1.5 text-xs text-text-muted">
            Even with an enrolled key, matching access still needs a per-session
            browser approval before the session proceeds.
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
