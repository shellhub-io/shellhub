import { Link } from "react-router-dom";

interface ResourceNotFoundProps {
  icon: React.ComponentType<React.SVGProps<SVGSVGElement>>;
  resource: string;
  backTo: string;
}

/**
 * The screen for a resource that is gone or was never there, with the way back. Used instead of
 * a bare 404 so the user lands somewhere they can act.
 */
export default function ResourceNotFound({
  icon: Icon,
  resource,
  backTo,
}: ResourceNotFoundProps) {
  return (
    <div className="text-center py-24" role="status">
      <Icon
        className="w-10 h-10 text-text-muted/30 mx-auto mb-3"
        strokeWidth={1}
        aria-hidden="true"
      />
      <p className="text-sm text-text-muted mb-2">{resource} not found</p>
      <Link to={backTo} className="text-sm text-primary hover:underline">
        Back to {resource.toLowerCase()}s
      </Link>
    </div>
  );
}
