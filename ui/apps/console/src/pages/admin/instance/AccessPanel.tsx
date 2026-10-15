import type { ReactNode } from "react";
import { isEnterprise } from "@/env";
import { useAdminUsers } from "@/hooks/useAdminUsers";
import { useServerInfo } from "@/hooks/useServerInfo";
import { ADMIN_AUTHENTICATION_PATH } from "@/utils/adminRoute";
import { Count, Loaded, Panel, PanelError, Row } from "./Panel";

/**
 * Who holds instance admin rights and how people sign in. children are extra rows placed above
 * those two, for an edition that shows more here.
 */
export default function AccessPanel({ children }: { children?: ReactNode }) {
  const admins = useAdminUsers({ perPage: 1, subset: "admin" });
  const info = useServerInfo();

  return (
    <Panel title="Access">
      {admins.isError && (
        <PanelError>Couldn&apos;t load the instance admins.</PanelError>
      )}
      {children}
      <Row label="Instance admins" to="/admin/users?subset=admin">
        <Count
          value={admins.totalCount}
          isLoading={admins.isLoading}
          isError={admins.isError}
        />
      </Row>
      <Row
        label="Sign-in methods"
        to={isEnterprise() ? ADMIN_AUTHENTICATION_PATH : undefined}
      >
        <Loaded isLoading={info.isLoading} isError={info.isError}>
          {info.signInMethods.length > 0
            ? info.signInMethods.join(" · ")
            : "None"}
        </Loaded>
      </Row>
    </Panel>
  );
}
