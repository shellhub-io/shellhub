<template>
  <BaseDialog v-model="showDialog" @close="close">
    <v-card>
      <v-card-title class="text-h5 pb-2" data-test="dialog-title">Configure Single Sign-on</v-card-title>
      <v-card-text>
        <v-form>
          <v-checkbox
            v-model="useMetadataUrl"
            label="Use IDP Metadata URL"
            data-test="checkbox-idp-metadata"
            hide-details
          />
          <v-text-field
            v-model="IdPMetadataURL"
            density="compact"
            v-if="useMetadataUrl"
            :error-messages="IdPMetadataURLError"
            class="mb-4 pt-0"
            label="IDP Metadata URL"
            hint="Found in your identity provider's SAML app settings. Alternative to manual configuration"
            persistent-hint
            variant="underlined"
            required
            data-test="idp-metadata-url"
          />
          <div v-else data-test="idp-manual-section">
            <v-alert
              type="warning"
              class="mb-4"
              data-test="manual-config-info"
              v-if="!isAtLeastOneUrlValid()"
            >
              You need to provide at least one of the following URLs: POST URL or Redirect URL.
            </v-alert>
            <v-text-field
              v-model="postUrl"
              :error-messages="postUrlError"
              label="IdP SignOn POST URL"
              hint="SAML sign-on URL from your IdP console or metadata (HTTP-POST binding)"
              variant="underlined"
              :required="!redirectUrl"
              data-test="idp-signon-post-url"
            />
            <v-text-field
              v-model="redirectUrl"
              :error-messages="redirectUrlError"
              label="IdP SignOn Redirect URL"
              hint="SAML sign-on URL from your IdP console or metadata (HTTP-Redirect binding)"
              variant="underlined"
              :required="!postUrl"
              data-test="idp-signon-redirect-url"
            />
            <v-text-field
              v-model="entityID"
              label="IdP Entity ID"
              hint="Issuer/Entity ID from your IdP's SAML configuration"
              variant="underlined"
              required
              data-test="idp-entity-id"
            />
            <v-textarea
              :model-value="x509Certificate"
              @update:model-value="handleCertificateChange"
              label="IdP X.509 Certificate"
              hint="Public certificate used by IdP to sign SAML responses. Found in IdP console or metadata"
              variant="underlined"
              required
              data-test="idp-x509-certificate"
              :error-messages="x509CertificateErrorMessage"
            />
          </div>
          <v-expansion-panels elevation="0">
            <v-expansion-panel title="Advanced Settings">
              <v-expansion-panel-text class="pt-4">
                <h3>SAML Mappings</h3>
                <p class="mt-1 mb-4 text-justify">
                  Maps SAML attributes to user fields.
                  Enable to change the values, or leave disabled to use defaults.
                </p>
                <div class="d-flex justify-start align-center mb-4 ga-4">
                  <v-checkbox-btn
                    v-model="enableEmailMapping"
                    @update:model-value="(val: boolean) => { if (!val) resetEmailMapping(); }"
                    density="compact"
                    data-test="enable-email-mapping-checkbox"
                  />
                  <v-text-field
                    :disabled="!enableEmailMapping"
                    v-model.trim="emailMappingValue"
                    class="w-100"
                    label="SAML Email Attribute"
                    hint="The SAML attribute name that contains the user's email"
                    variant="outlined"
                    hide-details
                    data-test="email-mapping-field"
                  />
                </div>
                <div class="d-flex justify-start align-center mb-2 ga-4">
                  <v-checkbox-btn
                    v-model="enableNameMapping"
                    @update:model-value="(val: boolean) => { if (!val) resetNameMapping(); }"
                    density="compact"
                    data-test="enable-name-mapping-checkbox"
                  />
                  <v-text-field
                    :disabled="!enableNameMapping"
                    v-model.trim="nameMappingValue"
                    class="w-100"
                    label="SAML Name Attribute"
                    hint="The SAML attribute name that contains the user's display name"
                    variant="outlined"
                    hide-details
                    data-test="name-mapping-field"
                  />
                </div>
                <v-tooltip location="bottom" contained offset="-10">
                  <template v-slot:activator="{ props }">
                    <v-checkbox
                      v-bind="props"
                      class="mt-4"
                      v-model="signRequest"
                      label="Sign authorization requests"
                      hide-details
                      data-test="sign-request-checkbox"
                    />
                  </template>
                  <span>Allows IdP to verify that SAML requests originated from ShellHub</span>
                </v-tooltip>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-form>
      </v-card-text>
      <v-card-actions>
        <v-btn @click="close()" data-test="close-btn">
          Close
        </v-btn>
        <v-btn :disabled="hasErrors" @click="updateSAMLConfiguration" color="primary" data-test="save-btn">
          Save Configuration
        </v-btn>
      </v-card-actions>
    </v-card>
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import useInstanceStore from "@admin/store/modules/instance";
import { useField } from "vee-validate";
import * as yup from "yup";
import { IAdminUpdateSAML } from "@admin/interfaces/IInstance";
import useSnackbar from "@/helpers/snackbar";
import { isX509CertificateValid } from "@/utils/sshKeys";
import BaseDialog from "@/components/BaseDialog.vue";

