import { useLocation } from "react-router-dom";

/**
 * Reports the router's current query string to a test. Rendered inside a route so a case can
 * assert where a navigation ended up without reaching into the router.
 */
export function LocationProbe({
  onLocation,
}: {
  onLocation: (search: string) => void;
}) {
  const loc = useLocation();
  onLocation(loc.search);
  return null;
}
