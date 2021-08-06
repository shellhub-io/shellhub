<template>
  <div
    v-if="!isOwner"
    style="text-align:center"
  >
    <div data-test="message-div">
      <h3 class="pl-6">
        <span>
          You're not the owner of this namespace.
        </span>

        <p data-test="contactUser-p">
          Contact {{ namespaceOwnerName() }} user for more information.
        </p>
      </h3>
    </div>
    <br>
  </div>
</template>

<script>

export default {
  props: {
    isOwner: {
      type: Boolean,
      required: true,
    },
  },

  computed: {
    namespace() {
      return this.$store.getters['namespaces/get'];
    },

    owner() {
      return this.$store.getters['namespaces/get'].owner;
    },
  },

  methods: {
    namespaceOwnerName() {
      if (this.namespace.members !== undefined) {
        return this.namespace.members.find((x) => x.id === this.owner).name;
      }
      return null;
    },
  },
};

</script>