const useMetadataUrl = ref(false);
const signRequest = ref(false);
const showDialog = defineModel({ default: false });
const snackbar = useSnackbar();
const instanceStore = useInstanceStore();

const entityID = ref("");
const x509Certificate = ref("");
const x509CertificateErrorMessage = ref("");

const { value: IdPMetadataURL,
  errorMessage: IdPMetadataURLError,
} = useField<string>("IdPMetadataURL", yup.string().url(), { initialValue: "" });

const { value: postUrl,
  errorMessage: postUrlError,
} = useField<string>("postUrl", yup.string().url(), { initialValue: "" });

const { value: redirectUrl,
  errorMessage: redirectUrlError,
} = useField<string>("redirectUrl", yup.string().url(), { initialValue: "" });

const enableEmailMapping = ref(false);
const emailMappingValue = ref("email");
const enableNameMapping = ref(false);
const nameMappingValue = ref("username");

const resetNameMapping = () => { nameMappingValue.value = "username"; };
const resetEmailMapping = () => { emailMappingValue.value = "email"; };

const resetFields = () => {
  useMetadataUrl.value = false;
  signRequest.value = false;
  IdPMetadataURL.value = "";
  postUrl.value = "";
  redirectUrl.value = "";
  entityID.value = "";
  x509Certificate.value = "";
  enableEmailMapping.value = false;
  enableNameMapping.value = false;
  resetEmailMapping();
  resetNameMapping();
};

const close = () => {
  showDialog.value = false;
  resetFields();
};

const beginCertificate = "-----BEGIN CERTIFICATE-----";
const endCertificate = "-----END CERTIFICATE-----";
const isCertificateValid = computed(() => isX509CertificateValid(x509Certificate.value));

const handleCertificateChange = (value: string) => {
  x509Certificate.value = value.trim();

  if (!x509Certificate.value) {
    x509CertificateErrorMessage.value = "The certificate field is required.";
    return;
  }

  if (!x509Certificate.value.includes(beginCertificate) || !x509Certificate.value.includes(endCertificate)) {
    x509CertificateErrorMessage.value = `Certificate must include ${beginCertificate} and ${endCertificate} blocks.`;
    return;
  }

  if (!isCertificateValid.value) {
    x509CertificateErrorMessage.value = "Invalid X.509 certificate.";
    return;
  }

  x509CertificateErrorMessage.value = "";
};

const isAtLeastOneUrlValid = (): boolean => {
  const isPostUrlValid = postUrl.value.trim() !== "" && !postUrlError.value;
  const isRedirectUrlValid = redirectUrl.value.trim() !== "" && !redirectUrlError.value;

  return isPostUrlValid || isRedirectUrlValid;
};

const hasErrors = computed((): boolean => {
  // If using metadata URL, validate only it and stop.
  if (useMetadataUrl.value) {
    return IdPMetadataURL.value.trim() === "" || !!IdPMetadataURLError.value;
  }

  // Manual configuration checks
  if (
    !isAtLeastOneUrlValid()
    || postUrlError.value
    || redirectUrlError.value
    || entityID.value.trim() === ""
    || x509Certificate.value.trim() === ""
    || !isCertificateValid.value
  ) {
    return true;
  }

  return false;
});

// Trim linebreaks on the certificate
const normalizeCertificate = (c: string) => c.replace(
  /-----BEGIN CERTIFICATE-----[\s\S]+?-----END CERTIFICATE-----/,
  (m) => m.replace(/\s+/g, "\n").replace(/\n+/g, "\n").trim(),
);

const updateSAMLConfiguration = async (): Promise<void> => {
  const idpConfig: IAdminUpdateSAML["idp"] = useMetadataUrl.value
    ? { metadata_url: IdPMetadataURL.value }
    : {
      entity_id: entityID.value,
      binding: {
        post: postUrl.value,
        redirect: redirectUrl.value,
      },
      certificate: normalizeCertificate(x509Certificate.value),
    };

  const data: IAdminUpdateSAML = {
    enable: true,
    idp: {
      ...idpConfig,
      mappings: {
        email: emailMappingValue.value || "email",
        name: nameMappingValue.value || "username",
      },
    },
    sp: { sign_requests: signRequest.value },
  };

  try {
    await instanceStore.updateSamlAuthentication(data);
    snackbar.showSuccess("Successfully updated SAML configuration.");
    showDialog.value = false;
  } catch {
    snackbar.showError("Failed to update SAML configuration.");
  }
};

defineExpose({ IdPMetadataURL, useMetadataUrl, showDialog, handleCertificateChange, x509CertificateErrorMessage });
</script>
