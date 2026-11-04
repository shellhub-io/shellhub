import { useId, useState } from "react";
import { UserGroupIcon } from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import PageHeader from "@/components/common/PageHeader";
import SegmentedTabs, {
  type SegmentedTab,
} from "@/components/common/SegmentedTabs";
import { useNavSectionTitle } from "@/components/layout/navSections";
import MembersTab from "./MembersTab";
import ApiKeysTab from "./ApiKeysTab";

type TeamTab = "members" | "api-keys";

/** The team page: members and API keys. */
export default function Team() {
  const sectionTitle = useNavSectionTitle("/team");
  const [tab, setTab] = useState<TeamTab>("members");
  const tenant = useAuthStore((s) => s.tenant);
  const baseId = useId();
  const ids = (value: TeamTab) => ({
    tabId: `${baseId}-${value}-tab`,
    panelId: `${baseId}-${value}-panel`,
  });

  const tabs: SegmentedTab<TeamTab>[] = [
    { label: "Members", value: "members", panel: ids("members") },
    { label: "API Keys", value: "api-keys", panel: ids("api-keys") },
  ];
  const selected = ids(tab);

  return (
    <div>
      <PageHeader
        icon={<UserGroupIcon className="w-6 h-6" />}
        overline={sectionTitle}
        title="Team"
        description="Manage namespace members, invitations, and API keys"
      />

      <div className="w-fit mb-6 animate-fade-in">
        <SegmentedTabs
          tabs={tabs}
          value={tab}
          onChange={setTab}
          label="Team sections"
        />
      </div>

      <div
        role="tabpanel"
        id={selected.panelId}
        aria-labelledby={selected.tabId}
        tabIndex={0}
      >
        {tab === "members" && tenant && <MembersTab tenantId={tenant} />}
        {tab === "api-keys" && <ApiKeysTab />}
      </div>
    </div>
  );
}
