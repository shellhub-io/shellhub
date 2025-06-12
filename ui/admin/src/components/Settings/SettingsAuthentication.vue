<template>
  <ConfigureSSO v-model="dialogSSO" data-test="configure-sso-dialog" />

  <div class="pb-2">
    <h1 data-test="auth-header">Authentication</h1>
  </div>
  <v-card class="w-100" border="8" data-test="auth-card">
    <v-container fluid data-test="license-container">
      <v-row>
        <v-col>
          <h6 class="text-h6 text-center" data-test="auth-status-header">Authentication Status</h6>
        </v-col>
      </v-row>

      <v-row>
        <v-col md="auto" sm="auto">
          <v-card tile :elevation="0" data-test="local-auth-label">Local Authentication</v-card>
        </v-col>
        <v-spacer />
        <v-col md="auto" sm="auto" class="ml-auto pt-0">
          <v-switch
            v-model="localEnabled"
            @click="changeLocalAuthStatus"
            :color="localEnabled ? 'success' : 'error'"
            data-test="local-auth-switch"
            hide-details
          />
        </v-col>
      </v-row>

      <v-row>
        <v-col md="auto" sm="auto">
          <v-card tile :elevation="0" data-test="saml-auth-label">SAML Authentication</v-card>
        </v-col>
        <v-spacer />
        <v-col md="auto" sm="auto" class="ml-auto pt-0">
          <v-switch
            v-model="samlEnabled"
            @click="changeSamlAuthStatus"
            :color="samlEnabled ? 'success' : 'error'"
            data-test="saml-auth-switch"
            hide-details
          />
        </v-col>
      </v-row>

      <div v-if="samlEnabled">
        <v-row>
          <v-divider class="mt-4 mb-4" />
        </v-row>

        <v-row>
          <v-col>
            <v-card-title class="text-h6 text-center" data-test="sso-header">Single Sign-on (SSO)</v-card-title>
            <v-card-subtitle
              class="text-center"
              data-test="sso-subtitle"
            >
              Single Sign-On (SSO) simplifies access by enabling users to authenticate
              once and securely access multiple applications.
            </v-card-subtitle>
          </v-col>
        </v-row>

        <v-row cols="12" class="mt-2">
          <v-col md="10" sm="8">
            <v-card tile :elevation="0" data-test="idp-signon-label">Assertion URL</v-card>
            <v-card-subtitle class="pl-0">
              The Assertion URL is the endpoint where the IdP will redirect users after
              successful authentication. Many IdPs require this URL to be registered in
              their list of allowed callback URLs.
            </v-card-subtitle>
          </v-col>
          <v-spacer />
          <v-col md="2" sm="4" class="ml-auto d-flex w-100 justify-end align-center">
            <v-btn
              @click="copyAssertionURL(ssoSettings.saml?.assertion_url)"
              data-test="copy-assertion-btn"
            >
              Copy Assertion URL
            </v-btn>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="auto" sm="auto">
            <v-card tile :elevation="0" data-test="idp-signon-label">IdP SignOn URL</v-card>
          </v-col>
          <v-spacer />
          <v-col md="auto" sm="auto" class="ml-auto">
            <v-card tile :elevation="0" data-test="idp-signon-value">
              {{ ssoSettings.saml?.idp.signon_url }}
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="auto" sm="auto">
            <v-card tile :elevation="0" data-test="idp-entity-label">IdP Entity ID</v-card>
          </v-col>
          <v-spacer />
          <v-col md="auto" sm="auto" class="ml-auto mb-3">
            <v-card tile :elevation="0" data-test="idp-entity-value">
              {{ ssoSettings.saml?.idp.entity_id }}
            </v-card>
          </v-col>
        </v-row>

        <v-row cols=12 v-if="certificate">
          <v-col md="10" sm="8">
            <v-card tile :elevation="0" data-test="certificate-label">SP Certificate</v-card>
            <v-card-subtitle class="pl-0 text-overflow">
              The SP Certificate is an X.509 certificate that IdPs use to verify requests
              from ShellHub. Upload it to your IdP to validate the authenticity of authentication
              requests.
            </v-card-subtitle>
          </v-col>
          <v-spacer />
          <v-col md="2" sm="4" class="d-flex w-100 justify-end align-center">
            <v-btn
              class="align-content-lg-center text-none text-uppercase"
              @click="downloadSSOCertificates"
              data-test="download-certificate-btn"
            >Download SP Certificate</v-btn>
          </v-col>
        </v-row>

        <v-row cols="12">
          <v-col md="10" sm="6">
            <v-card tile :elevation="0" data-test="sso-config-label">SSO Configuration</v-card>
          </v-col>
          <v-spacer />
          <v-col md="1" sm="3" class="ml-auto d-flex w-100 justify-end align-center">
            <v-tooltip location="top center" contained target="cursor" offset="-10">
              <template v-slot:activator="{ props }">
                <v-row v-bind="props">
                  <v-col>
                    <v-btn
                      v-bind="props"
                      @click="redirectToAuthURL(ssoSettings.saml?.auth_url)"
                      data-test="redirect-auth-btn"
                    >
                      Test Auth Integration
                    </v-btn>
                  </v-col>
                </v-row>
              </template>
              <span>Opens a new window directly calling the Authentication URL
              </span>
            </v-tooltip>
          </v-col>
          <v-col md="1" sm="3" class="ml-auto d-flex w-100 justify-end align-center">
            <v-btn @click="dialogSSO = true" data-test="sso-config-btn">{{ ssoSettings.saml?.enabled ? "Edit" : "Configure" }}</v-btn>
          </v-col>
        </v-row>
      </div>
    </v-container>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import axios from "axios";
