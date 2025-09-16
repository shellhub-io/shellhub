import { defineStore } from "pinia";
import { ref, computed } from "vue";
import { IAdminLicense } from "@admin/interfaces/ILicense";
import * as apiLicense from "../api/license";

const useLicenseStore = defineStore("license", () => {
  const license = ref({} as IAdminLicense);
  const isExpired = computed(() => (license.value && license.value.expired)
    || (license.value && license.value.expired === undefined));

  const getLicense = async () => {
    const res = await apiLicense.getLicense();
    license.value = res.data as IAdminLicense;
  };

  const uploadLicense = async (file: File) => {
    await apiLicense.uploadLicense(file);
  };

  return {
    license,
    isExpired,
    getLicense,
    uploadLicense,
  };
});

export default useLicenseStore;
