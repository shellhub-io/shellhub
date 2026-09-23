import type { SessionPrincipal as Principal } from "@/client";
import { useAdminUser } from "@/hooks/useAdminUsers";

import SessionPrincipal from "./SessionPrincipal";

/**
 * A session's principal as the instance admin sees it: a person by their email, looked up by id.
 * An API key keeps its id, because no admin route reads a namespace's keys.
 */
export default function AdminSessionPrincipal({
  principal,
}: {
  principal: Principal;
}) {
  const { data: user } = useAdminUser(
    principal.kind === "user" ? principal.id : "",
  );

  return <SessionPrincipal principal={principal} name={user?.email} />;
}
