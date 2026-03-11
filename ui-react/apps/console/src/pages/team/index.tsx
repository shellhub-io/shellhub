import { useState } from "react";
import { UserGroupIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "../../stores/authStore";
import PageHeader from "../../components/common/PageHeader";
import MembersTab from "./MembersTab";
import ApiKeysTab from "./ApiKeysTab";

/* --- Page --- */

const tabs = [
  { label: "Members", value: "members" },
  { label: "API Keys", value: "api-keys" },
];

export default function Team() {
  const [tab, setTab] = useState("members");
  const tenant = useAuthStore((s) => s.tenant);

  return (
    <div>
      <PageHeader
        icon={<UserGroupIcon className="w-6 h-6" />}
        overline="Management"
        title="Team"
        description="Manage namespace members and API keys"
      />

      {/* Tabs */}
      <div className="flex items-center h-8 bg-card border border-border rounded-md p-0.5 w-fit mb-6 animate-fade-in">
        {tabs.map((t) => (
          <button
            key={t.value}
            onClick={() => setTab(t.value)}
            className={`h-full px-3.5 text-xs font-medium rounded transition-all duration-150 ${
              tab === t.value
                ? "bg-primary/15 text-primary border border-primary/25"
                : "text-text-muted hover:text-text-secondary border border-transparent"
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === "members" && tenant && <MembersTab tenantId={tenant} />}
      {tab === "api-keys" && <ApiKeysTab />}
    </div>
  );
}
