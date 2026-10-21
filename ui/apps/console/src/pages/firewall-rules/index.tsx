import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
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
import {
  Badge,
  Button,
  Callout,
  IconButton,
} from "@shellhub/design-system/primitives";
import { apiErrorMessage } from "@/api/errors";
import {
  deleteFirewallRuleMutation,
  getFirewallRules,
  getFirewallRulesQueryKey,
  type FirewallRulesResponse as FirewallRule,
} from "@/client";
import { paginatedQueryOptions } from "@/api/pagination";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";
import { useRowAction } from "@/hooks/useRowAction";
import ActiveBadge from "@/components/common/ActiveBadge";
import ConfirmDialog from "@/components/common/ConfirmDialog";
import DataTable, { type Column } from "@/components/common/DataTable";
import EmptyState from "@/components/common/EmptyState";
import FilterBadge from "@/components/common/FilterBadge";
import PageHeader from "@/components/common/PageHeader";
import RestrictedAction from "@/components/common/RestrictedAction";
import SearchField from "@/components/common/fields/SearchField";
import { usePagedList } from "@/hooks/usePagedList";
import { useNavSectionTitle } from "@/components/layout/navSections";
import RuleModal from "./RuleModal";
import { PER_PAGE } from "@/utils/pagination";
import ObjectName from "@/components/common/ObjectName";

type FirewallRulesParams = {
  page: number;
  search: string;
};

const DEFAULTS: FirewallRulesParams = {
  page: 1,
  search: "",
};

/**
 * The firewall rules list, in priority order — which is also the order they are evaluated in.
 */
export default function FirewallRules() {
  const sectionTitle = useNavSectionTitle("/firewall-rules");
  const {
    params, setSearch, rows: rules, totalCount, isLoading, error, pagination,
  } = usePagedList({
    defaults: DEFAULTS,
    query: ({ page }: FirewallRulesParams) =>
      paginatedQueryOptions(getFirewallRules, getFirewallRulesQueryKey, {
        query: { page, per_page: PER_PAGE },
      }),
  });
  const invalidateRules = useInvalidateByIds("getFirewallRules");
  const deleteRuleMutation = useMutation({
    ...deleteFirewallRuleMutation(),
    onSuccess: invalidateRules,
  });
  const deleteRule = useRowAction({
    item: "firewall rule",
    run: (rule: FirewallRule) =>
      deleteRuleMutation.mutateAsync({ path: { id: rule.id } }),
  });
  const [modalOpen, setModalOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<FirewallRule | null>(null);

  const openNew = () => {
    setEditTarget(null);
    setModalOpen(true);
  };

  const openEdit = (rule: FirewallRule) => {
    setEditTarget(rule);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditTarget(null);
  };

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
              onClick={() => deleteRule.open(rule)}
            >
              <TrashIcon className="w-4 h-4" aria-hidden="true" />
            </IconButton>
          </RestrictedAction>
        </div>
      ),
    },
  ];

  if (!isLoading && !error && totalCount === 0) {
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

        <RuleModal
          open={modalOpen}
          editRule={editTarget}
          onClose={closeModal}
        />
      </>
    );
  }

  return (
    <div>
      <PageHeader
        icon={<ExclamationTriangleIcon className="w-6 h-6" />}
        overline={sectionTitle}
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

      {error && (
        <Callout variant="error" className="mb-4">
          {apiErrorMessage(error)}
        </Callout>
      )}

      <DataTable
        columns={columns}
        data={filtered}
        rowKey={(rule) => rule.id}
        isLoading={isLoading}
        loadingMessage="Loading firewall rules..."
        {...(!params.search && {
          ...pagination,
          itemLabel: "rule",
        })}
        emptyMessage={
          params.search
            ? `No rules matching \u201C${params.search}\u201D`
            : "No firewall rules found"
        }
      />

      <RuleModal
        open={modalOpen}
        editRule={editTarget}
        onClose={closeModal}
      />

      <ConfirmDialog
        open={!!deleteRule.target}
        onClose={deleteRule.close}
        onConfirm={deleteRule.confirm}
        icon={<TrashIcon />}
        title="Delete firewall rule"
        description={
          <>
            The rule with priority{" "}
            <ObjectName>{deleteRule.target?.priority}</ObjectName> stops filtering
            connections. This can't be undone.
          </>
        }
        confirmLabel="Delete rule"
        errorMessage={deleteRule.error}
      />
    </div>
  );
}
