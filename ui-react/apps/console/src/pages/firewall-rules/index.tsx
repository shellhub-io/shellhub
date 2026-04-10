import { useState } from "react";
import { useFirewallRules, type FirewallRule } from "@/hooks/useFirewallRules";
import { useDeleteFirewallRule } from "@/hooks/useFirewallRuleMutations";
import PageHeader from "@/components/common/PageHeader";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import FilterBadge from "@/components/common/FilterBadge";
import RuleDrawer from "./RuleDrawer";
import RestrictedAction from "@/components/common/RestrictedAction";
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
  NoSymbolIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  UsersIcon,
  Bars3Icon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";

const PER_PAGE = 10;

export default function FirewallRules() {
  const [page, setPage] = useState(1);
  const { rules, totalCount, isLoading } = useFirewallRules({ page });
  const deleteRule = useDeleteFirewallRule();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<FirewallRule | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    priority: number;
  } | null>(null);
  const [search, setSearch] = useState("");

  const openNew = () => {
    setEditTarget(null);
    setDrawerOpen(true);
  };

  const openEdit = (rule: FirewallRule) => {
    setEditTarget(rule);
    setDrawerOpen(true);
  };

  const closeDrawer = () => {
    setDrawerOpen(false);
    setEditTarget(null);
  };

  const totalPages = Math.ceil(totalCount / PER_PAGE);

  const filtered = search
    ? rules.filter(
      (r) =>
        r.action.toLowerCase().includes(search.toLowerCase())
        || r.source_ip.toLowerCase().includes(search.toLowerCase())
        || r.username.toLowerCase().includes(search.toLowerCase())
        || String(r.priority).includes(search),
    )
    : rules;

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
              <span className="text-xs font-medium text-accent-green">Allow</span>
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
          <span className="text-xs font-mono text-text-primary">{rule.source_ip}</span>
        ),
    },
    {
      key: "username",
      header: "Username",
      render: (rule) =>
        rule.username === ".*" ? (
          <span className="text-xs text-text-secondary">All users</span>
        ) : (
          <span className="text-xs font-mono text-text-primary">{rule.username}</span>
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
      key: "actions",
      header: "",
      headerClassName: "w-20",
      render: (rule) => (
        <div className="flex items-center justify-end gap-0.5">
          <RestrictedAction action="firewall:edit">
            <button
              onClick={() => openEdit(rule)}
              className="p-1.5 rounded-md text-text-muted hover:text-primary hover:bg-primary/10 transition-all"
              title="Edit"
            >
              <PencilSquareIcon className="w-4 h-4" />
            </button>
          </RestrictedAction>
          <RestrictedAction action="firewall:remove">
            <button
              onClick={() =>
                setDeleteTarget({ id: rule.id, priority: rule.priority })}
              className="p-1.5 rounded-md text-text-muted hover:text-accent-red hover:bg-accent-red/10 transition-all"
              title="Delete"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  /* Full-page onboarding empty state (no rules at all) */
  if (!isLoading && rules.length === 0) {
    return (
      <div>
        <div className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col">
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] animate-pulse-subtle" />
            <div
              className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-accent-blue/5 rounded-full blur-[100px] animate-pulse-subtle"
              style={{ animationDelay: "1s" }}
            />
            <div className="absolute inset-0 grid-bg opacity-30" />
          </div>

          <div className="flex-1 flex items-center justify-center px-8 py-12">
            <div className="w-full max-w-2xl animate-fade-in">
              <div className="text-center mb-10">
                <div className="w-16 h-16 rounded-2xl bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-6 shadow-lg shadow-primary/5">
                  <ExclamationTriangleIcon className="w-8 h-8 text-primary" />
                </div>

                <span className="inline-block text-2xs font-mono font-semibold uppercase tracking-wide text-primary/80 mb-2">
                  Network Security
                </span>
                <h1 className="text-3xl font-bold text-text-primary mb-3">
                  Firewall Rules
                </h1>
                <p className="text-sm text-text-muted max-w-md mx-auto leading-relaxed">
                  Control who can access your devices and from where. Define
                  rules based on source IP, username, and device filter to
                  enforce your security policies.
                </p>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-10">
                {[
                  {
                    icon: <ExclamationTriangleIcon className="w-5 h-5" />,
                    title: "Allow & Deny",
                    description:
                      "Create rules to allow or block SSH connections based on your criteria.",
                  },
                  {
                    icon: <UsersIcon className="w-5 h-5" />,
                    title: "User Filtering",
                    description:
                      "Restrict access per username, hostname, or source IP address.",
                  },
                  {
                    icon: <Bars3Icon className="w-5 h-5" />,
                    title: "Priority Order",
                    description:
                      "Organize rules by priority to control evaluation order.",
                  },
                ].map((h, idx) => (
                  <div
                    key={h.title}
                    className="bg-card/60 border border-border rounded-xl p-5 text-center animate-slide-up"
                    style={{ animationDelay: `${150 + idx * 100}ms` }}
                  >
                    <div className="w-10 h-10 rounded-lg bg-primary/10 border border-primary/20 flex items-center justify-center mx-auto mb-3 text-primary">
                      {h.icon}
                    </div>
                    <h3 className="text-sm font-semibold text-text-primary mb-1">
                      {h.title}
                    </h3>
                    <p className="text-xs text-text-muted leading-relaxed">
                      {h.description}
                    </p>
                  </div>
                ))}
              </div>

              <div
                className="text-center animate-slide-up"
                style={{ animationDelay: "450ms" }}
              >
                <RestrictedAction action="firewall:create">
                  <button
                    onClick={openNew}
                    className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200 shadow-lg shadow-primary/20"
                  >
                    <PlusIcon className="w-4 h-4" strokeWidth={2} />
                    Add your first rule
                  </button>
                </RestrictedAction>
                <p className="mt-4 text-2xs text-text-muted">
                  Rules are evaluated by priority before connections reach
                  devices.
                </p>
              </div>
            </div>
          </div>
        </div>

        <RuleDrawer
          open={drawerOpen}
          editRule={editTarget}
          onClose={closeDrawer}
        />
      </div>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<ExclamationTriangleIcon className="w-6 h-6" />}
        overline="Security"
        title="Firewall Rules"
        description="Control SSH connections to your devices with allow and deny rules evaluated by priority."
      >
        <RestrictedAction action="firewall:create">
          <button
            onClick={openNew}
            className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
          >
            <PlusIcon className="w-4 h-4" strokeWidth={2} />
            Add Rule
          </button>
        </RestrictedAction>
      </PageHeader>

      <div className="mb-4 animate-fade-in">
        <div className="relative max-w-sm">
          <MagnifyingGlassIcon className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted pointer-events-none" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search by action, priority, IP, or username..."
            className="w-full pl-9 pr-3.5 py-2 bg-card border border-border rounded-lg text-sm text-text-primary placeholder:text-text-secondary focus:outline-none focus:border-primary/50 focus:ring-1 focus:ring-primary/20 transition-all"
          />
        </div>
      </div>

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(rule) => rule.id}
        isLoading={isLoading}
        loadingMessage="Loading firewall rules..."
        page={page}
        totalPages={totalPages}
        totalCount={totalCount}
        itemLabel="rule"
        onPageChange={setPage}
        emptyMessage={search ? `No rules matching \u201C${search}\u201D` : "No firewall rules found"}
      />

      <RuleDrawer
        open={drawerOpen}
        editRule={editTarget}
        onClose={closeDrawer}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          await deleteRule.mutateAsync({ path: { id: deleteTarget!.id } });
          if (rules.length === 1 && page > 1) setPage(page - 1);
          setDeleteTarget(null);
        }}
        title="Delete Firewall Rule"
        description={(
          <>
            Are you sure you want to delete the rule with priority
            {" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.priority}
            </span>
            ? This action cannot be undone.
          </>
        )}
        confirmLabel="Delete"
      />
    </div>
  );
}
