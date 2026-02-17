import { useState, useEffect, FormEvent } from "react";
import { useFirewallRulesStore } from "../../stores/firewallRulesStore";
import { FirewallRule, FirewallFilter } from "../../types/firewallRule";
import Drawer from "../../components/common/Drawer";
import { LABEL, INPUT, INPUT_MONO } from "../../utils/styles";
import RadioCard from "../../components/common/RadioCard";
import TagsSelector from "../../components/common/TagsSelector";
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
import { DevicesIcon as DevicesIconComponent } from "../../components/icons";

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
  editRule: FirewallRule | null;
  onClose: () => void;
}) {
  const { create, update } = useFirewallRulesStore();
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

  useEffect(() => {
    if (!open) return;
    if (editRule) {
      setPriority(String(editRule.priority));
      setAction(editRule.action);
      setActive(editRule.active);
      setSourceIpOption(editRule.source_ip === ".*" ? "all" : "restrict");
      setSourceIp(editRule.source_ip === ".*" ? "" : editRule.source_ip);
      setUsernameOption(editRule.username === ".*" ? "all" : "restrict");
      setUsername(editRule.username === ".*" ? "" : editRule.username);
      if (editRule.filter.tags && editRule.filter.tags.length > 0) {
        setFilterOption("tags");
        setSelectedTags(editRule.filter.tags);
      } else if (
        editRule.filter.hostname &&
        editRule.filter.hostname !== ".*"
      ) {
        setFilterOption("hostname");
        setHostname(editRule.filter.hostname);
      } else {
        setFilterOption("all");
        setHostname("");
        setSelectedTags([]);
      }
      setError(null);
    } else {
      setPriority("");
      setAction("allow");
      setActive(true);
      setSourceIpOption("all");
      setSourceIp("");
      setUsernameOption("all");
      setUsername("");
      setFilterOption("all");
      setHostname("");
      setSelectedTags([]);
      setError(null);
    }
  }, [open, editRule]);

  const buildFilter = (): FirewallFilter => {
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

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    const payload = {
      priority: priorityNum,
      action,
      active,
      source_ip: sourceIpOption === "all" ? ".*" : sourceIp.trim(),
      username: usernameOption === "all" ? ".*" : username.trim(),
      filter: buildFilter(),
    };
    try {
      if (isEdit && editRule) {
        await update(editRule.id, payload);
      } else {
        await create(payload);
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
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={submitting || confirmDisabled}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            {submitting ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Saving...
              </span>
            ) : isEdit ? (
              "Save Changes"
            ) : (
              "Create Rule"
            )}
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Status toggle */}
        <div>
          <label className={LABEL}>Status</label>
          <button
            type="button"
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
        <div>
          <label className={LABEL}>Priority</label>
          <input
            type="number"
            min={1}
            value={priority}
            onChange={(e) => setPriority(e.target.value)}
            placeholder="e.g. 100"
            autoFocus={open}
            className={INPUT}
          />
          {priorityError && (
            <p className="mt-1 text-2xs text-accent-red">{priorityError}</p>
          )}
          <p className="mt-1 text-2xs text-text-muted">
            Higher values are evaluated first. Must be greater than 0.
          </p>
        </div>

        {/* Policy */}
        <div>
          <label className={LABEL}>Policy</label>
          <div className="space-y-2">
            <RadioCard
              selected={action === "allow"}
              onClick={() => setAction("allow")}
              icon={<CheckCircleIcon className="w-4 h-4" />}
              label="Allow"
              description="Permit the connection when this rule matches."
            />
            <RadioCard
              selected={action === "deny"}
              onClick={() => setAction("deny")}
              icon={<NoSymbolIcon className="w-4 h-4" />}
              label="Deny"
              description="Block the connection when this rule matches."
            />
          </div>
        </div>

        {/* Source IP */}
        <div>
          <label className={LABEL}>Source IP</label>
          <div className="space-y-2">
            <RadioCard
              selected={sourceIpOption === "all"}
              onClick={() => setSourceIpOption("all")}
              icon={GlobeIcon}
              label="All source IPs"
              description="Match connections from any IP address."
            />
            <RadioCard
              selected={sourceIpOption === "restrict"}
              onClick={() => setSourceIpOption("restrict")}
              icon={IpIcon}
              label="Restrict with regexp"
              description="Match connections from IPs matching a pattern."
            />
          </div>
          {sourceIpOption === "restrict" && (
            <input
              type="text"
              value={sourceIp}
              onChange={(e) => setSourceIp(e.target.value)}
              placeholder="e.g. 192\.168\.1\..*"
              className={`${INPUT_MONO} mt-2`}
            />
          )}
        </div>

        {/* Username */}
        <div>
          <label className={LABEL}>Username</label>
          <div className="space-y-2">
            <RadioCard
              selected={usernameOption === "all"}
              onClick={() => setUsernameOption("all")}
              icon={UsersIcon}
              label="All users"
              description="Match connections for any username."
            />
            <RadioCard
              selected={usernameOption === "restrict"}
              onClick={() => setUsernameOption("restrict")}
              icon={UserIcon}
              label="Restrict with regexp"
              description="Match connections for usernames matching a pattern."
            />
          </div>
          {usernameOption === "restrict" && (
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="e.g. root"
              className={`${INPUT_MONO} mt-2`}
            />
          )}
        </div>

        {/* Device filter */}
        <div>
          <label className={LABEL}>Device filter</label>
          <div className="space-y-2">
            <RadioCard
              selected={filterOption === "all"}
              onClick={() => setFilterOption("all")}
              icon={DevicesIcon}
              label="All devices"
              description="Match connections to any device in the namespace."
            />
            <RadioCard
              selected={filterOption === "hostname"}
              onClick={() => setFilterOption("hostname")}
              icon={HostnameIcon}
              label="Filter by hostname"
              description="Restrict to devices matching a hostname pattern."
            />
            <RadioCard
              selected={filterOption === "tags"}
              onClick={() => setFilterOption("tags")}
              icon={TagIcon}
              label="Filter by tags"
              description="Restrict to devices matching specific tags."
            />
          </div>
          {filterOption === "hostname" && (
            <input
              type="text"
              value={hostname}
              onChange={(e) => setHostname(e.target.value)}
              placeholder="e.g. web-.*"
              className={`${INPUT_MONO} mt-2`}
            />
          )}
          {filterOption === "tags" && (
            <div className="mt-2">
              <TagsSelector
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
