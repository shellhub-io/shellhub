<template>
  <fragment>
    <v-navigation-drawer
      app
      dark
    >
      <v-list-item>
        <v-list-item-content>
          <v-list-item-title>
            <router-link to="/">
              <v-img
                class="d-sm-flex hidden-sm-and-down"
                src="@/assets/logo-inverted.png"
                max-width="140"
              />
            </router-link>
          </v-list-item-title>
        </v-list-item-content>
      </v-list-item>

      <v-divider class="ma-2" />

      <div class="pr-2 pl-2">
        <Namespace data-test="namespace-component" />
      </div>

      <v-divider class="ma-2" />

      <v-list>
        <v-list-item
          v-for="item in visibleItems"
          :key="item.title"
          :to="item.path"
          two-line
          :disabled="disableItem(item.icon)"
        >
          <v-list-item-action>
            <v-icon v-text="item.icon" />
          </v-list-item-action>

          <v-list-item-content>
            <v-list-item-title
              :data-test="item.icon+'-listItem'"
              v-text="item.title"
            />
          </v-list-item-content>
        </v-list-item>
      </v-list>
    </v-navigation-drawer>

    <AppBar />

    <v-main>
      <v-container
        class="pl-8 pr-8"
        fluid
      >
        <router-view :key="$route.fullPath" />
      </v-container>
    </v-main>

    <v-overlay :value="hasSpinner">
      <v-progress-circular
        indeterminate
        size="64"
      />
    </v-overlay>

    <UserWarning data-test="userWarning-component" />
  </fragment>
</template>

<script>

import AppBar from '@/components/app_bar/AppBar';
import UserWarning from '@/components/user/UserWarning';
import Namespace from '@/components/app_bar/namespace/Namespace';

export default {
  name: 'AppLayoutComponent',

  components: {
    AppBar,
    UserWarning,
    Namespace,
  },

  data() {
    return {
      items: [
        {
          icon: 'dashboard',
          title: 'Dashboard',
          path: '/',
        },
        {
          icon: 'devices',
          title: 'Devices',
          path: '/devices',
        },
        {
          icon: 'history',
          title: 'Sessions',
          path: '/sessions',
        },
        {
          icon: 'security',
          title: 'Firewall Rules',
          path: '/firewall/rules',
          hidden: !this.$env.isEnterprise,
        },
        {
          icon: 'vpn_key',
          title: 'Public Keys',
          path: '/sshkeys/public-keys',
        },
        {
          icon: 'mdi-cog',
          title: 'Settings',
          path: '/settings/namespace-manager',
        },
      ],
      admins: [
        ['Management', 'people_outline'],
        ['Settings', 'settings'],
      ],
    };
  },

  computed: {
    visibleItems() {
      return this.items.filter((item) => !item.hidden);
    },

    hasNamespaces() {
      return this.$store.getters['namespaces/getNumberNamespaces'] !== 0;
    },

    hasSpinner() {
      return this.$store.getters['spinner/getStatus'];
    },
  },

  beforeDestroy() {
    if (typeof window === 'undefined') return;

    window.removeEventListener('resize', this.onResize, { passive: true });
  },

  created() {
    this.onResize();
    window.addEventListener('resize', this.onResize, { passive: true });

    this.$store.dispatch('privatekeys/fetch');
  },

  methods: {
    triggerClick(item) {
      switch (item.type) {
      case 'path':
        this.$router.push(item.path).catch(() => {});
        break;
      case 'method':
        this[item.method]();
        break;
      default:
        break;
      }
    },

    onResize() {
      const isMobile = this.$vuetify.breakpoint.mobile;
      this.$store.dispatch('mobile/setIsMobileStatus', isMobile);
    },

    disableItem(item) {
      return !this.hasNamespaces && item !== 'dashboard';
    },
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
