import {
  CheckCircleIcon,
  NoSymbolIcon,
  ComputerDesktopIcon,
  UserIcon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { FirewallRule } from "../../types/firewallRule";
import FilterBadge from "../../components/common/FilterBadge";

/* --- Rule Card --- */
export default function RuleCard({
  rule,
  onEdit,
  onDelete,
}: {
  rule: FirewallRule;
  onEdit: () => void;
  onDelete: () => void;
}) {
  return (
    <div className="group bg-card border border-border rounded-xl p-4 hover:border-border-light hover:bg-hover-subtle transition-all">
      <div className="flex items-start justify-between gap-3">
        {/* Left: icon + info */}
        <div className="flex items-start gap-3.5 min-w-0 flex-1">
          <div
            className={`shrink-0 w-9 h-9 rounded-lg flex items-center justify-center mt-0.5 ${
              rule.action === "allow"
                ? "bg-accent-green/10"
                : "bg-accent-red/10"
            }`}
          >
            {rule.action === "allow" ? (
              <CheckCircleIcon className="w-4.5 h-4.5 text-accent-green" />
            ) : (
              <NoSymbolIcon className="w-4.5 h-4.5 text-accent-red" />
            )}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <h3 className="text-sm font-semibold text-text-primary">
                {rule.action === "allow" ? "Allow" : "Deny"}
              </h3>
              <span className="inline-flex items-center px-1.5 py-0.5 bg-hover-strong text-text-muted text-2xs rounded font-mono">
                Priority {rule.priority}
              </span>
              {!rule.active && (
                <span className="inline-flex items-center px-1.5 py-0.5 bg-accent-yellow/10 text-accent-yellow text-2xs rounded font-medium">
                  Inactive
                </span>
              )}
            </div>

            {/* Badges row */}
            <div className="flex flex-wrap items-center gap-2 mt-2.5">
              {/* Source IP */}
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-hover-medium text-text-muted text-2xs rounded">
                <ComputerDesktopIcon className="w-2.5 h-2.5" strokeWidth={2} />
                {rule.source_ip === ".*" ? (
                  "Any IP"
                ) : (
                  <span className="font-mono">{rule.source_ip}</span>
                )}
              </span>
              {/* Username */}
              <span className="inline-flex items-center gap-1 px-1.5 py-0.5 bg-hover-medium text-text-muted text-2xs rounded">
                <UserIcon className="w-2.5 h-2.5" strokeWidth={2} />
                {rule.username === ".*" ? (
                  "All users"
                ) : (
                  <span className="font-mono">{rule.username}</span>
                )}
              </span>
              {/* Device filter */}
              <FilterBadge filter={rule.filter} />
            </div>
          </div>
        </div>

        {/* Right: actions */}
        <div className="flex items-center gap-0.5 shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
          <button
            onClick={onEdit}
            className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
            title="Edit"
          >
            <PencilSquareIcon className="w-4 h-4" />
          </button>
          <button
            onClick={onDelete}
            className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
            title="Delete"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      </div>
    </div>
  );
}
