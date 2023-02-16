<template>
  <div v-if="!isOwner" class="text-center">
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

<script lang="ts">
import { defineComponent, computed } from "vue";
import { useStore } from "../../store";
import { INamespaceMember } from "@/interfaces/INamespace";

export default defineComponent({
  props: {
    isOwner: {
      type: Boolean,
      required: true,
    },
  },
  setup() {
    const store = useStore();

    const namespace = computed(() => store.getters["namespaces/get"]);

    const owner = computed(() => store.getters["namespaces/get"].owner);

    const namespaceOwnerName = computed(() => {
      if (namespace.value.members !== undefined) {
        const ownerName = namespace.value.members.find(
          (member: INamespaceMember) => member.id === owner.value,
        );
        return ownerName?.username;
      }
      return null;
    });

    return {
      namespaceOwnerName,
    };
  },
});
</script>
