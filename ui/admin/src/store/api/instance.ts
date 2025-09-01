import { adminApi } from "@admin/api/http";
import { IAdminUpdateSAML } from "@admin/interfaces/IInstance";

const getAuthenticationSettings = async () => adminApi.getAuthenticationSettings();

const configureLocalAuthentication = async (
  status: boolean,
) => adminApi.configureLocalAuthentication({ enable: status });

const configureSAMLAuthentication = async (
  data: IAdminUpdateSAML,
) => adminApi.configureSAMLAuthentication(data);

export { getAuthenticationSettings, configureLocalAuthentication, configureSAMLAuthentication };
