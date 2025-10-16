<template>
  <ConfigureSSO v-model="showSSODialog" data-test="configure-sso-dialog" />

  <div class="pb-2">
    <h1 data-test="auth-header">Authentication</h1>
  </div>
  <v-card class="w-100 pa-4 bg-background border" data-test="auth-card">
    <v-card-item class="pa-0">
      <v-card-title class="text-center" data-test="auth-status-header">Authentication Status</v-card-title>
      <v-row class="my-0">
        <span data-test="local-auth-label">Local Authentication</span>
        <v-switch
          v-model="isLocalAuthEnabled"
          @click.prevent="changeLocalAuthStatus"
          :color="isLocalAuthEnabled ? 'success' : 'error'"
          data-test="local-auth-switch"
          hide-details
        />
      </v-row>
      <v-row class="my-0">
        <span data-test="saml-auth-label">SAML Authentication</span>
        <v-switch
          v-model="isSamlEnabled"
          @click.prevent="changeSamlAuthStatus"
          :color="isSamlEnabled ? 'success' : 'error'"
          data-test="saml-auth-switch"
          hide-details
        />
      </v-row>
    </v-card-item>

    <v-card-item v-if="isSamlEnabled" class="pa-0">
      <v-divider class="mt-4 mb-4" />
      <v-card-title class="text-center" data-test="sso-header">Single Sign-on (SSO)</v-card-title>
      <v-card-subtitle
        class="text-center"
        data-test="sso-subtitle"
      >
        Single Sign-On (SSO) simplifies access by enabling users to authenticate
        once and securely access multiple applications.
      </v-card-subtitle>

      <v-row>
        <div class="d-flex flex-column w-75">
          <span>Assertion URL</span>
          <span v-if="smAndUp" class="text-subtitle-2 text-medium-emphasis text-truncate font-weight-regular">
            The Assertion URL is the endpoint where the IdP will redirect users after
            successful authentication. Many IdPs require this URL to be registered in
            their list of allowed callback URLs.
          </span>
        </div>
        <v-btn
          @click="copyAssertionURL(ssoSettings.saml?.assertion_url)"
          data-test="copy-assertion-btn"
        >
          Copy Assertion URL
        </v-btn>
      </v-row>

      <v-row v-if="'post' in binding && binding?.post">
        <span data-test="idp-signon-post-label">IdP SignOn POST URL</span>
        <span data-test="idp-signon-post-value">
          {{ binding?.post }}
        </span>
      </v-row>

      <v-row v-if="'redirect' in binding && binding?.redirect">
        <span data-test="idp-signon-redirect-label">IdP SignOn Redirect URL</span>
        <span data-test="idp-signon-redirect-value">
          {{ binding?.redirect }}
        </span>
      </v-row>

      <v-row>
        <span data-test="idp-entity-label">IdP Entity ID</span>
        <span data-test="idp-entity-value">
          {{ ssoSettings.saml?.idp.entity_id }}
        </span>
      </v-row>

      <v-row v-if="certificate">
        <div class="d-flex flex-column w-75">

          <span data-test="certificate-label">SP Certificate</span>
          <span v-if="smAndUp" class="text-subtitle-2 text-medium-emphasis text-truncate font-weight-regular">
            The SP Certificate is an X.509 certificate that IdPs use to verify requests
            from ShellHub. Upload it to your IdP to validate the authenticity of authentication
            requests.
          </span>
        </div>
        <v-btn
          class="align-content-lg-center text-none text-uppercase"
          @click="downloadSSOCertificate"
          data-test="download-certificate-btn"
        >Download SP Certificate</v-btn>
      </v-row>

      <v-card-actions class="justify-end pa-0 mt-4">
        <v-tooltip location="top">
          <template v-slot:activator="{ props }">
            <v-btn
              v-bind="props"
              @click="redirectToAuthURL(ssoSettings.saml?.auth_url)"
              data-test="redirect-auth-btn"
            >
              Test Auth Integration
            </v-btn>
          </template>
          <span>Opens a new window directly calling the Authentication URL</span>
        </v-tooltip>
        <v-btn @click="showSSODialog = true" data-test="sso-config-btn">{{ ssoSettings.saml?.enabled ? "Edit" : "Configure" }}</v-btn>
      </v-card-actions>
    </v-card-item>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { AxiosError } from "axios";
import { useDisplay } from "vuetify";
import useInstanceStore from "@admin/store/modules/instance";
import useSnackbar from "@/helpers/snackbar";
import ConfigureSSO from "../Instance/SSO/ConfigureSSO.vue";

const showSSODialog = ref(false);
const snackbar = useSnackbar();
const instanceStore = useInstanceStore();
const { smAndUp } = useDisplay();
const ssoSettings = computed(() => instanceStore.authenticationSettings);
const certificate = computed(() => ssoSettings.value.saml?.sp?.certificate);
const binding = computed(() => ssoSettings.value.saml?.idp.binding);
const isLocalAuthEnabled = computed(() => instanceStore.isLocalAuthEnabled);
const isSamlEnabled = computed(() => instanceStore.isSamlEnabled);

const disableSaml = async () => {
  await instanceStore.updateSamlAuthentication({
    enable: false,
    idp: {
      entity_id: "",
      binding: { post: "", redirect: "" },
      certificate: "",
    },
    sp: { sign_requests: false },
  });
};

const handleAuthUpdateError = (error: unknown) => {
  if ((error as AxiosError).status === 400) snackbar.showError("You cannot disable all authentication methods.");
  else snackbar.showError("An error occurred while updating local authentication.");
};

const changeSamlAuthStatus = async () => {
  try {
    if (isSamlEnabled.value) await disableSaml();
    else showSSODialog.value = true;
  } catch (error: unknown) { handleAuthUpdateError(error); }
};

const changeLocalAuthStatus = async () => {
  try {
    await instanceStore.updateLocalAuthentication(!isLocalAuthEnabled.value);
  } catch (error: unknown) { handleAuthUpdateError(error); }
};

const downloadSSOCertificate = () => {
  if (!certificate.value) {
    snackbar.showError("No certificates available to download.");
    return;
  }

  const fileName = "saml_shellhub_certificate.pem";

  const blob = new Blob([certificate.value], { type: "text/plain" });
  const url = URL.createObjectURL(blob);

  const a = document.createElement("a");
  a.href = url;
  a.download = fileName;
  a.click();

  URL.revokeObjectURL(url);
};

const redirectToAuthURL = (url?: string) => { window.open(url, "_blank"); };

const copyAssertionURL = (url?: string) => {
  if (url) {
    navigator.clipboard.writeText(url);
    snackbar.showInfo("Authentication URL copied to clipboard.");
  }
};

onMounted(async () => { await instanceStore.fetchAuthenticationSettings(); });

defineExpose({ showSSODialog, isSamlEnabled, certificate });
</script>

<style lang="scss" scoped>
.v-row {
  margin: 1.25rem 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
