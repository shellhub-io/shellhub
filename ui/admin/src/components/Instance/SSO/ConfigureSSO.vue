<template>
  <BaseDialog v-model="showDialog" @close="close">
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
            class="mb-2"
            variant="underlined"
            :required="!redirectUrl"
            data-test="idp-signon-post-url"
          />
          <v-text-field
            v-model="redirectUrl"
            :error-messages="redirectUrlError"
            label="IdP SignOn Redirect URL"
            hint="SAML sign-on URL from your IdP console or metadata (HTTP-Redirect binding)"
            class="mb-2"
            variant="underlined"
            :required="!postUrl"
            data-test="idp-signon-redirect-url"
          />
          <v-text-field
            v-model="entityID"
            label="IdP Entity ID"
            hint="Issuer/Entity ID from your IdP's SAML configuration"
            variant="underlined"
            class="mb-2"
            required
            data-test="idp-entity-id"
          />
          <v-textarea
            :model-value="x509Certificate"
            @update:model-value="handleCertificateChange"
            label="IdP X.509 Certificate"
            hint="Public certificate used by IdP to sign SAML responses. Found in IdP console or metadata"
            class="mb-3"
            variant="underlined"
            required
            data-test="idp-x509-certificate"
            :error-messages="x509CertificateErrorMessage"
          />
        </div>
        <v-expansion-panels elevation="0">
          <v-expansion-panel>
            <v-expansion-panel-title data-test="advanced-settings-title">Advanced Settings</v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-data-table
                :items="mappings"
                item-value="key"
                density="compact"
                disable-sort
                hide-default-footer
                data-test="saml-mappings-table"
              >
                <template v-slot:top>
                  <v-row class="justify-space-between align-center mb-3">
                    <h3>SAML Mappings</h3>
                    <v-btn
                      color="primary"
                      :disabled="mappings.length >= 2"
                      @click="addMapping"
                      data-test="add-mapping-btn"
                    >
                      Add Mapping
                    </v-btn>
                  </v-row>
                  <p class="mb-3">Maps SAML attributes to user fields.</p>
                </template>
                <template v-slot:headers>
                  <tr>
                    <th v-for="(header, i) in tableHeaders" :key="i" :class="`text-${header.align}`">
                      <span>{{ header.text }}</span>
                    </th>
                  </tr>
                </template>
                <template v-slot:item="{ item, index }">
                  <tr>
                    <td>
                      <v-select
                        :items="getSelectableKeys(index)"
                        hide-details
                        :model-value="item.key"
                        @update:model-value="(newKey) => handleKeyChange(index, newKey)"
                        variant="outlined"
                        density="compact"
                        placeholder="Select Key"
                        :menu-props="{ closeOnContentClick: false }"
                        data-test="saml-mapping-key"
                      />
                    </td>
                    <td>
                      <v-text-field
                        class="py-4"
                        hide-details
                        v-model="item.value"
                        placeholder="Value"
                        density="compact"
                        variant="outlined"
                        data-test="saml-mapping-value"
                      />
                    </td>
                    <td class="text-center">
                      <v-btn
                        color="red"
                        icon="mdi-delete"
                        size="small"
                        elevation="0"
                        @click="removeMapping(index)"
                        data-test="remove-mapping-btn"
                      />
                    </td>
                  </tr>
                </template>
              </v-data-table>
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
  </BaseDialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import useInstanceStore from "@admin/store/modules/instance";
import { useField } from "vee-validate";
import * as yup from "yup";
import { IAdminUpdateSAML } from "@admin/interfaces/IInstance";
import useSnackbar from "@/helpers/snackbar";
import { isX509CertificateValid } from "@/utils/sshKeys";
import BaseDialog from "@/components/Dialogs/BaseDialog.vue";

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

const mappings = ref<{ key: string; value: string }[]>([]);

const tableHeaders = ref([
  { text: "Attribute Key", value: "key", align: "center" },
  { text: "SAML Attribute Value", value: "value", align: "center" },
  { text: "Actions", align: "center" },
]);

