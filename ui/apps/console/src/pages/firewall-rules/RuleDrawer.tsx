import { useState, FormEvent } from "react";
import { useResetOnOpen } from "@/hooks/useResetOnOpen";
import {
  useCreateFirewallRule,
  useUpdateFirewallRule,
} from "@/hooks/useFirewallRuleMutations";
import type { FirewallRulesRequest, FirewallRulesResponse } from "@/client";
import Drawer from "@/components/common/Drawer";
import RadioCard from "@/components/common/fields/RadioCard";
import RadioGroupField from "@/components/common/fields/RadioGroupField";
import TagsSelector from "@/components/common/fields/TagsSelector";
import InputField from "@/components/common/fields/InputField";
import {
  UserGroupIcon,
  UserIcon as UserIconHero,
  ClipboardDocumentListIcon,
  TagIcon as TagIconHero,
  GlobeAltIcon,
  ComputerDesktopIcon,
  CheckCircleIcon,
  NoSymbolIcon,
  ExclamationCircleIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon as DevicesIconComponent } from "@/components/icons";
import NumericInput from "@/components/common/fields/NumericInput";
import { LABEL } from "@/utils/styles";
import { Button } from "@shellhub/design-system/primitives";

/* ─── Icons ─── */
const UsersIcon = <UserGroupIcon className="w-4 h-4" />;

const UserIcon = <UserIconHero className="w-4 h-4" />;

const DevicesIcon = <DevicesIconComponent className="w-4 h-4" />;

const HostnameIcon = <ClipboardDocumentListIcon className="w-4 h-4" />;

const TagIcon = <TagIconHero className="w-4 h-4" />;

const GlobeIcon = <GlobeAltIcon className="w-4 h-4" />;

const IpIcon = <ComputerDesktopIcon className="w-4 h-4" />;

