import { useState, useMemo } from "react";
import { useNavigate, Link } from "react-router-dom";
import {
  ShieldExclamationIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import Alert from "@/components/common/Alert";
import { useAdminFirewallRules } from "@/hooks/useAdminFirewallRules";
import PageHeader from "@/components/common/PageHeader";
import DataTable, { type Column } from "@/components/common/DataTable";
import SearchField from "@/components/common/fields/SearchField";
import FilterBadge from "@/components/common/FilterBadge";
import { type FirewallRulesResponse as FirewallRule } from "@/client";

const PER_PAGE = 10;

export default function AdminFirewallRules() {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");

  const { rules, totalCount, isLoading, error } = useAdminFirewallRules({
    page,
    perPage: PER_PAGE,
  });

  const filtered = useMemo(() => {
    if (!search) return rules;
    const q = search.toLowerCase();
    return rules.filter(
      (r) =>
        r.action.toLowerCase().includes(q) ||
        r.source_ip.toLowerCase().includes(q) ||
        r.username.toLowerCase().includes(q) ||
        String(r.priority).includes(q),
    );
  }, [rules, search]);

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const columns: Column<FirewallRule>[] = [
    {
      key: "priority",
      header: "Priority",
      render: (rule) => (
        <span className="inline-flex items-center px-1.5 py-0.5 bg-primary/10 text-primary text-2xs rounded font-mono font-medium">
          {rule.priority}
        </span>
      ),
    },
    {
      key: "action",
      header: "Action",
      render: (rule) => (
        <div className="flex items-center gap-1.5">
          {rule.action === "allow" ? (
            <>
              <CheckCircleIcon className="w-4 h-4 text-accent-green" />
              <span className="text-xs font-medium text-accent-green">
                Allow
              </span>
            </>
          ) : (
            <>
              <NoSymbolIcon className="w-4 h-4 text-accent-red" />
              <span className="text-xs font-medium text-accent-red">Deny</span>
            </>
          )}
        </div>
      ),
    },
    {
      key: "source_ip",
      header: "Source IP",
      render: (rule) =>
        rule.source_ip === ".*" ? (
          <span className="text-xs text-text-secondary">Any IP</span>
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {rule.source_ip}
          </span>
        ),
    },
    {
      key: "username",
      header: "Username",
      render: (rule) =>
        rule.username === ".*" ? (
          <span className="text-xs text-text-secondary">All users</span>
        ) : (
          <span className="text-xs font-mono text-text-primary">
            {rule.username}
          </span>
        ),
    },
    {
      key: "filter",
      header: "Device Filter",
      render: (rule) => <FilterBadge filter={rule.filter} />,
    },
    {
      key: "status",
      header: "Status",
      render: (rule) =>
        rule.active ? (
          <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-green/10 text-accent-green border border-accent-green/20">
            Active
          </span>
        ) : (
          <span className="inline-flex items-center px-2 py-0.5 text-2xs font-semibold rounded-md bg-accent-yellow/10 text-accent-yellow border border-accent-yellow/20">
            Inactive
          </span>
        ),
    },
    {
      key: "namespace",
      header: "Namespace",
      render: (rule) => (
        <Link
          to={`/admin/namespaces/${rule.tenant_id}`}
          onClick={(e) => e.stopPropagation()}
          className="text-xs text-primary hover:underline font-mono truncate block max-w-[180px]"
          title={rule.tenant_id}
        >
          {rule.tenant_id}
        </Link>
      ),
    },
  ];

  return (
    <div>
      <PageHeader
        icon={<ShieldExclamationIcon className="w-6 h-6" />}
        overline="Firewall Administration"
        title="Firewall Rules"
        description="View all firewall rules configured across the instance"
      />

      <SearchField
        className="mb-5"
        value={search}
        onChange={(next) => {
          setSearch(next);
          setPage(1);
        }}
        placeholder="Search by action, priority, IP, or username..."
        aria-label="Search firewall rules by action, priority, IP, or username"
      />

      {error && (
        <Alert variant="error" className="mb-4">
          {error.message}
        </Alert>
      )}

      <DataTable
        columns={columns}
        data={search ? filtered : rules}
        rowKey={(rule) => rule.id}
        isLoading={isLoading}
        loadingMessage="Loading firewall rules..."
        onRowClick={(rule) => void navigate(`/admin/firewall-rules/${rule.id}`)}
        {...(!search && {
          page,
          totalPages,
          totalCount,
          itemLabel: "rule",
          onPageChange: setPage,
        })}
        emptyState={
          <div className="text-center">
            <ShieldExclamationIcon
              className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
              strokeWidth={1}
            />
            <p className="text-xs font-mono text-text-muted">
              {search
                ? `No rules matching "${search}"`
                : "No firewall rules found"}
            </p>
          </div>
        }
      />
    </div>
  );
}
