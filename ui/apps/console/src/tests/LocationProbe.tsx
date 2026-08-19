import { useLocation } from "react-router-dom";

export function LocationProbe({
  onLocation,
}: {
  onLocation: (search: string) => void;
}) {
  const loc = useLocation();
  onLocation(loc.search);
  return null;
}
