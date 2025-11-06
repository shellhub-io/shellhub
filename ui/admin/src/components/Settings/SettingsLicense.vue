<template>
  <v-alert
    v-if="licenseAlert"
    class="mt-4 pl-4 pr-4 d-flex justify-center align-center"
    variant="outlined"
    :type="licenseAlert.type"
    :text="licenseAlert.message"
  />
  <h1 class="pb-2">
    License Details
  </h1>
  <v-card
    class="w-100 pa-4 bg-background border"
    data-test="license-card"
  >
    <div v-if="isLicenseInstalled">
      <v-card-item class="pa-0">
        <v-row>
          <span>Issued at</span>
          <v-chip data-test="issued-at-field">
            {{ convertValueToDate(license.issued_at) }}
          </v-chip>
        </v-row>
        <v-row>
          <span>Starts at</span>
          <v-chip data-test="starts-at-field">
            {{ convertValueToDate(license.starts_at) }}
          </v-chip>
        </v-row>
        <v-row>
          <span>Expires at</span>
          <v-chip data-test="expires-at-field">
            {{ convertValueToDate(license.expires_at) }}
          </v-chip>
        </v-row>
        <v-row>
          <span>Allowed at</span>
          <v-chip v-if="isLicenseGlobal">
            <v-icon
              left
              class="mr-2"
              icon="mdi-earth"
            />
            Global
          </v-chip>
          <v-chip v-else>
            <v-icon
              left
              icon="mdi-flag"
            />
            Limited ({{ license.allowed_regions.join(", ") }})
          </v-chip>
        </v-row>
      </v-card-item>
      <v-divider class="my-4" />
      <v-card-item class="pa-0">
        <v-card-title class="text-h6 text-center">
          License Owner
        </v-card-title>
        <v-row
          v-for="(value, name) in license.customer"
          :key="name"
          :data-test="name"
        >
          <span>{{ formatName(name) }}</span>
          <CopyWarning
            v-if="name === 'id'"
            copied-item="Tenant ID"
          >
            <template #default="{ copyText }">
              <v-chip>
                <v-tooltip anchor="top">
                  <template #activator="{ props }">
                    <span
                      v-bind="props"
                      class="hover-text"
                      @click="copyText(value as string)"
                      @keypress.enter="copyText(value as string)"
                    >
                      {{ value }}
                    </span>
                  </template>
                  <span>Copy ID</span>
                </v-tooltip>
              </v-chip>
            </template>
          </CopyWarning>
          <span v-else>{{ value }}</span>
        </v-row>
      </v-card-item>
      <v-divider class="my-4" />
      <v-card-item class="pa-0">
        <v-card-title class="text-h6 text-center">
          Features
        </v-card-title>
        <v-row
          v-for="(value, name) in getFeatures(license.features)"
          :key="name"
          :data-test="name"
        >
          <span>{{ formatName(`${name}`) }}</span>
          <div v-if="typeof value === 'boolean'">
            <v-icon
              v-if="value"
              color="success"
              data-test="included-icon"
              icon="mdi-check-circle"
            />
            <v-icon
              v-else
              color="#E53935"
              data-test="not-included-icon"
              icon="mdi-close-circle"
            />
          </div>
          <v-chip v-else>
            {{ formatFeatureValue(value) }}
          </v-chip>
        </v-row>
      </v-card-item>
      <v-divider class="my-4" />
    </div>
    <v-card-item>
      <v-card-title class="text-h6 text-center">
        License Field
      </v-card-title>
      <v-file-input
        v-model="currentFile"
        class="mt-4 mb-2"
        accept=".dat"
        show-size
        variant="outlined"
        label="Select license file"
        counter
        :rules="[validateLicenseFile]"
        @update:model-value="disableUploadButton = !currentFile"
      />
      <v-btn
        variant="outlined"
        :disabled="disableUploadButton"
        text="Upload"
        @click="uploadLicense"
      />
    </v-card-item>
  </v-card>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import moment from "moment";
import useLicenseStore from "@admin/store/modules/license";
import { AdminLicenseFeatures } from "@admin/interfaces/ILicense";
import CopyWarning from "@/components/User/CopyWarning.vue";
import useSnackbar from "@/helpers/snackbar";
import handleError from "@/utils/handleError";

const licenseStore = useLicenseStore();
const snackbar = useSnackbar();
const license = computed(() => licenseStore.license);
const isLicenseInstalled = computed(() => license.value && license.value.grace_period !== undefined);
const isLicenseGlobal = computed(() => license.value.allowed_regions.length === 0);
const currentFile = ref<File | null>(null);
const disableUploadButton = ref(true);
const licenseAlert = computed(() => {
  if (!isLicenseInstalled.value) {
    return {
      type: "info" as const,
      message: "You do not have an installed license",
    };
  }

  if (license.value.about_to_expire) {
    return {
      type: "info" as const,
      message: "Your license is about to expire!",
    };
  }

  if (license.value.expired && license.value.grace_period) {
    return {
      type: "warning" as const,
      message: "Your license has expired, but you are still within the grace period.",
    };
  }

  if (license.value.expired && !license.value.grace_period) {
    return {
      type: "error" as const,
      message: "Your license has expired!",
    };
  }

  return null;
});

const convertValueToDate = (value: string | number) => value === -1 ? "Now" : moment.unix(+value).format("LL");

const formatName = (name: string) => name.charAt(0).toUpperCase() + name.slice(1).replace("_", " ");

const getFeatures = (jsonLicense: AdminLicenseFeatures) => {
  const updatedJsonLicense = jsonLicense;
  delete updatedJsonLicense?.login_link;
  delete updatedJsonLicense?.reports;
  return updatedJsonLicense;
};

const formatFeatureValue = (value?: number) => value === -1 ? "Unlimited" : value;

const validateLicenseFile = (file: File | null): string | boolean => {
  if (!file) return true;
  if (!file.name.endsWith(".dat")) return "Only .dat files are allowed!";
  if (file.size >= 32 * 1024) return "License size must be less than 32 KB!";
  return true;
};

const uploadLicense = async () => {
  if (!currentFile.value) return;
  try {
    await licenseStore.uploadLicense(currentFile.value);
    await licenseStore.getLicense();
    snackbar.showSuccess("License uploaded successfully.");
    currentFile.value = null;
  } catch (error) {
    handleError(error);
    snackbar.showError("Failed to upload the license.");
  }
};

onMounted(async () => {
  try {
    await licenseStore.getLicense();
  } catch { snackbar.showError("Error loading license."); }
});

defineExpose({ license });
</script>

<style scoped>
.v-row {
  margin: 1.25rem 0;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.hover-text {
  cursor: pointer;
  animation: fadeIn 0.5s;
}

.hover-text:hover {
  text-decoration: underline;
}
</style>
