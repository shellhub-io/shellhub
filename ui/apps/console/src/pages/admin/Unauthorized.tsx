import { useNavigate } from "react-router-dom";
import {
  ShieldExclamationIcon,
  ArrowLeftIcon,
  ArrowRightStartOnRectangleIcon,
  HomeIcon,
  UserGroupIcon,
  CpuChipIcon,
} from "@heroicons/react/24/outline";
import { useAuthStore } from "@/stores/authStore";
import EmptyState, {
  type EmptyStateFeature,
} from "@/components/common/EmptyState";
import { Button } from "@shellhub/design-system/primitives";

const highlights: EmptyStateFeature[] = [
  {
    icon: <HomeIcon className="w-5 h-5" />,
    title: "Main Application",
    description:
      "Return to the main ShellHub application to manage your resources.",
  },
  {
    icon: <UserGroupIcon className="w-5 h-5" />,
    title: "Request Access",
    description:
      "Contact your system administrator to request admin privileges.",
  },
  {
    icon: <CpuChipIcon className="w-5 h-5" />,
    title: "Your Workspace",
    description:
      "Manage your devices, sessions, and namespaces in the main app.",
  },
];

export default function AdminUnauthorized() {
  const logout = useAuthStore((s) => s.logout);
  const navigate = useNavigate();

  const handleLogout = () => {
    logout();
    void navigate("/login");
  };

  return (
    <EmptyState
      accent="yellow"
      icon={<ShieldExclamationIcon className="w-8 h-8" />}
      overline="Access Restricted"
      title="Admin Access Required"
      description="You don't have administrator privileges to access the Admin Console. This area is restricted to system administrators only."
      features={highlights}
      footnote="If you believe you should have admin access, contact your system administrator."
    >
      <div className="flex flex-col sm:flex-row items-center justify-center gap-3">
        <Button
          size="lg"
          icon={<ArrowLeftIcon className="w-4 h-4" aria-hidden="true" />}
          onClick={() => void navigate("/dashboard")}
        >
          Go to ShellHub
        </Button>
        <Button
          size="lg"
          variant="outline"
          icon={
            <ArrowRightStartOnRectangleIcon
              className="w-4 h-4"
              aria-hidden="true"
            />
          }
          onClick={handleLogout}
        >
          Logout
        </Button>
      </div>
    </EmptyState>
  );
}
