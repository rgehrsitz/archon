<script lang="ts" setup>
// import { useTheme } from 'vuetify'

import VerticalNavSectionTitle from '@/@layouts/components/VerticalNavSectionTitle.vue'
import VerticalNavLayout from '@layouts/components/VerticalNavLayout.vue'
import VerticalNavLink from '@layouts/components/VerticalNavLink.vue'

// Components
import Footer from '@/layouts/components/Footer.vue'
import NavbarThemeSwitcher from '@/layouts/components/NavbarThemeSwitcher.vue'
import UserProfile from '@/layouts/components/UserProfile.vue'

// Banner

// const vuetifyTheme = useTheme()


</script>

<template>
  <VerticalNavLayout>
    <!-- 👉 navbar -->
    <template #navbar="{ toggleVerticalOverlayNavActive }">
      <div class="d-flex h-100 align-center">
        <!-- 👉 Vertical nav toggle in overlay mode -->
        <IconBtn class="ms-n3 d-lg-none" @click="toggleVerticalOverlayNavActive(true)">
          <VIcon icon="bx-menu" />
        </IconBtn>

        <!-- 👉 Search -->
        <div class="d-flex align-center cursor-pointer" style="user-select: none;">
          <!-- 👉 Search Trigger button -->
          <IconBtn>
            <VIcon icon="bx-search" />
          </IconBtn>

          <span class="d-none d-md-flex align-center text-disabled">
            <span class="me-3">Search</span>
            <span class="meta-key">&#8984;K</span>
          </span>
        </div>

        <VSpacer />

        <IconBtn class="me-2" href="https://github.com/themeselection/sneat-vuetify-vuejs-admin-template-free"
          target="_blank" rel="noopener noreferrer">
          <VIcon icon="bxl-github" />
        </IconBtn>

        <IconBtn class="me-2">
          <VIcon icon="bx-bell" />
        </IconBtn>

        <NavbarThemeSwitcher class="me-2" />

        <UserProfile />
      </div>
    </template>

    <template #vertical-nav-content>
      <VerticalNavLink :item="{
        title: 'Dashboard',
        icon: 'bx-home',
        to: '/dashboard',
      }" />
      <VerticalNavLink :item="{
        title: 'Account Settings',
        icon: 'mdi-account-cog-outline',
        to: '/account-settings',
      }" />

      <!-- 👉 Pages -->
      <VerticalNavSectionTitle :item="{
        heading: 'Project',
      }" />
      <VerticalNavLink :item="{
        title: 'Equipment List',
        icon: 'bx-list-ul',
        to: '/equipment-list',
      }" />
      <VerticalNavLink :item="{
        title: 'Diff View',
        icon: 'bx-git-compare',
        to: '/diff-view',
      }" />
      <VerticalNavLink :item="{
        title: 'Git Functions',
        icon: 'bx-git-branch',
        to: '/git-functions',
      }" />

      <!-- 👉 User Interface -->
      <VerticalNavSectionTitle :item="{
        heading: 'Other',
      }" />
      <VerticalNavLink :item="{
        title: 'Settings',
        icon: 'bx-cog',
        to: '/settings',
      }" />
    </template>

    <template #after-vertical-nav-items>
      <!-- 👉 illustration -->
      <!--       <a href="https://themeselection.com/item/sneat-vuetify-vuejs-admin-template" target="_blank"
        rel="noopener noreferrer" style="margin-left: 7px;">
        <img :src="upgradeBanner" alt="upgrade-banner" transition="scale-transition" class="upgrade-banner mx-auto"
          style="max-width: 230px;">
      </a> -->
    </template>

    <!-- 👉 Pages -->
    <slot />

    <!-- 👉 Footer -->
    <template #footer>
      <Footer />
    </template>
  </VerticalNavLayout>
</template>

<style lang="scss" scoped>
.meta-key {
  border: thin solid rgba(var(--v-border-color), var(--v-border-opacity));
  border-radius: 6px;
  block-size: 1.5625rem;
  line-height: 1.3125rem;
  padding-block: 0.125rem;
  padding-inline: 0.25rem;
}
</style>
