import { adminApi } from "@admin/api/http";
import { IAdminSAMLConfig } from "@admin/interfaces/IInstance";

const getAuthenticationSettings = async () => adminApi.getAuthenticationSettings();

const configureLocalAuthentication = async (
  status: boolean,
) => adminApi.configureLocalAuthentication({ enable: status });

const configureSAMLAuthentication = async (
  data: IAdminSAMLConfig,
) => adminApi.configureSAMLAuthentication(data);

export { getAuthenticationSettings, configureLocalAuthentication, configureSAMLAuthentication };
