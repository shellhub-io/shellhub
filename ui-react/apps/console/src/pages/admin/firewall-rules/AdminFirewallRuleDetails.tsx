import { useParams, Link } from "react-router-dom";
import {
  ChevronRightIcon,
  ShieldExclamationIcon,
  InformationCircleIcon,
  FunnelIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { useAdminFirewallRule } from "@/hooks/useAdminFirewallRules";
import CopyButton from "@/components/common/CopyButton";
import FilterBadge from "@/components/common/FilterBadge";

const LABEL
  = "text-2xs font-mono font-semibold uppercase tracking-label text-text-muted";
const VALUE = "text-sm text-text-primary font-medium mt-0.5";

export default function AdminFirewallRuleDetails() {
  const { id } = useParams<{ id: string }>();
  const { data: rule, isLoading, error } = useAdminFirewallRule(id ?? "");

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-24" role="status">
        <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        <span className="sr-only">Loading firewall rule details</span>
      </div>
    );
  }

  if (error || !rule) {
    return (
      <div className="text-center py-24">
        <ShieldExclamationIcon
          className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
          strokeWidth={1}
        />
        <p className="text-sm text-text-muted mb-2">Firewall rule not found</p>
        <Link
          to="/admin/firewall-rules"
          className="text-sm text-primary hover:underline"
        >
          Back to firewall rules
        </Link>
      </div>
    );
  }

  const isAllow = rule.action === "allow";

  return (
    <div className="animate-fade-in">
      {/* Breadcrumb */}
      <nav aria-label="Breadcrumb" className="flex items-center gap-1.5 mb-5">
        <Link
          to="/admin/firewall-rules"
          className="text-2xs font-mono text-text-muted hover:text-primary transition-colors"
        >
          Firewall Rules
        </Link>
        <ChevronRightIcon
          className="w-3 h-3 text-text-muted/40"
          strokeWidth={2}
        />
        <span className="text-2xs font-mono text-text-secondary truncate min-w-0">
          {rule.id}
        </span>
      </nav>

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div
          className={`w-14 h-14 rounded-xl flex items-center justify-center shrink-0 border ${
            isAllow
              ? "bg-accent-green/10 border-accent-green/20"
              : "bg-accent-red/10 border-accent-red/20"
          }`}
        >
          {isAllow ? (
            <CheckCircleIcon className="w-7 h-7 text-accent-green" />
          ) : (
            <NoSymbolIcon className="w-7 h-7 text-accent-red" />
          )}
        </div>
        <div>
          <h1 className="text-2xl font-bold text-text-primary">
            {isAllow ? "Allow Rule" : "Deny Rule"}
          </h1>
          <div className="flex items-center gap-2 mt-1.5">
            <span className="inline-flex items-center px-1.5 py-0.5 bg-hover-strong text-text-muted text-2xs rounded font-mono">
              Priority {rule.priority}
            </span>
            {rule.active ? (
              <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green border border-accent-green/20">
                Active
              </span>
            ) : (
              <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                Inactive
              </span>
            )}
          </div>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Rule Properties Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Rule Properties
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>ID</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <span
                  className="text-xs font-mono text-text-primary truncate min-w-0"
                  title={rule.id}
                >
                  {rule.id}
                </span>
                <CopyButton text={rule.id} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Namespace</dt>
              <dd className="flex items-center gap-1 mt-0.5">
                <Link
                  to={`/admin/namespaces/${rule.tenant_id}`}
                  className="text-xs font-mono text-primary hover:underline truncate min-w-0"
                >
                  {rule.tenant_id}
                </Link>
                <CopyButton text={rule.tenant_id} />
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Priority</dt>
              <dd className={VALUE}>{rule.priority}</dd>
            </div>
            <div>
              <dt className={LABEL}>Action</dt>
              <dd className="flex items-center gap-1.5 mt-0.5">
                {isAllow ? (
                  <>
                    <CheckCircleIcon className="w-4 h-4 text-accent-green" />
                    <span className="text-sm font-medium text-accent-green">
                      Allow
                    </span>
                  </>
                ) : (
                  <>
                    <NoSymbolIcon className="w-4 h-4 text-accent-red" />
                    <span className="text-sm font-medium text-accent-red">
                      Deny
                    </span>
                  </>
                )}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Status</dt>
              <dd className="mt-0.5">
                {rule.active ? (
                  <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green border border-accent-green/20">
                    Active
                  </span>
                ) : (
                  <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
                    Inactive
                  </span>
                )}
              </dd>
            </div>
          </dl>
        </div>

        {/* Connection Criteria Card */}
        <div className="bg-card border border-border rounded-xl p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <FunnelIcon className="w-4 h-4 text-primary" />
            Connection Criteria
          </h3>
          <dl className="space-y-3">
            <div>
              <dt className={LABEL}>Source IP</dt>
              <dd className={VALUE}>
                {rule.source_ip === ".*" ? (
                  <span className="text-text-secondary">Any IP</span>
                ) : (
                  <span className="font-mono text-xs">{rule.source_ip}</span>
                )}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Username</dt>
              <dd className={VALUE}>
                {rule.username === ".*" ? (
                  <span className="text-text-secondary">All users</span>
                ) : (
                  <span className="font-mono text-xs">{rule.username}</span>
                )}
              </dd>
            </div>
            <div>
              <dt className={LABEL}>Device Filter</dt>
              <dd className="mt-1">
                <FilterBadge filter={rule.filter} />
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </div>
  );
}
