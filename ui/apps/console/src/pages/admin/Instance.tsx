import { useState } from "react";
import { ServerIcon, UserPlusIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import { isEnterprise } from "@/env";
import { useAdminStats } from "@/hooks/useAdminStats";
import { useAdminNamespaces } from "@/hooks/useAdminNamespaces";
import PageHeader from "@/components/common/PageHeader";
import CreateUserModal from "./users/CreateUserModal";
import NeedsAttention from "./instance/NeedsAttention";
import LicensePanel from "./instance/LicensePanel";
import AccessPanel from "./instance/AccessPanel";
import ServerPanel from "./instance/ServerPanel";
import InstanceNumbers from "./instance/InstanceNumbers";
import { Count, Row } from "./instance/Panel";

function EnterpriseInstance() {
  const {
    stats,
    isLoading: statsLoading,
    isError: statsError,
  } = useAdminStats();
  const namespaces = useAdminNamespaces({ perPage: 1 });

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-5 items-start">
      <div className="lg:col-span-2 space-y-5">
        <NeedsAttention />
        <LicensePanel />
      </div>
      <div className="space-y-5">
        <AccessPanel>
          <Row label="Users" to="/admin/users">
            <Count
              value={stats?.registered_users}
              isLoading={statsLoading}
              isError={statsError}
            />
          </Row>
          <Row label="Namespaces" to="/admin/namespaces">
            <Count
              value={namespaces.totalCount}
              isLoading={namespaces.isLoading}
              isError={namespaces.isError}
            />
          </Row>
        </AccessPanel>
        <ServerPanel />
      </div>
    </div>
  );
}

function CloudInstance() {
  return (
    <div className="space-y-5">
      <InstanceNumbers />
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-5 items-start">
        <AccessPanel />
        <ServerPanel />
      </div>
    </div>
  );
}

/**
 * The instance admin's starting point. Enterprise leads with what waits on the admin and the
 * licence; Cloud, where people sign themselves up, leads with the instance's figures.
 */
export default function AdminInstance() {
  const [createUserOpen, setCreateUserOpen] = useState(false);
  const enterprise = isEnterprise();

  return (
    <div>
      <PageHeader
        icon={<ServerIcon className="w-6 h-6" />}
        title="Instance"
        description={
          enterprise
            ? "What's waiting on you, and what this instance runs."
            : "How much this instance carries, and what it runs."
        }
      >
        <Button
          onClick={() => setCreateUserOpen(true)}
          icon={<UserPlusIcon className="w-4 h-4" />}
        >
          Create user
        </Button>
      </PageHeader>

      {enterprise ? <EnterpriseInstance /> : <CloudInstance />}

      <CreateUserModal
        open={createUserOpen}
        onClose={() => setCreateUserOpen(false)}
      />
    </div>
  );
}
