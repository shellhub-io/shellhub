import { useLocation } from "react-router-dom";

/**
 * Shows the router's current path and query string in an output labelled URL, so a test can assert
 * where a navigation ended up by role.
 */
export function CurrentUrl() {
  const { pathname, search } = useLocation();

  return <output aria-label="URL">{pathname + search}</output>;
}
