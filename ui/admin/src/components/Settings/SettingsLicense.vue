<template>
  <v-alert
    v-if="!installedLicense"
    class="mt-4 pl-4 pr-4 d-flex justify-center align-center"
    variant="outlined"
    type="info"
  >
    You do not have an installed license
  </v-alert>
  <v-alert
    v-else-if="license.about_to_expire"
    class="mt-4 pl-4 pr-4 d-flex justify-center justify-center align-center"
    variant="outlined"
    type="info"
  >
    Your license is about to expired
  </v-alert>
  <v-alert
    v-else-if="license && license.expired && license.grace_period"
    class="mt-4 pl-4 pr-4 d-flex justify-center justify-center align-center"
    variant="outlined"
    type="warning"
  >
    You are in grace period, your license has expired
  </v-alert>
  <v-alert
    v-else-if="license && license.expired && !license.grace_period"
    class="mt-4 pl-4 pr-4 d-flex justify-center justify-center align-center"
    variant="outlined"
    type="error"
  >
    Your license has expired
  </v-alert>
  <div class="pb-2">
    <h1>License Details</h1>
  </div>
  <v-card>
    <v-container fluid data-test="license-container">

      <div v-if="installedLicense">
        <v-row>
          <v-col md="auto">
            <v-card tile :elevation="0"> Issued at </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card tile :elevation="0">
              <v-chip data-test="issuedAt-field">
                {{ formatDetailsNow(license.issued_at) }}
              </v-chip>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="auto">
            <v-card tile :elevation="0"> Starts at </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card tile :elevation="0">
              <v-chip data-test="issuedAt-field">
                {{ formatDetailsNow(license.starts_at) }}
              </v-chip>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="auto">
            <v-card tile :elevation="0"> Expires at </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card tile :elevation="0">
              <v-chip data-test="issuedAt-field">
                {{ formatDetailsNow(license.expires_at) }}
              </v-chip>
            </v-card>
          </v-col>
        </v-row>

        <v-row>
          <v-col md="auto">
            <v-card tile :elevation="0"> Allowed at </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card data-test="allowedRegions-field" tile :elevation="0">
              <v-chip v-if="licenseIsGlobal()">
                <v-icon left class="mr-2"> mdi-earth </v-icon>
                Global
              </v-chip>
              <v-chip v-else>
                <v-icon left> mdi-flag </v-icon>
                Limited ({{ license.allowed_regions.join(", ") }})
              </v-chip>
            </v-card>
          </v-col>
        </v-row>

        <v-divider class="mb-4 mt-4" />

        <h6 class="text-h6 text-center">License Owner</h6>

        <v-row v-for="(value, name) in license.customer" :key="name" :data-test="name">
          <v-col md="auto">
            <v-card tile :elevation="0">
              {{ formatName(`${name}`) }}
            </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card tile :elevation="0">
              <div v-if="isId(`${name}`)">
                <v-chip>
                  <v-tooltip anchor="top">
                    <template v-slot:activator="{ props }">
                      <span
                        v-bind="props"
                        @click="copyText(value)"
                        @keypress="copyText(value)"
                        class="hover-text"
                      >
                        {{ value }}
                      </span>
                    </template>
                    <span>Copy ID</span>
                  </v-tooltip>
                </v-chip>
              </div>

              <div v-else>
                {{ value }}
              </div>
            </v-card>
          </v-col>
        </v-row>

        <v-divider class="mb-4 mt-4" />
        <h6 class="text-h6 text-center">Features</h6>

        <v-row
          v-for="(value, name) in removeField(license.features)"
          :key="name"
          :data-test="name"
        >
          <v-col md="auto">
            <v-card tile :elevation="0">
              {{ formatName(`${name}`) }}
            </v-card>
          </v-col>

          <v-spacer />

          <v-col md="auto" class="ml-auto">
            <v-card tile :elevation="0">
              <div v-if="isBooleanType(value)">
                <v-spacer />
                <v-icon v-if="value" color="success" data-test="sucess-icon">
                  mdi-check-circle
                </v-icon>
                <v-icon v-else color="#E53935" data-test="error-icon">
                  mdi-close-circle
                </v-icon>
              </div>
              <div v-else>
                <v-chip>
                  {{ formatFeatureValue(value) }}
                </v-chip>
              </div>
            </v-card>
          </v-col>
        </v-row>

        <v-divider class="mb-4 mt-4" />
        <h6 class="text-h6 text-center">License Field</h6>

      </div>
      <v-file-input
        class="mt-4"
        accept=".dat"
        show-size
        variant="outlined"
        label="Select license file"
        counter
        v-model="currentFile"
        @change="licenseUploadStatus = !!currentFile"
        :rules="rules"
      />
      <v-btn class="mr-2" variant="outlined" @click="uploadLicense"> Upload </v-btn>
    </v-container>

  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import moment from "moment";
import useLicenseStore from "@admin/store/modules/license";
import useSnackbar from "@/helpers/snackbar";
import { Features } from "../../interfaces/ILicense";
import handleError from "@/utils/handleError";

const currentFile = ref<File | null>(null);
const licenseUploadStatus = ref(false);
const snackbar = useSnackbar();
const licenseStore = useLicenseStore();

onMounted(async () => {
  try {
    await licenseStore.get();
  } catch {
    snackbar.showError("Error loading license.");
  }
});

const license = computed(() => licenseStore.getLicense);

const installedLicense = computed(() => license.value
    && license.value.grace_period !== undefined);

const licenseIsGlobal = () => license.value.allowed_regions.length === 0;

const formatDetailsNow = (value: string | number) => value === -1 ? "now" : moment.unix(+value).format("LL");

const isId = (name: string) => name === "id";

const formatName = (name: string) => name.charAt(0).toUpperCase() + name.slice(1).replace(/_/g, " ");

const copyText = (value: string | undefined) => {
  if (value) {
    navigator.clipboard.writeText(value);
    snackbar.showInfo("Tenant ID copied to clipboard.");
  }
};

const removeField = (jsonLicense: Features) => {
  const jsonLicenseChanged = jsonLicense;

  if (jsonLicenseChanged !== undefined) {
    delete jsonLicenseChanged.login_link;
    delete jsonLicenseChanged.reports;
  }
  return jsonLicenseChanged;
};

const formatFeatureValue = (value: number | boolean | undefined) => value === -1 ? "unlimited" : value;

const isBooleanType = (value: number | boolean | undefined) => typeof value === "boolean";

const validateLicenseFile = (file: File | null): string | boolean => {
  if (!file) return true;
  if (file.size >= 32 * 1024) return "License size must be less than 32 KB!";
  if (!file.name.endsWith(".dat")) return "Only .dat files are allowed!";
  return true;
};

const rules = [validateLicenseFile];

const uploadLicense = async () => {
  if (currentFile.value) {
    try {
      await licenseStore.post(currentFile.value);
      await licenseStore.get();
      snackbar.showSuccess("License uploaded successfully.");
      licenseUploadStatus.value = false;
    } catch (error) {
      handleError(error);
      snackbar.showError("Failed to upload the license.");
    }
  }
};

defineExpose({ license });
</script>

<style scoped>
.hover-text {
  cursor: pointer;
  animation: fadeIn 0.5s;
}

.hover-text:hover {
  text-decoration: underline;
}
</style>
