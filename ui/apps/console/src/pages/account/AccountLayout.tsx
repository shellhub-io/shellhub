import { useEffect } from "react";
import { useAuthStore } from "@/stores/authStore";
import SectionedLayout from "@/components/settings/SectionedLayout";

import {
  UserIcon,
  ExclamationTriangleIcon,
  ShieldCheckIcon,
  SwatchIcon,
} from "@heroicons/react/24/outline";

const ACCOUNT_SECTIONS = [
  { to: "profile", label: "Profile", icon: UserIcon },
  { to: "security", label: "Security", icon: ShieldCheckIcon },
  { to: "appearance", label: "Appearance", icon: SwatchIcon },
  { to: "danger-zone", label: "Danger zone", icon: ExclamationTriangleIcon },
];

/**
 * The user's own account, one section per URL under /account: who they are, how they sign in, and
 * deleting the account.
 */
export default function AccountLayout() {
  const fetchUser = useAuthStore((s) => s.fetchUser);

  useEffect(() => {
    void fetchUser();
  }, [fetchUser]);

  return (
    <SectionedLayout
      base="/account"
      icon={<UserIcon className="w-6 h-6" />}
      title="Account"
      description="Your profile and how you sign in"
      sections={ACCOUNT_SECTIONS}
    />
  );
}
