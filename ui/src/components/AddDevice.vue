<template>
<v-dialog v-model="show" max-width="1024px">
    <template v-slot:activator="{ on }">
        <v-btn text @click="show = true">Add Device</v-btn>
    </template>
    <v-card>
        <v-card-title>Run the following command to register your device:</v-card-title>
        <v-card-text>
<code class="pa-2">
docker run -d --restart=unless-stopped --privileged --net=host --pid=host -v /:/host -v /etc/os-release:/etc/os-release -e SERVER_ADDRESS=http://{{ hostname }} -e PRIVATE_KEY=/host/etc/shellhub.key -e TENANT_ID={{ tenant }} shellhub/agent:latest
</code>
        </v-card-text>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn text @click="show = false">Close</v-btn>
        </v-card-actions>
    </v-card>
</v-dialog>
</template>

<script>
export default {
  name: "AddDevice",

  data() {
    return {
      show: false,
      hostname: window.location.hostname
    };
  },

  computed: {
    tenant() {
      return this.$store.getters["auth/tenant"];
    }
  }
};
</script>
