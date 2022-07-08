<template>
  <div v-if="!isOwner" class="text-center">
    <div data-test="message-div">
      <h3 class="pl-6">
        <span> You're not the owner of this namespace. </span>

        <p data-test="contactUser-p">
          Contact {{ namespaceOwnerName() }} user for more information.
        </p>
      </h3>
    </div>
    <br />
  </div>
</template>

<script lang="ts">
import { useStore } from "../../store";
import { defineComponent, computed } from "vue";

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

    const owner = computed(
      () =>
        store.getters["namespaces/get"].owner &&
        store.getters["namespaces/get"].owner
    );

    const namespaceOwnerName = () => {
      if (namespace.value.members !== undefined) {
        return namespace.value.members.find((x: any) => x.id === owner.value).name;
      }
      return null;
    };

    return {
      namespaceOwnerName,
    };
  },
});
</script>