import useInstanceStore from "@admin/store/modules/instance";
import useSnackbar from "@/helpers/snackbar";
import ConfigureSSO from "../Instance/SSO/ConfigureSSO.vue";

const dialogSSO = ref(false);
const snackbar = useSnackbar();
const instanceStore = useInstanceStore();

onMounted(async () => {
  await instanceStore.fetchAuthenticationSettings();
});

const ssoSettings = computed(() => instanceStore.getAuthenticationSettings);
const certificate = computed(() => instanceStore.getAuthenticationSettings?.saml?.sp?.certificate);

const localEnabled = computed({
  get: () => instanceStore.isLocalAuthEnabled,
  set: (val: boolean) => {
    instanceStore.updateLocalAuthentication(val);
  },
});

const samlEnabled = computed({
  get: () => instanceStore.isSamlEnabled,
  set: (val: boolean) => {
    if (val === false) {
      const payload = {
        enable: false,
        idp: {
          entity_id: "",
          signon_url: "",
          certificate: "",
        },
        sp: {
          sign_requests: false,
        },
      };

      instanceStore.updateSamlAuthentication(payload);
    } else {
      dialogSSO.value = true;
    }
  },
});

const changeLocalAuthStatus = async () => {
  try {
    await instanceStore.updateLocalAuthentication(!localEnabled.value);
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      switch (error.status) {
        case 400:
          snackbar.showError("You cannot disable all authentication methods.");
          break;
        default:
          snackbar.showError("An error occurred while updating local authentication.");
          break;
      }
    }
  }
};

const changeSamlAuthStatus = async () => {
  try {
    if (samlEnabled.value) {
      const payload = {
        enable: false,
        idp: {
          entity_id: "",
          signon_url: "",
          certificate: "",
        },
        sp: {
          sign_requests: false,
        },
      };

      await instanceStore.updateSamlAuthentication(payload);
    } else {
      dialogSSO.value = true;
    }
  } catch (error: unknown) {
    if (axios.isAxiosError(error)) {
      switch (error.status) {
        case 400:
          snackbar.showError("You cannot disable all authentication methods.");
          break;
        default:
          snackbar.showError("An error occurred while updating SAML authentication.");
      }
    }
  }
};

const downloadSSOCertificates = () => {
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

const redirectToAuthURL = (value: string | undefined) => {
  window.open(value, "_blank");
};

const copyAssertionURL = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    snackbar.showInfo("Authentication URL copied to clipboard.");
  }
};

defineExpose({ dialogSSO, samlEnabled, certificate });
</script>
