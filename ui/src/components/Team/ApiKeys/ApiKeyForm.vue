<template>
  <v-card-text data-test="api-key-form-content">
    <v-text-field
      class="mt-6"
      v-model="keyName"
      :error-messages="keyNameError"
      label="Key Name"
      persistent-placeholder
      required
      data-test="key-name-text"
      :hint="mode === 'edit'
        ? 'Please note that the new name must be unique and not already in use by another key.'
        : 'Provide a descriptive name for this key'"
      persistent-hint
    />

    <v-row v-if="mode === 'create'" class="mt-6">
      <v-col>
        <v-select
          v-model="selectedDate"
          label="Expiration date"
          :items="itemsDate"
          :item-props="true"
          :hint="expirationHint"
          return-object
          data-test="api-key-expiration-date"
        />
      </v-col>
      <v-col>
        <RoleSelect
          v-if="canManageRoles"
          v-model="selectedRole"
          data-test="api-key-role"
        />
      </v-col>
    </v-row>

    <div v-else-if="mode === 'edit'" class="mt-6">
      <RoleSelect
        v-if="canManageRoles"
        v-model="selectedRole"
        data-test="api-key-role"
      />
    </div>
  </v-card-text>
</template>

<script setup lang="ts">
import { ref, watch, computed } from "vue";
import moment from "moment";
import * as yup from "yup";
import { useField, useForm } from "vee-validate";
import RoleSelect from "@/components/Team/RoleSelect.vue";
import { BasicRole } from "@/interfaces/INamespace";

interface Props {
  mode: "create" | "edit";
  initialKeyName?: string;
  initialRole?: BasicRole;
  canManageRoles?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  initialKeyName: "",
  initialRole: "administrator",
  canManageRoles: true,
});

const emit = defineEmits<{
  submit: [data: {
    name: string;
    expires_in?: number;
    role: BasicRole;
  }];
  "update:valid": [valid: boolean];
}>();

const { handleSubmit, resetForm, meta } = useForm();

const {
  value: keyName,
  errorMessage: keyNameError,
} = useField<string>(
  "keyName",
  yup
    .string()
    .required("Key name is required")
    .min(3, "Key name must be at least 3 characters")
    .max(20, "Key name must be at most 20 characters")
    .matches(/^(?!.*\s).*$/, "This field cannot contain any blank spaces"),
  {
    initialValue: props.initialKeyName,
  },
);

const getExpiryDate = (item: string) => {
  if (item === "No expire") {
    return "Never Expires";
  }

  const [value, unit] = item.split(" ");
  return `Expires in ${
    moment()
      .add(Number(value), unit as moment.unitOfTime.DurationConstructor)
      .format("MMMM, YYYY")
  }`;
};

const itemsDate = [
  {
    title: "30 days",
    subtitle: getExpiryDate("30 days"),
    time: 30,
  },
  {
    title: "60 days",
    subtitle: getExpiryDate("60 days"),
    time: 60,
  },
  {
    title: "90 days",
    subtitle: getExpiryDate("90 days"),
    time: 90,
  },
  {
    title: "1 year",
    subtitle: getExpiryDate("1 year"),
    time: 365,
  },
  {
    title: "No expire",
    subtitle: getExpiryDate("No expire"),
    time: -1,
  },
];

const selectedDate = ref(itemsDate[0]);
const selectedRole = ref<BasicRole>(props.initialRole);
const expirationHint = ref(getExpiryDate(selectedDate.value.title));

watch(selectedDate, (newVal) => {
  expirationHint.value = getExpiryDate(newVal.title);
});

// Emit validity changes
watch(() => meta.value.valid, (isValid) => {
  emit("update:valid", isValid);
}, { immediate: true });

const submitForm = handleSubmit(() => {
  const data = {
    name: keyName.value,
    role: selectedRole.value,
    ...(props.mode === "create" && { expires_in: selectedDate.value.time }),
  };
  emit("submit", data);
});

const reset = () => {
  resetForm();
  keyName.value = props.initialKeyName;
  selectedRole.value = props.initialRole;
  [selectedDate.value] = itemsDate;
};

defineExpose({
  submitForm,
  reset,
  isValid: computed(() => meta.value.valid),
});
</script>