const mappingFields = {
  Email: "emailAddress",
  Name: "displayName",
};

const usedKeys = computed(() => mappings.value.map((item) => item.key).filter(Boolean));

const getSelectableKeys = (index: number) => {
  const currentKey = mappings.value[index]?.key;
  return Object.keys(mappingFields).filter((key) => !usedKeys.value.includes(key) || key === currentKey);
};

const addMapping = () => {
  if (mappings.value.length < 2) {
    const availableKey = Object.keys(mappingFields).find((key) => !usedKeys.value.includes(key));
    const defaultValue = mappingFields[availableKey || ""];

    mappings.value.push({
      key: availableKey || "",
      value: defaultValue || "",
    });
  }
};

const removeMapping = (index: number) => { mappings.value.splice(index, 1); };

// Update value to default when key (Email/Name) changes
const handleKeyChange = (index: number, newKey: string) => {
  mappings.value[index].key = newKey;
  mappings.value[index].value = mappingFields[newKey] || "";
};

const resetFields = () => {
  useMetadataUrl.value = false;
  signRequest.value = false;
  IdPMetadataURL.value = "";
  postUrl.value = "";
  redirectUrl.value = "";
  entityID.value = "";
  x509Certificate.value = "";
  mappings.value = [];
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

  // Mapping validation
  if (
    mappings.value.length > 0
    && mappings.value.some((mapping) => !mapping.key || !mapping.value.trim())
  ) {
    return true;
  }

  return false;
});

const normalizeCertificate = (c: string) => c.replace(
  /(-----BEGIN CERTIFICATE-----)[\s\S]*?(-----END CERTIFICATE-----)/,
  (match, beginHeader, endHeader) => {
    const content = match.slice(beginHeader.length, -endHeader.length);
    const normalizedContent = content.replace(/\s+/g, "\n").replace(/\n+/g, "\n").trim();
    return `${beginHeader}\n${normalizedContent}\n${endHeader}`;
  },
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

  const validMappings = mappings.value.filter(
    (m) => m.key && m.value.trim(),
  );

  if (validMappings.length > 0) {
    const mappingObject: Partial<Record<"email" | "name", string>> = {};

    validMappings.forEach(({ key, value }) => {
      const lowerKey = key.toLowerCase() as "email" | "name";
      mappingObject[lowerKey] = value;
    });

    if (Object.keys(mappingObject).length > 0) {
      idpConfig.mappings = mappingObject as { email: string; name: string };
    }
  }

  const data: IAdminUpdateSAML = {
    enable: true,
    idp: idpConfig,
    sp: { sign_requests: signRequest.value },
  };

  try {
    await instanceStore.updateSamlAuthentication(data);
    snackbar.showSuccess("Successfully updated SAML configuration.");
    close();
  } catch {
    snackbar.showError("Failed to update SAML configuration.");
  }
};

const populateFields = () => {
  const { idp, sp } = instanceStore.authenticationSettings.saml;
  const { binding } = idp;
  postUrl.value = "post" in binding ? binding.post : "";
  redirectUrl.value = "redirect" in binding ? binding.redirect : "";
  entityID.value = idp.entity_id || "";
  x509Certificate.value = idp.certificates[0] || "";
  signRequest.value = !!sp.sign_auth_requests;

  if (idp.mappings) {
    const newMappings: { key: string; value: string }[] = [];
    if (idp.mappings.email) {
      newMappings.push({ key: "Email", value: idp.mappings.email });
    }
    if (idp.mappings.name) {
      newMappings.push({ key: "Name", value: idp.mappings.name });
    }
    mappings.value = newMappings;
  }
};

watch(showDialog, (newVal) => { if (newVal && instanceStore.isSamlEnabled) populateFields(); });

defineExpose({ IdPMetadataURL, useMetadataUrl, showDialog, handleCertificateChange, x509CertificateErrorMessage, mappings });
</script>
