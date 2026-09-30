import { Link } from "react-router-dom";
import { ArrowLeftIcon } from "@heroicons/react/24/outline";
import { Button } from "@shellhub/design-system/primitives";
import FramedShell from "@/components/layout/FramedShell";
import ScreenIntro from "@/components/layout/ScreenIntro";
import AuthActions from "@/components/auth/AuthActions";

/**
 * The 404 page.
 */
export default function NotFound() {
  return (
    <FramedShell>
      <ScreenIntro
        eyebrow="404"
        title="Page not found"
        lead="The page you're looking for doesn't exist or has been moved."
      />
      <AuthActions
        primary={
          <Button
            as={Link}
            to="/dashboard"
            size="lg"
            fullWidth
            icon={<ArrowLeftIcon className="w-4 h-4" strokeWidth={2} />}
          >
            Back to Dashboard
          </Button>
        }
      />
    </FramedShell>
  );
}
