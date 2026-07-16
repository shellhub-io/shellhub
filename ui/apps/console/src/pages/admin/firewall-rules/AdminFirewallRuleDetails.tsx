import { useParams, Link } from "react-router-dom";
import {
  ShieldExclamationIcon,
  InformationCircleIcon,
  FunnelIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { cn } from "@shellhub/design-system/cn";
import { useAdminFirewallRule } from "@/hooks/useAdminFirewallRules";
import ActiveBadge from "@/components/common/ActiveBadge";
import Breadcrumb from "@/components/common/Breadcrumb";
import CopyButton from "@/components/common/CopyButton";
import FilterBadge from "@/components/common/FilterBadge";
import InfoItem from "@/components/common/InfoItem";
import PageLoader from "@/components/common/PageLoader";
import { Card } from "@shellhub/design-system/primitives";

export default function AdminFirewallRuleDetails() {
  const { id } = useParams<{ id: string }>();
  const { data: rule, isLoading, error } = useAdminFirewallRule(id ?? "");

  if (isLoading) {
    return <PageLoader label="Loading firewall rule details" />;
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
      <Breadcrumb
        items={[
          { label: "Firewall Rules", to: "/admin/firewall-rules" },
          { label: rule.id },
        ]}
      />

      {/* Header */}
      <div className="flex items-start gap-4 mb-8">
        <div
          className={cn(
            "w-14 h-14 rounded-xl flex items-center justify-center shrink-0 border",
            isAllow
              ? "bg-accent-green/10 border-accent-green/20"
              : "bg-accent-red/10 border-accent-red/20",
          )}
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
            <ActiveBadge active={rule.active} />
          </div>
        </div>
      </div>

      {/* Info Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Rule Properties Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <InformationCircleIcon className="w-4 h-4 text-primary" />
            Rule Properties
          </h3>
          <dl className="space-y-3">
            <InfoItem label="ID" value={rule.id} mono copyable />
            <InfoItem label="Namespace">
              <Link
                to={`/admin/namespaces/${rule.tenant_id}`}
                className="text-xs font-mono text-primary hover:underline truncate min-w-0"
              >
                {rule.tenant_id}
              </Link>
              <CopyButton text={rule.tenant_id} />
            </InfoItem>
            <InfoItem label="Priority" value={String(rule.priority)} />
            <InfoItem label="Action">
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
            </InfoItem>
            <InfoItem label="Status">
              <ActiveBadge active={rule.active} />
            </InfoItem>
          </dl>
        </Card>

        {/* Connection Criteria Card */}
        <Card className="p-5 space-y-4">
          <h3 className="text-xs font-semibold text-text-primary flex items-center gap-2">
            <FunnelIcon className="w-4 h-4 text-primary" />
            Connection Criteria
          </h3>
          <dl className="space-y-3">
            <InfoItem label="Source IP">
              {rule.source_ip === ".*" ? (
                <span className="text-text-secondary">Any IP</span>
              ) : (
                <span className="font-mono text-xs">{rule.source_ip}</span>
              )}
            </InfoItem>
            <InfoItem label="Username">
              {rule.username === ".*" ? (
                <span className="text-text-secondary">All users</span>
              ) : (
                <span className="font-mono text-xs">{rule.username}</span>
              )}
            </InfoItem>
            <InfoItem label="Device Filter">
              <FilterBadge filter={rule.filter} />
            </InfoItem>
          </dl>
        </Card>
      </div>
    </div>
  );
}
