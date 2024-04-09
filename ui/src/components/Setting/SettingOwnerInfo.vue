<template>
  <div v-if="!props.isOwner" class="text-center">
    <div data-test="message-div">
      <h3 class="pl-6">
        <span> You're not the owner of this namespace. </span>

        <p data-test="contactUser-p">
          Contact {{ namespaceOwnerName }} user for more information.
        </p>
      </h3>
    </div>
    <br />
  </div>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useStore } from "../../store";
import { INamespaceMember } from "@/interfaces/INamespace";

const props = defineProps({
  isOwner: {
    type: Boolean,
    required: true,
  },
});
const store = useStore();

const namespace = computed(() => store.getters["namespaces/get"]);

const owner = computed(() => store.getters["namespaces/get"].owner);

const namespaceOwnerName = computed((): string | null => {
  const ownerName = namespace.value?.members?.find(
    (member: INamespaceMember) => member.id === owner.value,
  );

  return ownerName?.username ?? null;
});

</script>
