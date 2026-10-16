import { ClockIcon } from "@heroicons/react/24/outline";
import PageHeader from "@/components/common/PageHeader";
import { useNavSectionTitle } from "@/components/layout/navSections";
import { useTenantRegistrations } from "@/hooks/useProvisioningKeyEvents";
import { RegistrationTimeline } from "@/pages/provisioning-keys/ProvisioningKeyTimeline";

/**
 * Every device that registered with the namespace's tenant ID, newest first, with the ones still
 * waiting ready to accept or reject and a rejected one ready to accept instead. It is the same
 * page in a legacy and a new namespace, and it reads the device API, so any member who may accept
 * devices works the queue without listing provisioning keys.
 */
export default function PendingDevices() {
  const sectionTitle = useNavSectionTitle("/devices");

  return (
    <div className="max-w-3xl">
      <PageHeader
        icon={<ClockIcon className="w-6 h-6" />}
        overline={sectionTitle}
        title="Pending devices"
        description="Devices that registered with the tenant ID, and what was decided for each"
      />

      <RegistrationTimeline
        history={useTenantRegistrations()}
        emptyMessage="No device has registered with the tenant ID yet."
      />
    </div>
  );
}
