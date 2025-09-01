import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { IAdminAuth, IAdminUpdateSAML } from "@admin/interfaces/IInstance";
import * as instanceApi from "../api/instance";

const useInstanceStore = defineStore("instance", () => {
  const authenticationSettings = ref<IAdminAuth>({
    local: {
      enabled: false,
    },
    saml: {
      enabled: false,
      auth_url: "",
      assertion_url: "",
      idp: {
        entity_id: "",
        binding: {
          post: "",
          redirect: "",
        },
        certificates: [],
      },
      sp: {
        sign_auth_requests: false,
        certificate: "",
      },
    },
  });

  const isLocalAuthEnabled = computed(() => authenticationSettings.value?.local?.enabled);
  const isSamlEnabled = computed(() => authenticationSettings.value?.saml?.enabled);

  const fetchAuthenticationSettings = async () => {
    const response = await instanceApi.getAuthenticationSettings();
    authenticationSettings.value = response.data as IAdminAuth;
  };

  const updateLocalAuthentication = async (status: boolean) => {
    await instanceApi.configureLocalAuthentication(status);
    await fetchAuthenticationSettings();
  };

  const updateSamlAuthentication = async (data: IAdminUpdateSAML) => {
    await instanceApi.configureSAMLAuthentication(data);
    await fetchAuthenticationSettings();
  };

  return {
    authenticationSettings,
    isLocalAuthEnabled,
    isSamlEnabled,
    fetchAuthenticationSettings,
    updateLocalAuthentication,
    updateSamlAuthentication,
  };
});

export default useInstanceStore;
