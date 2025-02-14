import { adminApi } from "./../../api/http";
import { IConfigureSAML } from "../../interfaces/IInstance";

const getAuthenticationSettings = async () => adminApi.getAuthenticationSettings();

const configureLocalAuthentication = async (
  status: boolean,
) => adminApi.configureLocalAuthentication({ enable: status });

const configureSAMLAuthentication = async (
  data: IConfigureSAML,
) => adminApi.configureSAMLAuthentication(data);

export { getAuthenticationSettings, configureLocalAuthentication, configureSAMLAuthentication };
