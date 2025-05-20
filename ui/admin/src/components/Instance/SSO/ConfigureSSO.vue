<template>
  <v-dialog v-model="dialog" @click:outside="close" max-width="700">
    <v-card>
      <v-card-title class="text-h5 pb-2" data-test="dialog-title">Configure Single Sign-on</v-card-title>
      <v-container>
        <v-card-text>
          <v-checkbox
            v-model="checkbox"
            label="Use IDP Metadata URL"
            data-test="checkbox-idp-metadata"
          />
          <div v-if="checkbox" data-test="idp-metadata-section">
            <v-text-field
              v-model="IdPMetadataURL"
              label="IDP Metadata URL"
              variant="underlined"
              required
              data-test="idp-metadata-url"
            />
          </div>
          <div v-else data-test="idp-manual-section">
            <v-text-field
              v-model="acsUrl"
              label="IdP SignOn URL"
              variant="underlined"
              required
              data-test="idp-signon-url"
            />
            <v-text-field
              v-model="entityID"
              label="IdP Entity ID"
              variant="underlined"
              required
              data-test="idp-entity-id"
            />
            <v-textarea
              v-model="x509Certificate"
              label="IdP X.509 Certificate"
              variant="underlined"
              required
              data-test="idp-x509-certificate"
            />
          </div>

          <v-expansion-panels>
            <v-expansion-panel>
              <v-expansion-panel-title data-test="advanced-settings-title">Advanced Settings</v-expansion-panel-title>
              <v-expansion-panel-text>
                <v-data-table
                  :items="mappings"
                  item-value="key"
                  disable-sort
                  hide-default-footer
                  data-test="saml-mappings-table"
                >
                  <template v-slot:top>
                    <v-row cols="12">
                      <v-col cols="9">
                        <h3>SAML Mappings</h3>
                      </v-col>
                      <v-col cols="3">
                        <v-btn
                          color="primary"
                          :disabled="mappings.length >= 2"
                          @click="addMapping"
                          data-test="add-mapping-btn"
                        >
                          Add Mapping
                        </v-btn>
                      </v-col>
                    </v-row>
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
                          v-model="item.key"
                          variant="outlined"
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
                          variant="outlined"
                          data-test="saml-mapping-value"
                        />
                      </td>
                      <td>
                        <v-row>
                          <v-col align="center" class="pt-0 px-0 pb-1">
                            <v-btn color="red" elevation="0" @click="removeMapping(index)" data-test="remove-mapping-btn">
                              <v-icon>mdi-delete</v-icon>
                            </v-btn>
                          </v-col>
                        </v-row>
                      </td>
                    </tr>
                  </template>
                </v-data-table>
                <v-tooltip location="bottom" contained target="cursor" offset="-20">
                  <template v-slot:activator="{ props }">
                    <v-row v-bind="props">
                      <v-col>
                        <v-checkbox
                          class="mt-4"
                          v-model="signRequest"
                          label="Sign authorization requests"
                          hide-details
                          data-test="sign-request-checkbox"
                        />
                      </v-col>
                    </v-row>
                  </template>
                  <span>A security feature where the SP cryptographically signs authentication
                    requests sent to the IdP. You must upload the generated certificate to your
                    IdP when enabling this.
                  </span>
                </v-tooltip>
              </v-expansion-panel-text>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-card-text>
      </v-container>
      <v-card-actions>
        <v-btn @click="close()" data-test="close-btn">
          Close
        </v-btn>
        <v-spacer />
        <v-btn :disabled="hasError" @click="updateSAMLConfiguration" data-test="save-btn">
          Save Configuration
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import useInstanceStore from "@admin/store/modules/instance";
import useSnackbar from "@/helpers/snackbar";

const checkbox = ref(false);
const signRequest = ref(false);
const dialog = defineModel({ default: false });
const snackbar = useSnackbar();
const instanceStore = useInstanceStore();

const IdPMetadataURL = ref("");
const acsUrl = ref("");
const entityID = ref("");
const x509Certificate = ref("");

const mappings = ref([{ key: "", value: "" }]);

const tableHeaders = ref([
  { text: "Attribute Key", value: "key", align: "center" },
  { text: "SAML Attribute Value", value: "value", align: "center" },
  { text: "Actions", align: "center" },
]);

const availableKeys = ["Email", "Name"];

const usedKeys = computed(() => mappings.value.map((item) => item.key).filter(Boolean));

const getSelectableKeys = (index: number) => {
  const currentKey = mappings.value[index]?.key;
  return availableKeys.filter((key) => !usedKeys.value.includes(key) || key === currentKey);
};

const addMapping = () => {
  if (mappings.value.length < 2) {
    mappings.value.push({ key: "", value: "" });
  }
};

const removeMapping = (index: number) => {
  mappings.value.splice(index, 1);
};

const resetFields = () => {
  checkbox.value = false;
  signRequest.value = false;
  IdPMetadataURL.value = "";
  acsUrl.value = "";
  entityID.value = "";
  x509Certificate.value = "";
  mappings.value = [{ key: "", value: "" }];
};

const close = () => {
  dialog.value = false;
  resetFields();
};

const hasError = computed(() => {
  if (checkbox.value) {
    return IdPMetadataURL.value.trim() === "";
  }
  return (
    acsUrl.value.trim() === ""
    || entityID.value.trim() === ""
    || x509Certificate.value.trim() === ""
    || mappings.value.some((mapping) => !mapping.key || mapping.value.trim() === "")
  );
});

const updateSAMLConfiguration = async () => {
  const mappingObject: { email: string; name: string } = {
    email: "",
    name: "",
  };

  mappings.value.forEach((item) => {
    const key = item.key.toLowerCase();
    if (key === "email" || key === "name") {
      mappingObject[key] = item.value;
    }
  });

  const data = checkbox.value
    ? {
      enable: true,
      idp: {
        metadata_url: IdPMetadataURL.value,
        mappings: mappingObject,
      },
      sp: { sign_requests: signRequest.value },
    }
    : {
      enable: true,
      idp: {
        entity_id: entityID.value,
        signon_url: acsUrl.value,
        certificate: x509Certificate.value,
        mappings: mappingObject,
      },
      sp: { sign_requests: signRequest.value },
    };

  try {
    await instanceStore.updateSamlAuthentication(data);
    snackbar.showSuccess("Successfully updated SAML configuration.");
    dialog.value = false;
  } catch {
    snackbar.showError("Failed to update SAML configuration.");
  }
};

defineExpose({ IdPMetadataURL, checkbox, mappings, dialog });
</script>