/* ─── Rule Drawer ─── */
export default function RuleDrawer({
  open,
  editRule,
  onClose,
}: {
  open: boolean;
  editRule: FirewallRulesResponse | null;
  onClose: () => void;
}) {
  const createRule = useCreateFirewallRule();
  const updateRule = useUpdateFirewallRule();
  const isEdit = !!editRule;

  const [priority, setPriority] = useState("");
  const [action, setAction] = useState<"allow" | "deny">("allow");
  const [active, setActive] = useState(true);
  const [sourceIpOption, setSourceIpOption] = useState<"all" | "restrict">(
    "all",
  );
  const [sourceIp, setSourceIp] = useState("");
  const [usernameOption, setUsernameOption] = useState<"all" | "restrict">(
    "all",
  );
  const [username, setUsername] = useState("");
  const [filterOption, setFilterOption] = useState<"all" | "hostname" | "tags">(
    "all",
  );
  const [hostname, setHostname] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useResetOnOpen(open, () => {
    const filterInit = editRule
      ? editRule.filter.tags.length > 0
        ? "tags"
        : editRule.filter.hostname && editRule.filter.hostname !== ".*"
          ? "hostname"
          : "all"
      : "all";

    setPriority(editRule ? String(editRule.priority) : "");
    setAction(editRule?.action ?? "allow");
    setActive(editRule?.active ?? true);
    setSourceIpOption(
      editRule ? (editRule.source_ip === ".*" ? "all" : "restrict") : "all",
    );
    setSourceIp(
      editRule && editRule.source_ip !== ".*" ? editRule.source_ip : "",
    );
    setUsernameOption(
      editRule ? (editRule.username === ".*" ? "all" : "restrict") : "all",
    );
    setUsername(
      editRule && editRule.username !== ".*" ? editRule.username : "",
    );
    setFilterOption(filterInit);
    setHostname(
      editRule && filterInit === "hostname"
        ? (editRule.filter.hostname ?? "")
        : "",
    );
    setSelectedTags(
      editRule && filterInit === "tags"
        ? editRule.filter.tags.map((t) => t.name)
        : [],
    );
    setSubmitting(false);
    setError(null);
  });

  const buildFilter = (): FirewallRulesRequest["filter"] => {
    if (filterOption === "hostname" && hostname) return { hostname };
    if (filterOption === "tags" && selectedTags.length > 0)
      return { tags: selectedTags };
    return { hostname: ".*" };
  };

  const tagError =
    selectedTags.length > 3
      ? "You can select up to 3 tags"
      : filterOption === "tags" && selectedTags.length === 0
        ? "Select at least one tag"
        : undefined;

  const priorityNum = parseInt(priority, 10);
  const priorityError =
    priority && (isNaN(priorityNum) || priorityNum <= 0)
      ? "Priority must be a positive integer"
      : undefined;

  const confirmDisabled =
    !priority.trim() ||
    !!priorityError ||
    (sourceIpOption === "restrict" && !sourceIp.trim()) ||
    (usernameOption === "restrict" && !username.trim()) ||
    (filterOption === "hostname" && !hostname.trim()) ||
    (filterOption === "tags" &&
      (selectedTags.length === 0 || selectedTags.length > 3));

  const handleSubmit = async (e?: FormEvent) => {
    e?.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    const body = {
      priority: priorityNum,
      action,
      active,
      source_ip: sourceIpOption === "all" ? ".*" : sourceIp.trim(),
      username: usernameOption === "all" ? ".*" : username.trim(),
      filter: buildFilter(),
    } satisfies FirewallRulesRequest;
    try {
      if (isEdit && editRule) {
        await updateRule.mutateAsync({ path: { id: editRule.id }, body });
      } else {
        await createRule.mutateAsync({ body });
      }
      onClose();
    } catch (err: unknown) {
      setError(
        err instanceof Error
          ? err.message
          : `Failed to ${isEdit ? "update" : "create"} rule`,
      );
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Firewall Rule" : "New Firewall Rule"}
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
            {submitting ? "Saving..." : isEdit ? "Save Changes" : "Create Rule"}
          </Button>
        </>
      }
    >
      <form onSubmit={(e) => void handleSubmit(e)} className="space-y-5">
        {/* Status toggle */}
        <div>
          <span id="rule-status-label" className={LABEL}>
            Status
          </span>
          <button
            type="button"
            role="switch"
            aria-checked={active}
            aria-labelledby="rule-status-label"
            onClick={() => setActive(!active)}
            className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
              active ? "bg-accent-green" : "bg-border"
            }`}
          >
            <span
              className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
                active ? "translate-x-6" : "translate-x-1"
              }`}
            />
          </button>
          <span
            className={`ml-2.5 text-sm ${active ? "text-accent-green" : "text-text-muted"}`}
          >
            {active ? "Active" : "Inactive"}
          </span>
        </div>

        {/* Priority */}
        <NumericInput
          id="rule-priority"
          label="Priority"
          value={priority}
          onChange={setPriority}
          placeholder="e.g. 100"

          hint="Higher values are evaluated first. Must be greater than 0."
          error={priorityError || undefined}
        />

        {/* Policy */}
        <RadioGroupField label="Policy" value={action} onChange={setAction}>
          <RadioCard
            value="allow"
            icon={<CheckCircleIcon className="w-4 h-4" />}
            label="Allow"
            description="Permit the connection when this rule matches."
          />
          <RadioCard
            value="deny"
            icon={<NoSymbolIcon className="w-4 h-4" />}
            label="Deny"
            description="Block the connection when this rule matches."
          />
        </RadioGroupField>

        {/* Source IP */}
        <div>
          <RadioGroupField
            label="Source IP"
            value={sourceIpOption}
            onChange={setSourceIpOption}
          >
            <RadioCard
              value="all"
              icon={GlobeIcon}
              label="All source IPs"
              description="Match connections from any IP address."
            />
            <RadioCard
              value="restrict"
              icon={IpIcon}
              label="Restrict with regexp"
              description="Match connections from IPs matching a pattern."
            />
          </RadioGroupField>
          {sourceIpOption === "restrict" && (
            <div className="mt-2">
              <InputField
                id="rule-source-ip-pattern"
                label="Source IP pattern"
                hideLabel
                value={sourceIp}
                onChange={setSourceIp}
                placeholder="e.g. 192\.168\.1\..*"
                variant="mono"
              />
            </div>
          )}
        </div>

        {/* Username */}
        <div>
          <RadioGroupField
            label="Username"
            value={usernameOption}
            onChange={setUsernameOption}
          >
            <RadioCard
              value="all"
              icon={UsersIcon}
              label="All users"
              description="Match connections for any username."
            />
            <RadioCard
              value="restrict"
              icon={UserIcon}
              label="Restrict with regexp"
              description="Match connections for usernames matching a pattern."
            />
          </RadioGroupField>
          {usernameOption === "restrict" && (
            <div className="mt-2">
              <InputField
                id="rule-username-pattern"
                label="Username pattern"
                hideLabel
                value={username}
                onChange={setUsername}
                placeholder="e.g. root"
                variant="mono"
              />
            </div>
          )}
        </div>

        {/* Device filter */}
        <div>
          <RadioGroupField
            label="Device filter"
            value={filterOption}
            onChange={setFilterOption}
          >
            <RadioCard
              value="all"
              icon={DevicesIcon}
              label="All devices"
              description="Match connections to any device in the namespace."
            />
            <RadioCard
              value="hostname"
              icon={HostnameIcon}
              label="Filter by hostname"
              description="Restrict to devices matching a hostname pattern."
            />
            <RadioCard
              value="tags"
              icon={TagIcon}
              label="Filter by tags"
              description="Restrict to devices matching specific tags."
            />
          </RadioGroupField>
          {filterOption === "hostname" && (
            <div className="mt-2">
              <InputField
                id="rule-hostname-pattern"
                label="Hostname pattern"
                hideLabel
                value={hostname}
                onChange={setHostname}
                placeholder="e.g. web-.*"
                variant="mono"
              />
            </div>
          )}
          {filterOption === "tags" && (
            <div className="mt-2">
              <TagsSelector
                id="rule-filter-tags"
                label="Filter by tags"
                selected={selectedTags}
                onChange={setSelectedTags}
                error={tagError}
              />
            </div>
          )}
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
