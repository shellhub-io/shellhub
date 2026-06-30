import { useState } from "react";
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
  NoSymbolIcon,
  PlusIcon,
  UsersIcon,
  Bars3Icon,
  PencilSquareIcon,
  TrashIcon,
} from "@heroicons/react/24/outline";
import { Badge, Button, IconButton } from "@shellhub/design-system/primitives";
import { type FirewallRulesResponse as FirewallRule } from "@/client";
import ActiveBadge from "@/components/common/ActiveBadge";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import { DataTable, type Column } from "@shellhub/design-system/components";
import EmptyState from "@/components/common/EmptyState";
import FilterBadge from "@/components/common/FilterBadge";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import SearchField from "@/components/common/fields/SearchField";
import { useDeleteFirewallRule } from "@/hooks/useFirewallRuleMutations";
import { useFirewallRules } from "@/hooks/useFirewallRules";
import { usePaginatedListState } from "@/hooks/usePaginatedListState";
import RuleDrawer from "./RuleDrawer";

const PER_PAGE = 10;

type FirewallRulesParams = {
  page: number;
  search: string;
};

const DEFAULTS: FirewallRulesParams = {
  page: 1,
  search: "",
};

export default function FirewallRules() {
  const { params, setPage, setSearch } =
    usePaginatedListState<FirewallRulesParams>({ defaults: DEFAULTS });

  const { rules, totalCount, isLoading } = useFirewallRules({ page: params.page });
  const deleteRule = useDeleteFirewallRule();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<FirewallRule | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    priority: number;
  } | null>(null);
  const [deleteError, setDeleteError] = useState<string | null>(null);

  const closeDelete = () => {
    setDeleteError(null);
    setDeleteTarget(null);
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setDeleteError(null);
    try {
      await deleteRule.mutateAsync({ path: { id: deleteTarget.id } });
      if (rules.length === 1 && params.page > 1 && !params.search) setPage(params.page - 1);
      closeDelete();
    } catch (err) {
      setDeleteError(
        err instanceof Error ? err.message : "Failed to delete firewall rule.",
      );
    }
  };

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

  const filtered = params.search
    ? rules.filter(
        (r) =>
          r.action.toLowerCase().includes(params.search.toLowerCase()) ||
          r.source_ip.toLowerCase().includes(params.search.toLowerCase()) ||
          r.username.toLowerCase().includes(params.search.toLowerCase()) ||
          String(r.priority).includes(params.search),
      )
    : rules;

  const columns: Column<FirewallRule>[] = [
    {
      key: "priority",
      header: "Priority",
      render: (rule) => <Badge color="primary">{rule.priority}</Badge>,
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
      render: (rule) => <ActiveBadge active={rule.active} />,
    },
    {
      key: "actions",
      header: "",
      headerClassName: "w-20",
      render: (rule) => (
        <div className="flex items-center justify-end gap-0.5">
          <RestrictedAction action="firewall:edit">
            <IconButton
              variant="primary"
              aria-label={`Edit firewall rule with priority ${rule.priority}`}
              title="Edit"
              onClick={() => openEdit(rule)}
            >
              <PencilSquareIcon className="w-4 h-4" aria-hidden="true" />
            </IconButton>
          </RestrictedAction>
          <RestrictedAction action="firewall:remove">
            <IconButton
              variant="danger"
              aria-label={`Delete firewall rule with priority ${rule.priority}`}
              title="Delete"
              onClick={() =>
                setDeleteTarget({ id: rule.id, priority: rule.priority })
              }
            >
              <TrashIcon className="w-4 h-4" aria-hidden="true" />
            </IconButton>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  /* Full-page onboarding empty state (no rules at all) */
  if (!isLoading && rules.length === 0) {
    return (
      <>
        <EmptyState
          icon={<ExclamationTriangleIcon className="w-8 h-8" />}
          overline="Network Security"
          title="Firewall Rules"
          description="Control who can access your devices and from where. Define rules based on source IP, username, and device filter to enforce your security policies."
          features={[
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
          ]}
          footnote="Rules are evaluated by priority before connections reach devices."
        >
          <RestrictedAction action="firewall:create">
            <Button
              size="lg"
              onClick={openNew}
              icon={
                <PlusIcon
                  className="w-4 h-4"
                  strokeWidth={2}
                  aria-hidden="true"
                />
              }
            >
              Add your first rule
            </Button>
          </RestrictedAction>
        </EmptyState>

        <RuleDrawer
          open={drawerOpen}
          editRule={editTarget}
          onClose={closeDrawer}
        />
      </>
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
          <Button
            onClick={openNew}
            icon={
              <PlusIcon
                className="w-4 h-4"
                strokeWidth={2}
                aria-hidden="true"
              />
            }
          >
            Add Rule
          </Button>
        </RestrictedAction>
      </PageHeader>

      <SearchField
        className="mb-4"
        value={params.search}
        onChange={setSearch}
        placeholder="Search by action, priority, IP, or username..."
        aria-label="Search firewall rules by action, priority, IP, or username"
      />

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(rule) => rule.id}
        isLoading={isLoading}
        loadingMessage="Loading firewall rules..."
        {...(!params.search && {
          page: params.page,
          totalPages,
          totalCount,
          itemLabel: "rule",
          onPageChange: setPage,
        })}
        emptyMessage={
          params.search
            ? `No rules matching \u201C${params.search}\u201D`
            : "No firewall rules found"
        }
      />

      <RuleDrawer
        open={drawerOpen}
        editRule={editTarget}
        onClose={closeDrawer}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onClose={closeDelete}
        onConfirm={confirmDelete}
        title="Delete Firewall Rule"
        description={
          <>
            Are you sure you want to delete the rule with priority{" "}
            <span className="font-medium text-text-primary">
              {deleteTarget?.priority}
            </span>
            ? This action cannot be undone.
          </>
        }
        confirmLabel="Delete"
      >
        {deleteError && (
          <p className="text-xs text-accent-red" role="alert">
            {deleteError}
          </p>
        )}
      </ConfirmDialog>
    </div>
  );
}
