<template>
  <v-container>
    <v-card
      :elevation="0"
    >
      <v-app-bar
        flat
        color="transparent"
      >
        <v-tabs
          centered
        >
          <v-tab
            v-for="item in items"
            :key="item.title"
            :to="item.path"
          >
            {{ item.title }}
          </v-tab>
        </v-tabs>
      </v-app-bar>

      <v-divider />

      <v-container
        class="pa-0"
        fluid
      >
        <router-view />
      </v-container>
    </v-card>
  </v-container>
</template>

<script>

export default {
  name: 'Settings',

  data() {
    return {
      drawer: true,
      clipped: false,
      items: [
        {
          title: 'Profile',
          path: '/settings',
        },
      ],
    };
  },
};
</script>

<style>
.v-list-active {
  border-left: 4px solid var(--v-primary-base);
}

.text-shadow {
  text-shadow: #000 0 0 6px;
  color: transparent;
}
</style>
