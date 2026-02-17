import { useEffect, useState } from "react";
import { useFirewallRulesStore } from "../../stores/firewallRulesStore";
import { FirewallRule } from "../../types/firewallRule";
import PageHeader from "../../components/common/PageHeader";
import ConfirmDialog from "../../components/common/ConfirmDialog";
import RuleDrawer from "./RuleDrawer";
import RuleCard from "./RuleCard";
import Pagination from "../../components/common/Pagination";
import {
  ExclamationTriangleIcon,
  MagnifyingGlassIcon,
  PlusIcon,
  UsersIcon,
  Bars3Icon,
} from "@heroicons/react/24/outline";

/* ─── Page ─── */
export default function FirewallRules() {
  const { rules, totalCount, loading, page, perPage, fetch, remove } =
    useFirewallRulesStore();
  const [drawerOpen, setDrawerOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<FirewallRule | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    priority: number;
  } | null>(null);
  const [search, setSearch] = useState("");

  useEffect(() => {
    fetch();
  }, [fetch]);

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

  const totalPages = Math.ceil(totalCount / perPage);

  const filtered = search
    ? rules.filter(
        (r) =>
          r.action.toLowerCase().includes(search.toLowerCase()) ||
          r.source_ip.toLowerCase().includes(search.toLowerCase()) ||
          r.username.toLowerCase().includes(search.toLowerCase()) ||
          String(r.priority).includes(search),
      )
    : rules;

  return (
    <div>
      {loading && rules.length === 0 ? (
        <div className="flex items-center justify-center py-16">
          <span className="w-5 h-5 border-2 border-primary/30 border-t-primary rounded-full animate-spin" />
        </div>
      ) : rules.length === 0 ? (
        /* Empty state */
        <div className="relative -mx-8 -mt-8 min-h-[calc(100vh-3.5rem)] flex flex-col">
          <div className="absolute inset-0 overflow-hidden pointer-events-none">
            <div className="absolute -top-32 left-1/3 w-[500px] h-[500px] bg-primary/5 rounded-full blur-[120px] animate-pulse-subtle" />
            <div
              className="absolute bottom-0 right-1/4 w-[400px] h-[400px] bg-accent-blue/5 rounded-full blur-[100px] animate-pulse-subtle"
              style={{ animationDelay: "1s" }}
            />
            <div className="absolute inset-0 grid-bg opacity-30" />
          </div>

          <div className="relative z-10 flex-1 flex items-center justify-center px-8 py-12">
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
                <button
                  onClick={openNew}
                  className="inline-flex items-center gap-2 px-6 py-3 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200 shadow-lg shadow-primary/20"
                >
                  <PlusIcon className="w-4 h-4" strokeWidth={2} />
                  Add your first rule
                </button>
                <p className="mt-4 text-2xs text-text-muted">
                  Rules are evaluated by priority before connections reach
                  devices.
                </p>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <>
          <PageHeader
            variant="decorated"
            icon={<ExclamationTriangleIcon className="w-6 h-6" />}
            overline="Security"
            title="Firewall Rules"
            description="Control SSH connections to your devices with allow and deny rules evaluated by priority."
          >
            <button
              onClick={openNew}
              className="flex items-center gap-2 px-4 py-2 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold transition-all duration-200"
            >
              <PlusIcon className="w-4 h-4" strokeWidth={2} />
              Add Rule
            </button>
          </PageHeader>

          {/* Search bar */}
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

          {filtered.length === 0 ? (
            <div className="py-12 text-center animate-fade-in">
              <p className="text-sm text-text-muted">
                No rules matching &ldquo;{search}&rdquo;
              </p>
            </div>
          ) : (
            <>
              <div className="space-y-2 animate-fade-in">
                {filtered.map((rule) => (
                  <RuleCard
                    key={rule.id}
                    rule={rule}
                    onEdit={() => openEdit(rule)}
                    onDelete={() =>
                      setDeleteTarget({ id: rule.id, priority: rule.priority })
                    }
                  />
                ))}
              </div>

              <Pagination
                page={page}
                totalPages={totalPages}
                totalCount={totalCount}
                itemLabel="rule"
                onPageChange={fetch}
              />
            </>
          )}
        </>
      )}

      {/* Drawer */}
      <RuleDrawer
        open={drawerOpen}
        editRule={editTarget}
        onClose={closeDrawer}
      />

      {/* Delete Dialog */}
      <ConfirmDialog
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onConfirm={async () => {
          await remove(deleteTarget!.id);
          setDeleteTarget(null);
        }}
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
      />
    </div>
  );
}
